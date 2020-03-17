package twophase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/reconciler"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
	"github.com/pingcap/chaos-mesh/pkg/mock"
)

func TestTwoPhase(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"TwoPhase Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	Expect(addFakeToScheme(scheme.Scheme)).To(Succeed())

	close(done)
}, 60)

var _ = AfterSuite(func() {
})

var _ reconciler.InnerReconciler = (*fakeReconciler)(nil)

type fakeReconciler struct{}

func (r fakeReconciler) Apply(ctx context.Context, req ctrl.Request, chaos reconciler.InnerObject) error {
	if err := mock.On("MockApplyError"); err != nil {
		return err.(error)
	}
	return nil
}

func (r fakeReconciler) Recover(ctx context.Context, req ctrl.Request, chaos reconciler.InnerObject) error {
	if err := mock.On("MockRecoverError"); err != nil {
		return err.(error)
	}
	return nil
}

var _ twophase.InnerSchedulerObject = (*fakeTwoPhaseChaos)(nil)

type fakeTwoPhaseChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Status            v1alpha1.ChaosStatus `json:"status,omitempty"`

	// Selector is used to select pods that are used to inject chaos action.
	Selector v1alpha1.SelectorSpec `json:"selector"`

	Deleted bool `json:"deleted"`

	// Duration represents the duration of the chaos action
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	Scheduler *v1alpha1.SchedulerSpec `json:"scheduler,omitempty"`

	// Next time when this action will be applied again
	// +optional
	NextStart *metav1.Time `json:"nextStart,omitempty"`

	// Next time when this action will be recovered
	// +optional
	NextRecover *metav1.Time `json:"nextRecover,omitempty"`
}

func (in *fakeTwoPhaseChaos) GetStatus() *v1alpha1.ChaosStatus {
	return &in.Status
}

func (in *fakeTwoPhaseChaos) IsDeleted() bool {
	return in.Deleted
}

func (r fakeReconciler) Object() reconciler.InnerObject {
	return &fakeTwoPhaseChaos{}
}

func (in *fakeTwoPhaseChaos) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *fakeTwoPhaseChaos) GetNextStart() time.Time {
	if in.NextStart == nil {
		return time.Time{}
	}
	return in.NextStart.Time
}

func (in *fakeTwoPhaseChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.NextStart = nil
		return
	}

	if in.NextStart == nil {
		in.NextStart = &metav1.Time{}
	}
	in.NextStart.Time = t
}

func (in *fakeTwoPhaseChaos) GetNextRecover() time.Time {
	if in.NextRecover == nil {
		return time.Time{}
	}
	return in.NextRecover.Time
}

func (in *fakeTwoPhaseChaos) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.NextRecover = nil
		return
	}

	if in.NextRecover == nil {
		in.NextRecover = &metav1.Time{}
	}
	in.NextRecover.Time = t
}

func (in *fakeTwoPhaseChaos) GetScheduler() *v1alpha1.SchedulerSpec {
	return in.Scheduler
}

func (in *fakeTwoPhaseChaos) DeepCopyInto(out *fakeTwoPhaseChaos) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)

	in.Status.DeepCopyInto(&out.Status)
	in.Selector.DeepCopyInto(&out.Selector)

	out.Deleted = in.Deleted

	if in.Duration != nil {
		in, out := &in.Duration, &out.Duration
		*out = new(string)
		**out = **in
	}
	if in.Scheduler != nil {
		in, out := &in.Scheduler, &out.Scheduler
		*out = new(v1alpha1.SchedulerSpec)
		**out = **in
	}
	if in.NextRecover != nil {
		in, out := &in.NextRecover, &out.NextRecover
		*out = new(metav1.Time)
		**out = **in
	}
	if in.NextStart != nil {
		in, out := &in.NextStart, &out.NextStart
		*out = new(metav1.Time)
		**out = **in
	}
}

func (in *fakeTwoPhaseChaos) DeepCopy() *fakeTwoPhaseChaos {
	if in == nil {
		return nil
	}
	out := new(fakeTwoPhaseChaos)
	in.DeepCopyInto(out)
	return out
}

func (in *fakeTwoPhaseChaos) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

var (
	schemeBuilder   = runtime.NewSchemeBuilder(addKnownTypes)
	addFakeToScheme = schemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(schema.GroupVersion{Group: "", Version: "v1"},
		&fakeTwoPhaseChaos{},
	)
	return nil
}

