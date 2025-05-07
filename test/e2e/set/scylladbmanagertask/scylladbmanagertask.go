// Copyright (C) 2025 ScyllaDB

package scylladbmanagertask

import (
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	"github.com/scylladb/scylla-manager/v3/pkg/util/uuid"
	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/test/e2e/framework"
	"github.com/scylladb/scylla-operator/test/e2e/utils"
	utilsv1alpha1 "github.com/scylladb/scylla-operator/test/e2e/utils/v1alpha1"
	scylladbdatacenterverification "github.com/scylladb/scylla-operator/test/e2e/utils/verification/scylladbdatacenter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = g.FDescribe("ScyllaDBManagerTask integration with global ScyllaDB Manager", func() {
	f := framework.NewFramework("scylladbmanagertask")

	// TODO: creation/deletion of repair/backup for sdc with global manager
	g.FIt("TODO", func(ctx g.SpecContext) {
		ns, nsClient, ok := f.DefaultNamespaceIfAny()
		o.Expect(ok).To(o.BeTrue())

		sdc := f.GetDefaultScyllaDBDatacenter()
		metav1.SetMetaDataLabel(&sdc.ObjectMeta, naming.GlobalScyllaDBManagerRegistrationLabel, naming.LabelValueTrue)

		framework.By(`Creating a ScyllaDBDatacenter with the global ScyllaDB Manager registration label`)
		sdc, err := nsClient.ScyllaClient().ScyllaV1alpha1().ScyllaDBDatacenters(ns.Name).Create(ctx, sdc, metav1.CreateOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())

		framework.By("Waiting for ScyllaDBDatacenter to roll out (RV=%s)", sdc.ResourceVersion)
		rolloutCtx, rolloutCtxCancel := utilsv1alpha1.ContextForRollout(ctx, sdc)
		defer rolloutCtxCancel()
		sdc, err = controllerhelpers.WaitForScyllaDBDatacenterState(rolloutCtx, nsClient.ScyllaClient().ScyllaV1alpha1().ScyllaDBDatacenters(ns.Name), sdc.Name, controllerhelpers.WaitForStateOptions{}, utilsv1alpha1.IsScyllaDBDatacenterRolledOut)
		o.Expect(err).NotTo(o.HaveOccurred())

		scylladbdatacenterverification.Verify(ctx, nsClient.KubeClient(), nsClient.ScyllaClient(), sdc)
		scylladbdatacenterverification.WaitForFullQuorum(ctx, nsClient.KubeClient().CoreV1(), sdc)

		hosts, err := utilsv1alpha1.GetBroadcastRPCAddresses(ctx, f.KubeClient().CoreV1(), sdc)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(hosts).To(o.HaveLen(1))
		di := scylladbdatacenterverification.InsertAndVerifyCQLData(ctx, hosts)
		defer di.Close()

		smt := &scyllav1alpha1.ScyllaDBManagerTask{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "repair",
				Namespace: ns.Name,
			},
			Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
				ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
					Kind: naming.ScyllaDBDatacenterKind,
					Name: sdc.Name,
				},
				Type:   scyllav1alpha1.ScyllaDBManagerTaskTypeRepair,
				Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{},
			},
		}

		framework.By("Creating a repair ScyllaDBManagerTask")

		smt, err = nsClient.ScyllaClient().ScyllaV1alpha1().ScyllaDBManagerTasks(ns.Name).Create(
			ctx,
			smt,
			metav1.CreateOptions{},
		)
		o.Expect(err).NotTo(o.HaveOccurred())

		framework.By("Waiting for ScyllaDBManagerTask to register with global ScyllaDB Manager instance")
		// TODO: waitctx
		// TODO: has deletion finalizer
		smt, err = controllerhelpers.WaitForScyllaDBManagerTaskState(
			ctx,
			nsClient.ScyllaClient().ScyllaV1alpha1().ScyllaDBManagerTasks(ns.Name),
			smt.Name,
			controllerhelpers.WaitForStateOptions{},
			utilsv1alpha1.IsScyllaDBManagerTaskRolledOut,
		)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(smt.Status.TaskID).NotTo(o.BeNil())
		o.Expect(*smt.Status.TaskID).NotTo(o.BeEmpty())
		managerTaskID, err := uuid.Parse(*smt.Status.TaskID)
		o.Expect(err).NotTo(o.HaveOccurred())

		smcrName, err := naming.ScyllaDBManagerClusterRegistrationNameForScyllaDBDatacenter(sdc)
		o.Expect(err).NotTo(o.HaveOccurred())
		smcr, err := nsClient.ScyllaClient().ScyllaV1alpha1().ScyllaDBManagerClusterRegistrations(ns.Name).Get(ctx, smcrName, metav1.GetOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(smcr.Status.ClusterID).NotTo(o.BeNil())
		o.Expect(*smcr.Status.ClusterID).NotTo(o.BeEmpty())
		managerClusterID := *smcr.Status.ClusterID

		managerClient, err := utils.GetManagerClient(ctx, f.KubeAdminClient().CoreV1())
		o.Expect(err).NotTo(o.HaveOccurred())

		framework.By("Verifying that ScyllaDBManagerTask was registered with global ScyllaDB Manager")
		// TODO: task type
		managerTask, err := managerClient.GetTask(ctx, managerClusterID, "repair", managerTaskID)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(managerTask.Labels).NotTo(o.BeNil())
		o.Expect(managerTask.Labels[naming.OwnerUIDLabel]).To(o.Equal(string(smt.UID)))

		// TODO: verify parameters

	}, g.SpecTimeout(testTimeout))
})
