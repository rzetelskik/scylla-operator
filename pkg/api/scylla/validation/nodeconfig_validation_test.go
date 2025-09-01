// Copyright (c) 2023 ScyllaDB.

package validation_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/api/scylla/validation"
	"github.com/scylladb/scylla-operator/pkg/test/unit"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestValidateNodeConfig(t *testing.T) {
	t.Parallel()

	validNodeConfig := unit.ValidNodeConfig.ReadOrFail()

	tt := []struct {
		name                string
		nodeConfig          *scyllav1alpha1.NodeConfig
		expectedErrorList   field.ErrorList
		expectedErrorString string
	}{
		{
			name:                "valid",
			nodeConfig:          validNodeConfig,
			expectedErrorList:   nil,
			expectedErrorString: "",
		},
		{
			name: "duplicate raid device names",
			nodeConfig: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.LocalDiskSetup.RAIDs = append(nc.Spec.LocalDiskSetup.RAIDs, *nc.Spec.LocalDiskSetup.RAIDs[0].DeepCopy())
				return nc
			}(),
			expectedErrorList: field.ErrorList{
				&field.Error{Type: field.ErrorTypeDuplicate, Field: "spec.localDiskSetup.raids[1].name", BadValue: "nvmes"},
			},
			expectedErrorString: `spec.localDiskSetup.raids[1].name: Duplicate value: "nvmes"`,
		},
		{
			name: "duplicate mount points",
			nodeConfig: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.LocalDiskSetup.Mounts = append(nc.Spec.LocalDiskSetup.Mounts, *nc.Spec.LocalDiskSetup.Mounts[0].DeepCopy())
				return nc
			}(),
			expectedErrorList: field.ErrorList{
				&field.Error{Type: field.ErrorTypeDuplicate, Field: "spec.localDiskSetup.mounts[1].mountPoint", BadValue: "/var/lib/persistent-volumes"},
			},
			expectedErrorString: `spec.localDiskSetup.mounts[1].mountPoint: Duplicate value: "/var/lib/persistent-volumes"`,
		},
		{
			name: "raid type specified but without configuration",
			nodeConfig: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.LocalDiskSetup.RAIDs[0].Type = scyllav1alpha1.RAID0Type
				nc.Spec.LocalDiskSetup.RAIDs[0].RAID0 = nil
				return nc
			}(),
			expectedErrorList: field.ErrorList{
				&field.Error{Type: field.ErrorTypeInvalid, Field: "spec.localDiskSetup.raids[0].RAID0", BadValue: "", Detail: "RAID0 options must be provided when RAID0 type is set"},
			},
			expectedErrorString: `spec.localDiskSetup.raids[0].RAID0: Invalid value: "": RAID0 options must be provided when RAID0 type is set`,
		},
		{
			name: "name or model regexp must be provided in RAID0 configuration",
			nodeConfig: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.LocalDiskSetup.RAIDs[0].Type = scyllav1alpha1.RAID0Type
				nc.Spec.LocalDiskSetup.RAIDs[0].RAID0 = &scyllav1alpha1.RAID0Options{
					Devices: scyllav1alpha1.DeviceDiscovery{
						NameRegex:  "",
						ModelRegex: "",
					},
				}
				return nc
			}(),
			expectedErrorList: field.ErrorList{
				&field.Error{Type: field.ErrorTypeInvalid, Field: "spec.localDiskSetup.raids[0].RAID0.devices", BadValue: "", Detail: "nameRegex or modelRegex must be provided"},
			},
			expectedErrorString: `spec.localDiskSetup.raids[0].RAID0.devices: Invalid value: "": nameRegex or modelRegex must be provided`,
		},
		{
			name: "name regexp can be empty when model regexp is provided in RAID0 configuration",
			nodeConfig: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.LocalDiskSetup.RAIDs[0].Type = scyllav1alpha1.RAID0Type
				nc.Spec.LocalDiskSetup.RAIDs[0].RAID0 = &scyllav1alpha1.RAID0Options{
					Devices: scyllav1alpha1.DeviceDiscovery{
						NameRegex:  "",
						ModelRegex: ".*",
					},
				}
				return nc
			}(),
			expectedErrorList:   nil,
			expectedErrorString: "",
		},
		{
			name: "model regexp can be empty when name regexp is provided in RAID0 configuration",
			nodeConfig: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.LocalDiskSetup.RAIDs[0].Type = scyllav1alpha1.RAID0Type
				nc.Spec.LocalDiskSetup.RAIDs[0].RAID0 = &scyllav1alpha1.RAID0Options{
					Devices: scyllav1alpha1.DeviceDiscovery{
						NameRegex:  ".*",
						ModelRegex: "",
					},
				}
				return nc
			}(),
			expectedErrorList:   nil,
			expectedErrorString: "",
		},
		{
			name: "empty sysctl name",
			nodeConfig: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.Sysctls = []corev1.Sysctl{
					{
						Name:  "",
						Value: "1",
					},
				}
				return nc
			}(),
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeRequired,
					Field:    "spec.sysctls[0].name",
					BadValue: "",
					Detail:   "",
				},
			},
			expectedErrorString: `spec.sysctls[0].name: Required value`,
		},
		{
			name: "invalid sysctl name",
			nodeConfig: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.Sysctls = []corev1.Sysctl{
					{
						Name:  "invalid..name",
						Value: "1",
					},
				}
				return nc
			}(),
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.sysctls[0].name",
					BadValue: "invalid..name",
					Detail:   `must have at most 253 characters and match regex ^([a-z0-9]([-_a-z0-9]*[a-z0-9])?[\./])*[a-z0-9]([-_a-z0-9]*[a-z0-9])?$`,
				},
			},
			expectedErrorString: `spec.sysctls[0].name: Invalid value: "invalid..name": must have at most 253 characters and match regex ^([a-z0-9]([-_a-z0-9]*[a-z0-9])?[\./])*[a-z0-9]([-_a-z0-9]*[a-z0-9])?$`,
		},
		{
			name: "duplicated sysctl name",
			nodeConfig: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.Sysctls = []corev1.Sysctl{
					{
						Name:  "fs.aio-max-nr",
						Value: "30000000",
					},
					{
						Name:  "fs.aio-max-nr",
						Value: "2097152",
					},
				}
				return nc
			}(),
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeDuplicate,
					Field:    "spec.sysctls[1].name",
					BadValue: "fs.aio-max-nr",
					Detail:   "",
				},
			},
			expectedErrorString: `spec.sysctls[1].name: Duplicate value: "fs.aio-max-nr"`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			errList := validation.ValidateNodeConfig(tc.nodeConfig)
			if !reflect.DeepEqual(errList, tc.expectedErrorList) {
				t.Errorf("expected and actual error lists differ: %s", cmp.Diff(tc.expectedErrorList, errList))
			}

			errStr := ""
			agg := errList.ToAggregate()
			if agg != nil {
				errStr = agg.Error()
			}
			if !reflect.DeepEqual(errStr, tc.expectedErrorString) {
				t.Errorf("expected and actual error strings differ: %s", cmp.Diff(tc.expectedErrorString, errStr))
			}
		})
	}
}

