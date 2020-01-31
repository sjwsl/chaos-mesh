// Copyright 2019 PingCAP, Inc.
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

package podkill

import (
	"context"
	"errors"
	"time"

	"github.com/go-logr/logr"

	"golang.org/x/sync/errgroup"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/pkg/utils"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	podKillActionMsg = "delete pod"
)

type Reconciler struct {
	client.Client
	Log logr.Logger
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var err error
	now := time.Now()

	r.Log.Info("reconciling pod kill")
	ctx := context.Background()

	var podchaos v1alpha1.PodChaos
	if err = r.Get(ctx, req.NamespacedName, &podchaos); err != nil {
		r.Log.Error(err, "unable to get podchaos")
		return ctrl.Result{}, nil
	}

	shouldAct := podchaos.GetNextStart().Before(now)
	if !shouldAct {
		return ctrl.Result{RequeueAfter: podchaos.GetNextStart().Sub(now)}, nil
	}
	pods, err := utils.SelectPods(ctx, r.Client, podchaos.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "fail to get selected pods")
		return ctrl.Result{Requeue: true}, nil
	}

	if len(pods) == 0 {
		err = errors.New("no pod is selected")
		r.Log.Error(err, "no pod is selected")
		return ctrl.Result{Requeue: true}, nil
	}

	filteredPod, err := utils.GeneratePods(pods, podchaos.Spec.Mode, podchaos.Spec.Value)
	if err != nil {
		r.Log.Error(err, "fail to generate pods")
		return ctrl.Result{Requeue: true}, nil
	}

	g := errgroup.Group{}
	for index := range filteredPod {
		pod := &filteredPod[index]
		g.Go(func() error {
			r.Log.Info("Deleting", "namespace", pod.Namespace, "name", pod.Name)

			if err := r.Delete(ctx, pod, &client.DeleteOptions{
				GracePeriodSeconds: new(int64), // PeriodSeconds has to be set specifically
			}); err != nil {
				r.Log.Error(err, "unable to delete pod")
				return err
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return ctrl.Result{}, nil
	}
	next, err := utils.NextTime(*podchaos.Spec.Scheduler, now)
	if err != nil {
		r.Log.Error(err, "failed to get next time")
		return ctrl.Result{}, nil
	}

	podchaos.SetNextStart(*next)

	podchaos.Status.Experiment.StartTime = &metav1.Time{
		Time: now,
	}
	podchaos.Status.Experiment.EndTime = &metav1.Time{
		Time: now,
	}

	podchaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(podchaos.Spec.Action),
			Message:   podKillActionMsg,
		}

		podchaos.Status.Experiment.Pods = append(podchaos.Status.Experiment.Pods, ps)
	}
	if err := r.Update(ctx, &podchaos); err != nil {
		r.Log.Error(err, "unable to update chaosctl status")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}
