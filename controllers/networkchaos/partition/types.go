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

package partition

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/ipset"
	"github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/iptable"
	"github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/netutils"
	"github.com/chaos-mesh/chaos-mesh/controllers/twophase"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

const (
	networkPartitionActionMsg = "partition network duration %s"

	sourceIPSetPostFix = "src"
	targetIPSetPostFix = "tgt"
)

func newReconciler(c client.Client, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) twophase.Reconciler {
	return twophase.Reconciler{
		InnerReconciler: &Reconciler{
			Client:        c,
			EventRecorder: recorder,
			Log:           log,
		},
		Client: c,
		Log:    log,
	}
}

// NewTwoPhaseReconciler would create Reconciler for twophase package
func NewTwoPhaseReconciler(c client.Client, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) *twophase.Reconciler {
	r := newReconciler(c, log, req, recorder)
	return twophase.NewReconciler(r, r.Client, r.Log)
}

// NewCommonReconciler would create Reconciler for common package
func NewCommonReconciler(c client.Client, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) *common.Reconciler {
	r := newReconciler(c, log, req, recorder)
	return common.NewReconciler(r, r.Client, r.Log)
}

type Reconciler struct {
	client.Client
	record.EventRecorder
	Log logr.Logger
}

// Object implements the reconciler.InnerReconciler.Object
func (r *Reconciler) Object() v1alpha1.InnerObject {
	return &v1alpha1.NetworkChaos{}
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	r.Log.Info("Applying network partition")

	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)

		return err
	}

	sources, err := utils.SelectAndFilterPods(ctx, r.Client, &networkchaos.Spec)

	if err != nil {
		r.Log.Error(err, "failed to select and filter pods")
		return err
	}

	var targets []v1.Pod

	if networkchaos.Spec.Target != nil {
		targets, err = utils.SelectAndFilterPods(ctx, r.Client, networkchaos.Spec.Target)
		if err != nil {
			r.Log.Error(err, "failed to select and filter pods")
			return err
		}
	}

	sourceSet := ipset.BuildIPSet(sources, []string{}, networkchaos, sourceIPSetPostFix)
	externalCidrs, err := netutils.ResolveCidrs(networkchaos.Spec.ExternalTargets)
	if err != nil {
		r.Log.Error(err, "failed to resolve external targets")
		return err
	}
	targetSet := ipset.BuildIPSet(targets, externalCidrs, networkchaos, targetIPSetPostFix)

	allPods := append(sources, targets...)

	// Set up ipset in every related pods
	g := errgroup.Group{}
	for index := range allPods {
		pod := allPods[index]
		r.Log.Info("PODS", "name", pod.Name, "namespace", pod.Namespace)
		g.Go(func() error {
			err = ipset.FlushIPSets(ctx, r.Client, &pod, []*pb.IPSet{&sourceSet})
			if err != nil {
				return err
			}

			r.Log.Info("Flush ipset on pod", "name", pod.Name, "namespace", pod.Namespace)
			return ipset.FlushIPSets(ctx, r.Client, &pod, []*pb.IPSet{&targetSet})
		})
	}

	if err = g.Wait(); err != nil {
		r.Log.Error(err, "flush pod ipset error")
		return err
	}

	sourcesChains := []*pb.Chain{}
	targetsChains := []*pb.Chain{}
	if networkchaos.Spec.Direction == v1alpha1.To || networkchaos.Spec.Direction == v1alpha1.Both {
		sourcesChains = append(sourcesChains, &pb.Chain{
			Name:      iptable.GenerateName(pb.Chain_OUTPUT, networkchaos),
			Direction: pb.Chain_OUTPUT,
			Ipsets:    []string{targetSet.Name},
		})

		targetsChains = append(targetsChains, &pb.Chain{
			Name:      iptable.GenerateName(pb.Chain_INPUT, networkchaos),
			Direction: pb.Chain_INPUT,
			Ipsets:    []string{sourceSet.Name},
		})
	}

	if networkchaos.Spec.Direction == v1alpha1.From || networkchaos.Spec.Direction == v1alpha1.Both {
		sourcesChains = append(sourcesChains, &pb.Chain{
			Name:      iptable.GenerateName(pb.Chain_INPUT, networkchaos),
			Direction: pb.Chain_INPUT,
			Ipsets:    []string{targetSet.Name},
		})

		targetsChains = append(targetsChains, &pb.Chain{
			Name:      iptable.GenerateName(pb.Chain_OUTPUT, networkchaos),
			Direction: pb.Chain_OUTPUT,
			Ipsets:    []string{sourceSet.Name},
		})
	}

	err = r.SetChains(ctx, sources, sourcesChains, networkchaos)
	if err != nil {
		return err
	}

	err = r.SetChains(ctx, targets, targetsChains, networkchaos)
	if err != nil {
		return err
	}

	networkchaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(allPods))
	for _, pod := range allPods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(networkchaos.Spec.Action),
		}

		if networkchaos.Spec.Duration != nil {
			ps.Message = fmt.Sprintf(networkPartitionActionMsg, *networkchaos.Spec.Duration)
		}

		networkchaos.Status.Experiment.PodRecords = append(networkchaos.Status.Experiment.PodRecords, ps)
	}

	r.Event(networkchaos, v1.EventTypeNormal, utils.EventChaosInjected, "")
	return nil
}

// SetChains sets iptables chains for pods
func (r *Reconciler) SetChains(ctx context.Context, pods []v1.Pod, chains []*pb.Chain, networkchaos *v1alpha1.NetworkChaos) error {
	r.Log.Info("setting chains", "chains", chains, "pods", pods)

	g := errgroup.Group{}

	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}

		networkchaos.Finalizers = utils.InsertFinalizer(networkchaos.Finalizers, key)

		g.Go(func() error {
			return iptable.SetIptablesChains(ctx, r.Client, pod, chains)
		})
	}
	return g.Wait()
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
		r.Log.Error(err, "cleanFinalizersAndRecover failed")
		return err
	}
	r.Event(networkchaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")

	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, networkchaos *v1alpha1.NetworkChaos) error {
	var result error

	for _, key := range networkchaos.Finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		var pod v1.Pod
		err = r.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, &pod)

		if err != nil {
			if !k8sError.IsNotFound(err) {
				result = multierror.Append(result, err)
				continue
			}

			r.Log.Info("Pod not found", "namespace", ns, "name", name)
			networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, key)
			continue
		}

		chains := []*pb.Chain{}
		for _, direction := range []string{"INPUT", "OUTPUT"} {
			var chainName string
			var chainDirection pb.Chain_Direction

			switch direction {
			case "INPUT":
				chainName = "INPUT/" + netutils.CompressName(networkchaos.Name, 21, "")
				chainDirection = pb.Chain_INPUT
			case "OUTPUT":
				chainName = "OUTPUT/" + netutils.CompressName(networkchaos.Name, 20, "")
				chainDirection = pb.Chain_OUTPUT
			}

			chains = append(chains, &pb.Chain{
				Name:      chainName,
				Direction: chainDirection,
			})
		}

		err = iptable.SetIptablesChains(ctx, r.Client, &pod, chains)
		if err != nil {
			r.Log.Error(err, "error while deleting iptables rules")
			result = multierror.Append(result, err)
			continue
		}

		networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, key)
	}
	r.Log.Info("After recovering", "finalizers", networkchaos.Finalizers)

	if networkchaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", networkchaos)
		networkchaos.Finalizers = networkchaos.Finalizers[:0]
		return nil
	}

	return result
}