func TestValidateNodeConfigUpdate(t *testing.T) {
	t.Parallel()

	validNodeConfig := unit.ValidNodeConfig.ReadOrFail()

	tt := []struct {
		name                string
		old                 *scyllav1alpha1.NodeConfig
		new                 *scyllav1alpha1.NodeConfig
		expectedErrorList   field.ErrorList
		expectedErrorString string
	}{
		{
			name:                "identity",
			old:                 validNodeConfig,
			new:                 validNodeConfig,
			expectedErrorList:   nil,
			expectedErrorString: "",
		},
		{
			name: "adding a duplicate raid name",
			old: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.LocalDiskSetup.RAIDs = []scyllav1alpha1.RAIDConfiguration{
					{
						Name: "foo",
					},
					{
						Name: "bar",
					},
				}
				return nc
			}(),
			new: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.LocalDiskSetup.RAIDs = []scyllav1alpha1.RAIDConfiguration{
					{
						Name: "foo",
					},
					{
						Name: "foo",
					},
					{
						Name: "bar",
					},
					{
						Name: "bar",
					},
				}
				return nc
			}(),
			expectedErrorList: field.ErrorList{
				&field.Error{Type: field.ErrorTypeDuplicate, Field: "spec.localDiskSetup.raids[1].name", BadValue: "foo"},
				&field.Error{Type: field.ErrorTypeDuplicate, Field: "spec.localDiskSetup.raids[3].name", BadValue: "bar"},
			},
			expectedErrorString: `[spec.localDiskSetup.raids[1].name: Duplicate value: "foo", spec.localDiskSetup.raids[3].name: Duplicate value: "bar"]`,
		},
		{
			name: "adding a mount with duplicate mount point",
			old: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.LocalDiskSetup.Mounts = []scyllav1alpha1.MountConfiguration{
					{
						MountPoint: "/mnt/foo",
					},
					{
						MountPoint: "/mnt/bar",
					},
				}
				return nc
			}(),
			new: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.LocalDiskSetup.Mounts = []scyllav1alpha1.MountConfiguration{
					{
						MountPoint: "/mnt/foo",
					},
					{
						MountPoint: "/mnt/foo",
					},
					{
						MountPoint: "/mnt/bar",
					},
					{
						MountPoint: "/mnt/bar",
					},
				}
				return nc
			}(),
			expectedErrorList: field.ErrorList{
				&field.Error{Type: field.ErrorTypeDuplicate, Field: "spec.localDiskSetup.mounts[1].mountPoint", BadValue: "/mnt/foo"},
				&field.Error{Type: field.ErrorTypeDuplicate, Field: "spec.localDiskSetup.mounts[3].mountPoint", BadValue: "/mnt/bar"},
			},
			expectedErrorString: `[spec.localDiskSetup.mounts[1].mountPoint: Duplicate value: "/mnt/foo", spec.localDiskSetup.mounts[3].mountPoint: Duplicate value: "/mnt/bar"]`,
		},
		{
			name: "immutable loop device size",
			old: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.LocalDiskSetup.LoopDevices = []scyllav1alpha1.LoopDeviceConfiguration{
					{
						Name:      "foo",
						ImagePath: "/mnt/foo.img",
						Size:      resource.MustParse("100Mi"),
					},
				}
				return nc
			}(),
			new: func() *scyllav1alpha1.NodeConfig {
				nc := validNodeConfig.DeepCopy()
				nc.Spec.LocalDiskSetup.LoopDevices = []scyllav1alpha1.LoopDeviceConfiguration{
					{
						Name:      "foo",
						ImagePath: "/mnt/foo.img",
						Size:      resource.MustParse("200Mi"),
					},
				}
				return nc
			}(),
			expectedErrorList: field.ErrorList{
				&field.Error{Type: field.ErrorTypeInvalid, Field: "spec.localDiskSetup.loopDevices[0].size", BadValue: "200Mi", Detail: "field is immutable"},
			},
			expectedErrorString: fmt.Sprintf(`spec.localDiskSetup.loopDevices[0].size: Invalid value: "200Mi": field is immutable`),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			errList := validation.ValidateNodeConfigUpdate(tc.new, tc.old)
			if !reflect.DeepEqual(errList, tc.expectedErrorList) {
				t.Errorf("expected and actual error lists differ: %s", cmp.Diff(tc.expectedErrorList, errList))
			}

			errStr := ""
			agg := errList.ToAggregate()
			if agg != nil {
				errStr = agg.Error()
			}
			if !reflect.DeepEqual(errStr, tc.expectedErrorString) {
				t.Errorf("expected and actual error strings differ: %s", cmp.Diff(tc.expectedErrorString, errStr))
			}
		})
	}
}