var _ = Describe("TwoPhase", func() {
	Context("TwoPhase", func() {
		var err error

		zeroTime := time.Time{}
		var _ = zeroTime
		pastTime := time.Now().Add(-10 * time.Hour)
		futureTime := time.Now().Add(10 * time.Hour)

		req := ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      "fakechaos-name",
				Namespace: metav1.NamespaceDefault,
			},
		}

		typeMeta := metav1.TypeMeta{
			Kind:       "PodChaos",
			APIVersion: "v1",
		}
		objectMeta := metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      "fakechaos-name",
		}

		It("TwoPhase Action", func() {
			chaos := fakeTwoPhaseChaos{
				TypeMeta:   typeMeta,
				ObjectMeta: objectMeta,
			}

			c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos)

			r := twophase.Reconciler{
				InnerReconciler: fakeReconciler{},
				Client:          c,
				Log:             ctrl.Log.WithName("controllers").WithName("TwoPhase"),
			}

			_, err = r.Reconcile(req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("misdefined scheduler"))
		})

		It("TwoPhase Delete", func() {
			chaos := fakeTwoPhaseChaos{
				TypeMeta:   typeMeta,
				ObjectMeta: objectMeta,
				Scheduler:  &v1alpha1.SchedulerSpec{Cron: "@hourly"},
				Deleted:    true,
			}

			c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos)

			r := twophase.Reconciler{
				InnerReconciler: fakeReconciler{},
				Client:          c,
				Log:             ctrl.Log.WithName("controllers").WithName("TwoPhase"),
			}

			_, err = r.Reconcile(req)

			Expect(err).ToNot(HaveOccurred())
			_chaos := r.Object()
			err = r.Get(context.TODO(), req.NamespacedName, _chaos)
			Expect(err).ToNot(HaveOccurred())
			Expect(_chaos.(twophase.InnerSchedulerObject).GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhaseFinished))

			defer mock.With("MockRecoverError", errors.New("RecoverError"))()

			_, err = r.Reconcile(req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("RecoverError"))
		})

		It("TwoPhase ToRecover", func() {
			chaos := fakeTwoPhaseChaos{
				TypeMeta:   typeMeta,
				ObjectMeta: objectMeta,
				Scheduler:  &v1alpha1.SchedulerSpec{Cron: "@hourly"},
			}

			chaos.SetNextRecover(pastTime)

			c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos)

			r := twophase.Reconciler{
				InnerReconciler: fakeReconciler{},
				Client:          c,
				Log:             ctrl.Log.WithName("controllers").WithName("TwoPhase"),
			}

			_, err = r.Reconcile(req)

			Expect(err).ToNot(HaveOccurred())
			_chaos := r.Object()
			err = r.Get(context.TODO(), req.NamespacedName, _chaos)
			Expect(err).ToNot(HaveOccurred())
			Expect(_chaos.(twophase.InnerSchedulerObject).GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhaseFinished))
		})

		It("TwoPhase ToRecover Error", func() {
			chaos := fakeTwoPhaseChaos{
				TypeMeta:   typeMeta,
				ObjectMeta: objectMeta,
				Scheduler:  &v1alpha1.SchedulerSpec{Cron: "@hourly"},
			}

			defer mock.With("MockRecoverError", errors.New("RecoverError"))()
			chaos.SetNextRecover(pastTime)

			c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos)

			r := twophase.Reconciler{
				InnerReconciler: fakeReconciler{},
				Client:          c,
				Log:             ctrl.Log.WithName("controllers").WithName("TwoPhase"),
			}

			_, err = r.Reconcile(req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("RecoverError"))
		})

		It("TwoPhase ToApply", func() {
			chaos := fakeTwoPhaseChaos{
				TypeMeta:   typeMeta,
				ObjectMeta: objectMeta,
				Scheduler:  &v1alpha1.SchedulerSpec{Cron: "@hourly"},
			}

			chaos.SetNextRecover(futureTime)
			chaos.SetNextStart(pastTime)

			c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos)

			r := twophase.Reconciler{
				InnerReconciler: fakeReconciler{},
				Client:          c,
				Log:             ctrl.Log.WithName("controllers").WithName("TwoPhase"),
			}

			_, err = r.Reconcile(req)

			Expect(err).ToNot(HaveOccurred())
			_chaos := r.Object()
			err = r.Get(context.TODO(), req.NamespacedName, _chaos)
			Expect(err).ToNot(HaveOccurred())
			Expect(_chaos.(twophase.InnerSchedulerObject).GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhaseRunning))
		})

		It("TwoPhase ToApply Error", func() {
			chaos := fakeTwoPhaseChaos{
				TypeMeta:   typeMeta,
				ObjectMeta: objectMeta,
				Scheduler:  &v1alpha1.SchedulerSpec{Cron: "@hourly"},
			}

			chaos.SetNextRecover(futureTime)
			chaos.SetNextStart(pastTime)

			c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos)

			r := twophase.Reconciler{
				InnerReconciler: fakeReconciler{},
				Client:          c,
				Log:             ctrl.Log.WithName("controllers").WithName("TwoPhase"),
			}

			defer mock.With("MockApplyError", errors.New("ApplyError"))()

			_, err = r.Reconcile(req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ApplyError"))
		})
	})
})
