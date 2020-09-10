// Copyright 2020 Chaos Mesh Authors.
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

package trafficcontrol

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/podnetworkmanager"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/ipset"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/netutils"
	"github.com/chaos-mesh/chaos-mesh/controllers/twophase"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

const (
	networkTcActionMsg = "network traffic control action duration %s"
)

func newReconciler(c client.Client, r client.Reader, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) twophase.Reconciler {
	return twophase.Reconciler{
		InnerReconciler: &Reconciler{
			Client:        c,
			Reader:        r,
			EventRecorder: recorder,
			Log:           log,
		},
		Client: c,
		Log:    log,
	}
}

// NewTwoPhaseReconciler would create Reconciler for twophase package
func NewTwoPhaseReconciler(c client.Client, reader client.Reader, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) *twophase.Reconciler {
	r := newReconciler(c, reader, log, req, recorder)
	return twophase.NewReconciler(r, r.Client, r.Reader, r.Log)
}

// NewCommonReconciler would create Reconciler for common package
func NewCommonReconciler(c client.Client, reader client.Reader, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) *common.Reconciler {
	r := newReconciler(c, reader, log, req, recorder)
	return common.NewReconciler(r, r.Client, r.Reader, r.Log)
}

type Reconciler struct {
	client.Client
	client.Reader
	record.EventRecorder
	Log logr.Logger
}

// Object implements the reconciler.InnerReconciler.Object
func (r *Reconciler) Object() v1alpha1.InnerObject {
	return &v1alpha1.NetworkChaos{}
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	r.Log.Info("traffic control Apply", "req", req, "chaos", chaos)

	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)
		return err
	}

	source := networkchaos.Namespace + "/" + networkchaos.Name
	m := podnetworkmanager.New(source, r.Log, r.Client, r.Reader)

	sources, err := utils.SelectAndFilterPods(ctx, r.Client, r.Reader, &networkchaos.Spec)
	if err != nil {
		r.Log.Error(err, "failed to select and filter pods")
		return err
	}

	var targets []v1.Pod

	// We should only apply filter when we specify targets
	if networkchaos.Spec.Target != nil {
		targets, err = utils.SelectAndFilterPods(ctx, r.Client, r.Reader, networkchaos.Spec.Target)
		if err != nil {
			r.Log.Error(err, "failed to select and filter pods")
			return err
		}
	}

	pods := append(sources, targets...)

	externalCidrs, err := netutils.ResolveCidrs(networkchaos.Spec.ExternalTargets)
	if err != nil {
		r.Log.Error(err, "failed to resolve external targets")
		return err
	}

	switch networkchaos.Spec.Direction {
	case v1alpha1.To:
		err = r.applyTc(ctx, sources, targets, externalCidrs, m, networkchaos)
		if err != nil {
			r.Log.Error(err, "failed to apply traffic control", "sources", sources, "targets", targets)
			return err
		}
	case v1alpha1.From:
		err = r.applyTc(ctx, targets, sources, []string{}, m, networkchaos)
		if err != nil {
			r.Log.Error(err, "failed to apply traffic control", "sources", targets, "targets", sources)
			return err
		}
	case v1alpha1.Both:
		err = r.applyTc(ctx, pods, pods, externalCidrs, m, networkchaos)
		if err != nil {
			r.Log.Error(err, "failed to apply traffic control", "sources", pods, "targets", pods)
			return err
		}
	default:
		err = fmt.Errorf("unknown direction %s", networkchaos.Spec.Direction)
		r.Log.Error(err, "unknown direction", "direction", networkchaos.Spec.Direction)
		return err
	}

	err = m.Commit(ctx)
	if err != nil {
		r.Log.Error(err, "fail to commit")
		return err
	}

	networkchaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(pods))
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(networkchaos.Spec.Action),
		}

		if networkchaos.Spec.Duration != nil {
			ps.Message = fmt.Sprintf(networkTcActionMsg, *networkchaos.Spec.Duration)
		}

		networkchaos.Status.Experiment.PodRecords = append(networkchaos.Status.Experiment.PodRecords, ps)
	}
	r.Event(networkchaos, v1.EventTypeNormal, utils.EventChaosInjected, "")
	return nil
}

// Recover implements the reconciler.InnerReconciler.Recover
func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)
		return err
	}

	if err := r.cleanFinalizersAndRecover(ctx, networkchaos); err != nil {
		return err
	}
	r.Event(networkchaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")
	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, networkchaos *v1alpha1.NetworkChaos) error {
	var result error

	source := networkchaos.Namespace + "/" + networkchaos.Name
	m := podnetworkmanager.New(source, r.Log, r.Client, r.Reader)

	for _, key := range networkchaos.Finalizers {
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

		networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, key)
	}
	r.Log.Info("After recovering", "finalizers", networkchaos.Finalizers)

	if networkchaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", networkchaos)
		networkchaos.Finalizers = make([]string, 0)
		return nil
	}

	return result
}

func (r *Reconciler) applyTc(ctx context.Context, sources, targets []v1.Pod, externalTargets []string, m *podnetworkmanager.PodNetworkManager, networkchaos *v1alpha1.NetworkChaos) error {
	for index := range sources {
		pod := &sources[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}

		networkchaos.Finalizers = utils.InsertFinalizer(networkchaos.Finalizers, key)
	}

	tcType := v1alpha1.Bandwidth
	switch networkchaos.Spec.Action {
	case v1alpha1.NetemAction, v1alpha1.DelayAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction, v1alpha1.LossAction:
		tcType = v1alpha1.Netem
	case v1alpha1.BandwidthAction:
		tcType = v1alpha1.Bandwidth
	default:
		return fmt.Errorf("unknown action %s", networkchaos.Spec.Action)
	}

	// if we don't specify targets, then sources pods apply traffic control on all egress traffic
	if len(targets)+len(externalTargets) == 0 {
		r.Log.Info("apply traffic control", "sources", sources)
		for index := range sources {
			pod := &sources[index]

			t := m.WithInit(types.NamespacedName{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			})
			t.Append(v1alpha1.RawTrafficControl{
				Type:        tcType,
				TcParameter: networkchaos.Spec.TcParameter,
				Source:      m.Source,
			})
		}
		return nil
	}

	// create ipset contains all target ips
	dstIpset := ipset.BuildIPSet(targets, externalTargets, networkchaos, string(tcType)[0:5], m.Source)
	r.Log.Info("apply traffic control with filter", "sources", sources, "ipset", dstIpset)

	for index := range sources {
		pod := &sources[index]

		t := m.WithInit(types.NamespacedName{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		})
		t.Append(dstIpset)
		t.Append(v1alpha1.RawTrafficControl{
			Type:        tcType,
			TcParameter: networkchaos.Spec.TcParameter,
			Source:      m.Source,
			IPSet:       dstIpset.Name,
		})
	}

	return nil
}
