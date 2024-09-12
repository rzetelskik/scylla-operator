package systemd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/helpers/slices"
	"github.com/scylladb/scylla-operator/pkg/util/hash"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

type NamedUnit struct {
	FileName string
	Data     []byte
}

type managedUnitStatus struct {
	LastAppliedConfigurationHash string `yaml:"lastAppliedConfigurationHash"`
}

type unitManagerStatus struct {
	ManagedUnits        []string                     `yaml:"managedUnits"`
	ManagedUnitStatuses map[string]managedUnitStatus `yaml:"managedUnitStatuses"`
}

func newUnitManagerStatus() *unitManagerStatus {
	return &unitManagerStatus{
		ManagedUnitStatuses: map[string]managedUnitStatus{},
	}
}

func (s *unitManagerStatus) GetManagedUnitStatusOrDefault(name string) managedUnitStatus {
	mus, ok := s.ManagedUnitStatuses[name]
	if !ok {
		return managedUnitStatus{}
	}

	return mus
}

type UnitManager struct {
	rootPath string
	manager  string
}

func NewUnitManagerWithPath(manager, rootPath string) *UnitManager {
	return &UnitManager{
		rootPath: rootPath,
		manager:  manager,
	}
}

func NewUnitManager(manager string) *UnitManager {
	return NewUnitManagerWithPath(manager, "/etc/systemd/system/")
}

func (m *UnitManager) GetUnitPath(name string) string {
	return filepath.Join(m.rootPath, name)
}

func (m *UnitManager) getStatusName() string {
	return fmt.Sprintf(".%s.unit-manager-status.yaml", m.manager)
}

func (m *UnitManager) getStatusPath() string {
	return path.Join(m.rootPath, m.getStatusName())
}

func (m *UnitManager) ReadStatus() (*unitManagerStatus, error) {
	statusFile := m.getStatusPath()
	data, err := os.ReadFile(statusFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return newUnitManagerStatus(), nil
		}
		return nil, fmt.Errorf("can't open status file %q: %w", statusFile, err)
	}

	status := &unitManagerStatus{}
	err = yaml.Unmarshal(data, status)
	if err != nil {
		return nil, fmt.Errorf("can't decode unit manager status: %w", err)
	}

	return status, nil
}

func (m *UnitManager) WriteStatus(status *unitManagerStatus) error {
	data, err := yaml.Marshal(status)
	if err != nil {
		return fmt.Errorf("can't encode unit manager status: %w", err)
	}

	statusFile := m.getStatusPath()
	err = os.WriteFile(statusFile, data, 0666)
	if err != nil {
		return fmt.Errorf("can't write status file %q: %w", statusFile, err)
	}

	return nil
}

type EnsureControlInterface interface {
	DaemonReload(ctx context.Context) error
	//EnableAndStartUnit(ctx context.Context, unitFile string) error
	EnableUnit(ctx context.Context, unitFile string) error
	StartUnit(ctx context.Context, unitFile string) error
	RestartUnit(ctx context.Context, unitFile string) error
	DisableAndStopUnit(ctx context.Context, unitFile string) error
	GetUnitActiveState(ctx context.Context, unitName string) (string, error)
}

