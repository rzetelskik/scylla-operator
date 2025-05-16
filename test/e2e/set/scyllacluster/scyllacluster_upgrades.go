// Copyright (C) 2021 ScyllaDB

package scyllacluster

import (
	"context"
	"fmt"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	"github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"github.com/scylladb/scylla-operator/test/e2e/framework"
	"github.com/scylladb/scylla-operator/test/e2e/utils"
	"github.com/scylladb/scylla-operator/test/e2e/utils/verification"
	scyllaclusterverification "github.com/scylladb/scylla-operator/test/e2e/utils/verification/scyllacluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = g.Describe("ScyllaCluster upgrades", func() {
	f := framework.NewFramework("scyllacluster")

	type entry struct {
		rackSize       int32
		rackCount      int32
		initialVersion string
		updatedVersion string
	}

	describeEntry := func(e *entry) string {
		return fmt.Sprintf("with %d member(s) and %d rack(s) from %s to %s", e.rackSize, e.rackCount, e.initialVersion, e.updatedVersion)
	}

	g.DescribeTable("should deploy and update",
		func(e *entry) {
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			sc := f.GetDefaultScyllaCluster()
			sc.Spec.Version = e.initialVersion

			o.Expect(sc.Spec.Datacenter.Racks).To(o.HaveLen(1))
			rack := &sc.Spec.Datacenter.Racks[0]
			sc.Spec.Datacenter.Racks = make([]scyllav1.RackSpec, 0, e.rackCount)
			for i := int32(0); i < e.rackCount; i++ {
				r := rack.DeepCopy()
				r.Name = fmt.Sprintf("rack-%d", i)
				r.Members = e.rackSize
				sc.Spec.Datacenter.Racks = append(sc.Spec.Datacenter.Racks, *r)
			}

			framework.By("Creating a ScyllaCluster")
			sc, err := f.ScyllaClient().ScyllaV1().ScyllaClusters(f.Namespace()).Create(ctx, sc, metav1.CreateOptions{})
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(sc.Spec.Version).To(o.Equal(e.initialVersion))

			framework.By("Waiting for the ScyllaCluster to roll out (RV=%s)", sc.ResourceVersion)
			waitCtx1, waitCtx1Cancel := utils.ContextForRollout(ctx, sc)
			defer waitCtx1Cancel()
			sc, err = controllerhelpers.WaitForScyllaClusterState(waitCtx1, f.ScyllaClient().ScyllaV1().ScyllaClusters(sc.Namespace), sc.Name, controllerhelpers.WaitForStateOptions{}, utils.IsScyllaClusterRolledOut)
			o.Expect(err).NotTo(o.HaveOccurred())

			scyllaclusterverification.Verify(ctx, f.KubeClient(), f.ScyllaClient(), sc)
			scyllaclusterverification.WaitForFullQuorum(ctx, f.KubeClient().CoreV1(), sc)

			hosts, hostIDs, err := utils.GetBroadcastRPCAddressesAndUUIDs(ctx, f.KubeClient().CoreV1(), sc)
			o.Expect(err).NotTo(o.HaveOccurred())

			numNodes := int(e.rackCount * e.rackSize)
			o.Expect(hosts).To(o.HaveLen(numNodes))
			o.Expect(hostIDs).To(o.HaveLen(numNodes))

			di := verification.InsertAndVerifyCQLData(ctx, hosts)
			defer di.Close()

			framework.By("triggering and update")
			sc, err = f.ScyllaClient().ScyllaV1().ScyllaClusters(f.Namespace()).Patch(
				ctx,
				sc.Name,
				types.MergePatchType,
				[]byte(fmt.Sprintf(
					`{"metadata":{"uid":"%s"},"spec":{"version":"%s"}}`,
					sc.UID,
					e.updatedVersion,
				)),
				metav1.PatchOptions{},
			)
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(sc.Spec.Version).To(o.Equal(e.updatedVersion))

			framework.By("Waiting for the ScyllaCluster to re-deploy")
			waitCtx2, waitCtx2Cancel := utils.ContextForRollout(ctx, sc)
			defer waitCtx2Cancel()
			sc, err = controllerhelpers.WaitForScyllaClusterState(waitCtx2, f.ScyllaClient().ScyllaV1().ScyllaClusters(sc.Namespace), sc.Name, controllerhelpers.WaitForStateOptions{}, utils.IsScyllaClusterRolledOut)
			o.Expect(err).NotTo(o.HaveOccurred())

			scyllaclusterverification.Verify(ctx, f.KubeClient(), f.ScyllaClient(), sc)
			scyllaclusterverification.WaitForFullQuorum(ctx, f.KubeClient().CoreV1(), sc)

			oldHosts := hosts
			oldHostIDs := hostIDs
			hosts, hostIDs, err = utils.GetBroadcastRPCAddressesAndUUIDs(ctx, f.KubeClient().CoreV1(), sc)
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(hosts).To(o.HaveLen(len(oldHosts)))
			o.Expect(hostIDs).To(o.ConsistOf(oldHostIDs))
			// Reset hosts if there's only one node as the client won't be able to discover it.
			if numNodes == 1 {
				err = di.SetClientEndpoints(hosts)
				o.Expect(err).NotTo(o.HaveOccurred())
			}
			verification.VerifyCQLData(ctx, di)

			framework.By("Validating if all snapshots created by upgrade hooks are cleared for every node")
			scyllaClient, hosts, err := utils.GetScyllaClient(ctx, f.KubeClient().CoreV1(), sc)
			o.Expect(err).NotTo(o.HaveOccurred())

			for _, host := range hosts {
				snapshots, err := scyllaClient.ListSnapshots(ctx, host)
				o.Expect(err).NotTo(o.HaveOccurred())

				o.Expect(snapshots).To(o.BeEmpty())
			}
		},
		// Test 1 and 3 member rack to cover e.g. handling PDBs correctly.
		g.Entry(describeEntry, &entry{
			rackCount:      1,
			rackSize:       1,
			initialVersion: framework.TestContext.ScyllaDBUpdateFrom,
			updatedVersion: framework.TestContext.ScyllaDBVersion,
		}),
		g.Entry(describeEntry, &entry{
			rackCount:      1,
			rackSize:       3,
			initialVersion: framework.TestContext.ScyllaDBUpdateFrom,
			updatedVersion: framework.TestContext.ScyllaDBVersion,
		}),
		g.Entry(describeEntry, &entry{
			rackCount:      1,
			rackSize:       1,
			initialVersion: framework.TestContext.ScyllaDBUpgradeFrom,
			updatedVersion: framework.TestContext.ScyllaDBVersion,
		}),
		g.Entry(describeEntry, &entry{
			rackCount:      1,
			rackSize:       3,
			initialVersion: framework.TestContext.ScyllaDBUpgradeFrom,
			updatedVersion: framework.TestContext.ScyllaDBVersion,
		}),
		g.Entry(describeEntry, &entry{
			rackCount:      2,
			rackSize:       3,
			initialVersion: framework.TestContext.ScyllaDBUpgradeFrom,
			updatedVersion: framework.TestContext.ScyllaDBVersion,
		}),
	)
})
