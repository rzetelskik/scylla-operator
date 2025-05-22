// Copyright (C) 2025 ScyllaDB

package scylladbmanagertask

const (
	managerControllerProgressingCondition            = "ManagerControllerProgressing"
	managerControllerDegradedCondition               = "ManagerControllerDegraded"
	scyllaDBManagerTaskFinalizerProgressingCondition = "ScyllaDBManagerTaskFinalizerProgressing"
	scyllaDBManagerTaskFinalizerDegradedCondition    = "ScyllaDBManagerTaskFinalizerDegraded"

	// Hackathon
	jobControllerProgressingCondition = "JobControllerProgressing"
	jobControllerDegradedCondition    = "JobControllerDegraded"
)