// EnsureUnits will make sure to remove any unit that is no longer desired and create/update those that are.
func (m *UnitManager) EnsureUnits(ctx context.Context, nc *scyllav1alpha1.NodeConfig, recorder record.EventRecorder, requiredUnits []*NamedUnit, control EnsureControlInterface) error {
	status, err := m.ReadStatus()
	if err != nil {
		return fmt.Errorf("can't list managed units: %w", err)
	}

	klog.V(4).InfoS(
		"Checking if units need pruning",
		"Existing", len(status.ManagedUnits),
		"Desired", len(requiredUnits),
	)

	// Reload unit definitions for pruning in case that we didn't make it to daemon reload when writing them.
	err = control.DaemonReload(ctx)
	if err != nil {
		return fmt.Errorf("can't reload systemd: %w", err)
	}

	for _, existingUnitName := range status.ManagedUnits {
		isRequired := false
		for _, requiredUnit := range requiredUnits {
			if existingUnitName == requiredUnit.FileName {
				isRequired = true
				break
			}
		}
		if isRequired {
			continue
		}

		klog.V(2).InfoS("Disabling and stopping unit because it's no longer required", "Name", existingUnitName)
		err = control.DisableAndStopUnit(ctx, existingUnitName)
		if err != nil {
			if errors.Is(err, ErrNotExist) {
				klog.V(2).InfoS("Skipped disabling and stopping unit that doesn't exist", "Name", existingUnitName)
			} else {
				return fmt.Errorf("can't disable unit %q: %w", existingUnitName, err)
			}
		}

		klog.V(2).InfoS("Removing unit because it's no longer required", "Name", existingUnitName)
		existingUnitPath := m.GetUnitPath(existingUnitName)
		err = os.Remove(existingUnitPath)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("can't prune unit %q: %w", existingUnitName, err)
		}
		recorder.Eventf(
			nc,
			corev1.EventTypeNormal,
			"MountDeleted",
			"Mount unit %s has been deleted",
			existingUnitName,
		)
	}

	// TODO: do not try to reconcile any units not owned by us

	// First save the updated list of managed units first,
	// so we can clean up in the next run, if we were interrupted.
	status.ManagedUnits = slices.ConvertToSlice(func(unit *NamedUnit) string {
		return unit.FileName
	}, requiredUnits...)

	managedUnitsSet := sets.New(status.ManagedUnits...)
	for k := range status.ManagedUnitStatuses {
		if !managedUnitsSet.Has(k) {
			delete(status.ManagedUnitStatuses, k)
		}
	}

	err = m.WriteStatus(status)
	if err != nil {
		return fmt.Errorf("can't write status: %w", err)
	}

	for _, requiredUnit := range requiredUnits {
		klog.V(4).InfoS("Ensuring unit", "Name", requiredUnit.FileName)
		requiredUnitPath := m.GetUnitPath(requiredUnit.FileName)

		exists := true
		_, err := os.Stat(requiredUnitPath)
		if err != nil {
			if os.IsNotExist(err) {
				exists = false
			} else {
				return fmt.Errorf("can't stat unit %q: %w", requiredUnitPath, err)
			}
		}

		err = os.WriteFile(requiredUnitPath, requiredUnit.Data, 0666)
		if err != nil {
			return fmt.Errorf("can't write unit %q: %w", requiredUnitPath, err)
		}

		if !exists {
			klog.V(2).InfoS("Mount unit has been created", "Name", requiredUnit.FileName)
			recorder.Eventf(
				nc,
				corev1.EventTypeNormal,
				"MountCreated",
				"Mount unit %s has been created",
				requiredUnit.FileName,
			)
		}
	}

	// Reload unit definitions to enable and start them.
	err = control.DaemonReload(ctx)
	if err != nil {
		return fmt.Errorf("can't reload systemd: %w", err)
	}

	// TODO: aggregate errors
	for _, requiredUnit := range requiredUnits {
		klog.V(2).InfoS("Enabling unit", "Name", requiredUnit.FileName)
		err = control.EnableUnit(ctx, requiredUnit.FileName)
		if err != nil {
			return fmt.Errorf("can't enable unit %q: %w", requiredUnit.FileName, err)
		}

		var configurationHash string
		configurationHash, err = hash.HashObjects(requiredUnit.Data)
		if err != nil {
			return fmt.Errorf("can't calculate configuration hash for unit %q: %w", requiredUnit.FileName, err)
		}

		unitRequiresRestart := false
		mus := status.GetManagedUnitStatusOrDefault(requiredUnit.FileName)
		if mus.LastAppliedConfigurationHash != configurationHash {
			unitRequiresRestart = true
			klog.V(2).InfoS("Unit requires restart", "Name", requiredUnit.FileName)
		}

		var unitActiveState string
		unitActiveState, err = control.GetUnitActiveState(ctx, requiredUnit.FileName)
		if err != nil {
			return fmt.Errorf("can't get active state of unit %q: %w", requiredUnit.FileName, err)
		}

		switch unitActiveState {
		case "active":
			if !unitRequiresRestart {
				// TODO: remove this?
				klog.V(2).InfoS("Unit is already in a desired state", "Name", requiredUnit.FileName)
				break
			}

			klog.V(2).InfoS("Restarting active unit", "Name", requiredUnit.FileName)
			err = control.RestartUnit(ctx, requiredUnit.FileName)
			if err != nil {
				return fmt.Errorf("can't start unit %q: %w", requiredUnit.FileName, err)
			}

			// TODO: progressing
		case "inactive":
			klog.V(2).InfoS("Starting inactive unit", "Name", requiredUnit.FileName)
			err = control.StartUnit(ctx, requiredUnit.FileName)
			if err != nil {
				return fmt.Errorf("can't start unit %q: %w", requiredUnit.FileName, err)
			}

			// TODO: progressing
		case "activating", "deactivating":
			klog.V(2).InfoS("Unexpected state", "Name", requiredUnit.FileName, "ActiveState", unitActiveState)
			break
		case "failed":
			failedStateErr := fmt.Errorf("unit %q is in a failed state", requiredUnit.FileName)

			klog.V(2).InfoS("Restarting failed unit", "Name", requiredUnit.FileName)
			err = control.StartUnit(ctx, requiredUnit.FileName)
			if err != nil {
				return fmt.Errorf("can't start unit %q: %w", requiredUnit.FileName, err)
			}

			// TODO: progressing
		}

		mus.LastAppliedConfigurationHash = configurationHash
		status.ManagedUnitStatuses[requiredUnit.FileName] = mus
	}

	// Save managed units' statuses.
	err = m.WriteStatus(status)
	if err != nil {
		return fmt.Errorf("can't write status: %w", err)
	}

	return nil
}
