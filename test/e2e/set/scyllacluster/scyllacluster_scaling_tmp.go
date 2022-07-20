// Copyright (c) 2022 ScyllaDB

package scyllacluster

import (
	"context"
	"fmt"
	"os"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	v1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	"github.com/scylladb/scylla-operator/pkg/naming"
	scyllafixture "github.com/scylladb/scylla-operator/test/e2e/fixture/scylla"
	"github.com/scylladb/scylla-operator/test/e2e/framework"
	"github.com/scylladb/scylla-operator/test/e2e/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/drain"
)

var _ = g.Describe("ScyllaCluster", func() {
	defer g.GinkgoRecover()

	f := framework.NewFramework("scyllacluster")

	g.FIt("should support scaling", func() {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		sc := scyllafixture.BasicScyllaCluster.ReadOrFail()
		sc.Spec.Sysctls = []string{"fs.aio-max-nr=5578536"}
		sc.Spec.Datacenter.Racks[0].Members = 3
		sc.Spec.Datacenter.Racks[0].Placement = &v1.PlacementSpec{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "minimal-k8s-nodepool",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"scylla-pool"},
								},
							},
						},
					},
				},
			},
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "app.kubernetes.io/name",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"scylla"},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		}

		framework.By("Creating a ScyllaCluster with 3 members")
		sc, err := f.ScyllaClient().ScyllaV1().ScyllaClusters(f.Namespace()).Create(ctx, sc, metav1.CreateOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())

		framework.By("Waiting for the ScyllaCluster to rollout")
		waitCtx1, waitCtx1Cancel := utils.ContextForRollout(ctx, sc)
		defer waitCtx1Cancel()
		sc, err = utils.WaitForScyllaClusterState(waitCtx1, f.ScyllaClient().ScyllaV1(), sc.Namespace, sc.Name, utils.IsScyllaClusterRolledOut)
		o.Expect(err).NotTo(o.HaveOccurred())

		framework.By("Cordon and drain k8s node")
		waitCtx3, waitCtx3Cancel := utils.ContextForRollout(ctx, sc)
		defer waitCtx3Cancel()
		podName := fmt.Sprintf("%s-%d", naming.StatefulSetNameForRack(sc.Spec.Datacenter.Racks[0], sc), sc.Spec.Datacenter.Racks[0].Members-1)

		pod, err := f.KubeClient().CoreV1().Pods(sc.Namespace).Get(waitCtx3, podName, metav1.GetOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())

		nodeName := pod.Spec.NodeName
		o.Expect(nodeName).NotTo(o.BeEmpty())

		waitCtx6, waitCtx6Cancel := utils.ContextForRollout(ctx, sc)
		defer waitCtx6Cancel()
		node, err := f.KubeAdminClient().CoreV1().Nodes().Get(waitCtx6, nodeName, metav1.GetOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(node).NotTo(o.BeNil())

		waitCtx8, waitCtx8Cancel := utils.ContextForRollout(ctx, sc)
		defer waitCtx8Cancel()
		helper := &drain.Helper{
			Ctx:                 waitCtx8,
			Client:              f.KubeAdminClient(),
			Force:               true,
			Timeout:             testTimeout,
			GracePeriodSeconds:  -1,
			IgnoreAllDaemonSets: true,
			DeleteEmptyDirData:  true,
			Out:                 os.Stdout,
			ErrOut:              os.Stderr,
		}

		//err = drain.RunCordonOrUncordon(helper, node, true)
		//o.Expect(err).NotTo(o.HaveOccurred())

		err = drain.RunNodeDrain(helper, nodeName)
		o.Expect(err).NotTo(o.HaveOccurred())

		time.Sleep(5 * time.Second)

		err = drain.RunCordonOrUncordon(helper, node, false)

		framework.By("Scaling the ScyllaCluster down to 2 replicas (decommissioning)")
		sc, err = f.ScyllaClient().ScyllaV1().ScyllaClusters(sc.Namespace).Patch(
			ctx,
			sc.Name,
			types.JSONPatchType,
			[]byte(`[{"op": "replace", "path": "/spec/datacenter/racks/0/members", "value": 2}]`),
			metav1.PatchOptions{},
		)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(sc.Spec.Datacenter.Racks[0].Members).To(o.BeEquivalentTo(2))

		framework.By("waiting for pod's deletion")
		waitCtx4, waitCtx4Cancel := utils.ContextForRollout(ctx, sc)
		defer waitCtx4Cancel()
		pod, err = f.KubeClient().CoreV1().Pods(sc.Namespace).Get(waitCtx4, podName, metav1.GetOptions{})
		if err == nil || !apierrors.IsNotFound(err) {
			err = framework.WaitForObjectDeletion(ctx, f.DynamicAdminClient(), corev1.SchemeGroupVersion.WithResource("pods"), sc.Namespace, pod.Name, &pod.UID)
		}
		o.Expect(err).NotTo(o.HaveOccurred())

		framework.By("Scaling the ScyllaCluster back to 3 replicas")
		sc, err = f.ScyllaClient().ScyllaV1().ScyllaClusters(f.Namespace()).Patch(
			ctx,
			sc.Name,
			types.JSONPatchType,
			[]byte(`[{"op": "replace", "path": "/spec/datacenter/racks/0/members", "value": 3}]`),
			metav1.PatchOptions{},
		)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(sc.Spec.Datacenter.Racks).To(o.HaveLen(1))
		o.Expect(sc.Spec.Datacenter.Racks[0].Members).To(o.BeEquivalentTo(3))

		framework.By("Waiting for the ScyllaCluster to rollout")
		waitCtx5, waitCtx5Cancel := utils.ContextForRollout(ctx, sc)
		defer waitCtx5Cancel()
		sc, err = utils.WaitForScyllaClusterState(waitCtx5, f.ScyllaClient().ScyllaV1(), sc.Namespace, sc.Name, utils.IsScyllaClusterRolledOut)
		o.Expect(err).NotTo(o.HaveOccurred())

	})
})
