// Copyright (c) 2024 ScyllaDB.

package operator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/scylladb/scylla-operator/pkg/genericclioptions"
	"github.com/scylladb/scylla-operator/pkg/signals"
	"github.com/scylladb/scylla-operator/pkg/version"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	apimachineryutilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

type RlimitsJobOptions struct {
	PID int

	procfsPath string
}

func NewRlimitsJobOptions(streams genericclioptions.IOStreams) *RlimitsJobOptions {
	return &RlimitsJobOptions{
		procfsPath: "/proc",
	}
}

func NewRlimitsJobCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewRlimitsJobOptions(streams)

	cmd := &cobra.Command{
		Use:   "rlimits-job",
		Short: "Changes rlimits of process.",
		Long:  "Changes rlimits of process.",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate()
			if err != nil {
				return err
			}

			err = o.Complete()
			if err != nil {
				return err
			}

			err = o.Run(streams, cmd)
			if err != nil {
				return err
			}

			return nil
		},

		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().IntVarP(&o.PID, "pid", "", o.PID, "PID of the target process for which rlimit should be changed")

	return cmd
}

func (o *RlimitsJobOptions) Validate() error {
	var errs []error

	if o.PID == 0 {
		errs = append(errs, fmt.Errorf("pid cannot be zero"))
	}

	return apimachineryutilerrors.NewAggregate(errs)
}

func (o *RlimitsJobOptions) Complete() error {
	return nil
}

func (o *RlimitsJobOptions) Run(streams genericclioptions.IOStreams, cmd *cobra.Command) error {
	klog.InfoS("Starting rlimits Job", "version", version.Get())

	defer func(startTime time.Time) {
		klog.InfoS("Rlimits Job completed", "duration", time.Since(startTime))
	}(time.Now())

	cliflag.PrintFlags(cmd.Flags())

	stopCh := signals.StopChannel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-stopCh
		cancel()
	}()

	err := o.changeRlimits(ctx)
	if err != nil {
		return fmt.Errorf("can't change rlimits: %w", err)
	} else {
		klog.InfoS("Rlimits were changed successfully")
	}

	return nil
}

func (o *RlimitsJobOptions) changeRlimits(ctx context.Context) error {
	klog.InfoS("Checking maximum possible nofile limit")

	maxNofileLimit, err := getSysctlFSNROpen(o.procfsPath)
	if err != nil {
		return fmt.Errorf("can't check maximum possible nofile limit: %w", err)
	}

	childPIDs, err := getChildPIDs(o.procfsPath, o.PID)
	if err != nil {
		return fmt.Errorf("can't get child PIDs of process %d: %w", o.PID, err)
	}

	for _, pid := range childPIDs {
		klog.InfoS("Changing process nofile rlimits", "PID", pid, "nofile", maxNofileLimit)

		err = unix.Prlimit(o.PID, unix.RLIMIT_NOFILE, &unix.Rlimit{
			// Soft limit
			Cur: maxNofileLimit,
			// Hard limit
			Max: maxNofileLimit,
		}, nil)

		if err != nil {
			return fmt.Errorf("can't set nofile rlimit of %d process: %w", pid, err)
		}
	}

	return nil
}

func getSysctlFSNROpen(procfsPath string) (uint64, error) {
	fsNROpenPath := filepath.Join(procfsPath, "/sys/fs/nr_open")

	fsNROpenStr, err := os.ReadFile(fsNROpenPath)
	if err != nil {
		return 0, fmt.Errorf("can't read %q: %w", fsNROpenPath, err)
	}

	maxNofileLimit, err := strconv.ParseUint(strings.TrimSpace(string(fsNROpenStr)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("can't parse %q to uint64: %w", fsNROpenStr, err)
	}

	return maxNofileLimit, nil
}

// getChildPIDs recursively finds all PIDs in the process tree starting from pid.
// It reads /proc/<pid>/task/<tid>/children for each task (thread) to find children.
func getChildPIDs(procfsPath string, pid int) ([]int, error) {
	visited := sets.New[int]()

	var curr int
	queue := []int{pid}
	for len(queue) > 0 {
		curr = queue[0]
		queue = queue[1:]

		if visited.Has(curr) {
			continue
		}
		visited.Insert(curr)

		taskDirPath := fmt.Sprintf("%s/%d/task", procfsPath, curr)
		tasks, err := os.ReadDir(taskDirPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("can't read dir %q: %w", taskDirPath, err)
			}

			// Task likely ended, ignore.
			continue
		}

		for _, task := range tasks {
			childrenFilePath := fmt.Sprintf("%s/%s/children", taskDirPath, task.Name())
			children, err := os.ReadFile(childrenFilePath)
			if err != nil {
				if !os.IsNotExist(err) {
					return nil, fmt.Errorf("can't open file %q: %w", childrenFilePath, err)
				}

				// TODO: comment
				continue
			}

			childPIDs := strings.Fields(string(children))
			for _, childPID := range childPIDs {
				convertedChildPID, err := strconv.Atoi(childPID)
				if err != nil {
					return nil, fmt.Errorf("can't convert child PID %q to int: %w", childPID, err)
				}

				queue = append(queue, convertedChildPID)
			}
		}
	}

	return visited.UnsortedList(), nil
}
