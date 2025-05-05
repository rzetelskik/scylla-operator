// Copyright (C) 2025 ScyllaDB

package scylladbmanagertask

import (
	g "github.com/onsi/ginkgo/v2"
	"github.com/scylladb/scylla-operator/test/e2e/framework"
)

var _ = g.Describe("ScyllaDBManagerTask integration with global ScyllaDB Manager", func() {
	f := framework.NewFramework("scylladbmanagertask")

	// TODO: creation/deletion of repair/backup for sdc with global manager
})
