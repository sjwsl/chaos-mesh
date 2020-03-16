package podfailure_test

import (
	"context"
	"testing"

	"k8s.io/client-go/kubernetes/scheme"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/podchaos/podfailure"
	. "github.com/pingcap/chaos-mesh/controllers/test"
	"github.com/pingcap/chaos-mesh/pkg/mock"
)

func TestPodFailure(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"PodFailure Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	Expect(v1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(v1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())

	close(done)
}, 60)

var _ = AfterSuite(func() {
})

var _ = Describe("PodChaos", func() {
	Context("PodFailure", func() {
		objs, pods := GenerateNPods("p", 1, v1.PodRunning, metav1.NamespaceDefault, nil, nil, v1.ContainerStatus{
			ContainerID: "fake-container-id",
			Name:        "container-name",
		})

		podChaos := v1alpha1.PodChaos{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodChaos",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: metav1.NamespaceDefault,
				Name:      "podchaos-name",
			},
			Spec: v1alpha1.PodChaosSpec{
				Selector:      v1alpha1.SelectorSpec{Namespaces: []string{metav1.NamespaceDefault}},
				Mode:          v1alpha1.OnePodMode,
				ContainerName: "container-name",
				Scheduler:     &v1alpha1.SchedulerSpec{Cron: "@hourly"},
			},
		}

		r := podfailure.Reconciler{
			Client:        fake.NewFakeClientWithScheme(scheme.Scheme, objs...),
			EventRecorder: &record.FakeRecorder{},
			Log:           ctrl.Log.WithName("controllers").WithName("PodChaos"),
		}

		It("PodFailure Action", func() {
			defer mock.With("MockChaosDaemonClient", &MockChaosDaemonClient{})()
			defer mock.With("MockSelectAndFilterPods", func() []v1.Pod {
				return pods
			})()

			var err error

			err = r.Apply(context.TODO(), ctrl.Request{}, &podChaos)
			Expect(err).ToNot(HaveOccurred())

			err = r.Recover(context.TODO(), ctrl.Request{}, &podChaos)
			Expect(err).ToNot(HaveOccurred())
		})

	})
})
