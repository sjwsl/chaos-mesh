// Copyright 2019 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package iochaos

import (
	"context"
	"errors"

	"github.com/hashicorp/go-multierror"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/iochaos/podiochaosmanager"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

type endpoint struct {
	ctx.Context
}

func (r *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.IoChaos{}
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	iochaos, ok := chaos.(*v1alpha1.IoChaos)
	if !ok {
		err := errors.New("chaos is not IOChaos")
		r.Log.Error(err, "chaos is not IOChaos", "chaos", chaos)
		return err
	}

	source := iochaos.Namespace + "/" + iochaos.Name
	m := podiochaosmanager.New(source, r.Log, r.Client)

	pods, err := utils.SelectAndFilterPods(ctx, r.Client, r.Reader, &iochaos.Spec)
	if err != nil {
		r.Log.Error(err, "failed to select and filter pods")
		return err
	}

	r.Log.Info("applying iochaos", "iochaos", iochaos)

	for _, pod := range pods {
		t := m.WithInit(types.NamespacedName{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		})

		// TODO: support chaos on multiple volume
		t.SetVolumePath(iochaos.Spec.VolumePath)
		t.Append(v1alpha1.IoChaosAction{
			Type: iochaos.Spec.Action,
			Filter: v1alpha1.Filter{
				Path:    iochaos.Spec.Path,
				Percent: iochaos.Spec.Percent,
				Methods: iochaos.Spec.Methods,
			},
			Faults: []v1alpha1.IoFault{
				{
					Errno:  iochaos.Spec.Errno,
					Weight: 1,
				},
			},
			Latency:          iochaos.Spec.Delay,
			AttrOverrideSpec: iochaos.Spec.Attr,
			Source:           m.Source,
		})

		key, err := cache.MetaNamespaceKeyFunc(&pod)
		if err != nil {
			return err
		}
		iochaos.Finalizers = utils.InsertFinalizer(iochaos.Finalizers, key)
	}
	r.Log.Info("commiting updates of podiochaos")
	err = m.Commit(ctx)
	if err != nil {
		r.Log.Error(err, "fail to commit")
		return err
	}

	iochaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(pods))
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(iochaos.Spec.Action),
		}

		iochaos.Status.Experiment.PodRecords = append(iochaos.Status.Experiment.PodRecords, ps)
	}
	r.Event(iochaos, v1.EventTypeNormal, utils.EventChaosInjected, "")

	return nil
}

// Recover implements the reconciler.InnerReconciler.Recover
func (r *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	iochaos, ok := chaos.(*v1alpha1.IoChaos)
	if !ok {
		err := errors.New("chaos is not IOChaos")
		r.Log.Error(err, "chaos is not IOChaos", "chaos", chaos)
		return err
	}

	if err := r.cleanFinalizersAndRecover(ctx, iochaos); err != nil {
		return err
	}
	r.Event(iochaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")
	return nil
}

func (r *endpoint) cleanFinalizersAndRecover(ctx context.Context, chaos *v1alpha1.IoChaos) error {
	var result error

	source := chaos.Namespace + "/" + chaos.Name

	for _, key := range chaos.Finalizers {
		m := podiochaosmanager.New(source, r.Log, r.Client)

		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		_ = m.WithInit(types.NamespacedName{
			Namespace: ns,
			Name:      name,
		})

		err = m.Commit(ctx)
		if err != nil {
			r.Log.Error(err, "fail to commit")
		}

		chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, key)
	}
	r.Log.Info("After recovering", "finalizers", chaos.Finalizers)

	if chaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", chaos)
		chaos.Finalizers = make([]string, 0)
		return nil
	}

	return result
}

func init() {
	router.Register("iochaos", &v1alpha1.IoChaos{}, func(obj runtime.Object) bool {
		return true
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
