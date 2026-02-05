package configassests

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/scylladb/scylla-operator/pkg/api/scylla/validation"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/util/images"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func validateRequired(v string) error {
	if len(strings.TrimSpace(v)) == 0 {
		return fmt.Errorf("value %q is empty", v)
	}

	return nil
}

// validateSemanticVersion checks if the provided string is a valid semantic version.
func validateSemanticVersion(v string) error {
	return validation.ValidateSemanticVersion(v, &field.Path{}).ToAggregate()
}

// validateMultiPlatformImage checks if the provided image is a multi-platform image.
func validateMultiPlatformImage(ctx context.Context) func(image string) error {
	return func(image string) error {
		return images.IsImageMultiPlatform(ctx, image)
	}
}

// validateMultiPlatformVersionWithRepo checks if the provided version is a valid multi-platform image for a given repository.
func validateMultiPlatformVersionWithRepo(ctx context.Context, repo string) func(tag string) error {
	return func(tag string) error {
		image := fmt.Sprintf("%s:%s", repo, tag)
		return validateMultiPlatformImage(ctx)(image)
	}
}

var (
	dashboardPathRegexFmt = `^[^ /]+/[^ /]+$`
	dashboardPathRegex    = regexp.MustCompile(dashboardPathRegexFmt)
)

func validateDashboardPath(p string) error {
	if dashboardPathRegex.MatchString(p) {
		return nil
	}

	return fmt.Errorf("path %q is invalid: doesn't match regex %q", p, dashboardPathRegexFmt)
}

func TestManagerAndAgentVersionsMatch(t *testing.T) {
	t.Parallel()

	// Extract tag from version string (ignore digest after @)
	extractTag := func(version string) string {
		parts := strings.Split(version, "@")
		return parts[0]
	}

	managerTag := extractTag(Project.Operator.ScyllaDBManagerVersion)
	agentTag := extractTag(Project.Operator.ScyllaDBManagerAgentVersion)

	if managerTag != agentTag {
		t.Errorf("scyllaDBManagerVersion tag %q does not match scyllaDBManagerAgentVersion tag %q", managerTag, agentTag)
	}
}

func TestProjectConfig(t *testing.T) {
	t.Parallel()

	composeValidators := func(validators ...func(string) error) func(string) error {
		return func(value string) error {
			var errs []error
			for _, validate := range validators {
				if err := validate(value); err != nil {
					errs = append(errs, err)
				}
			}
			return errors.Join(errs...)
		}
	}

	ctx := t.Context()

	testCases := []struct {
		name        string
		configField string
		testFn      func(string) error
	}{
		{
			name:        "scyllaDBVersion",
			configField: Project.Operator.ScyllaDBVersion,
			testFn: composeValidators(
				validateRequired,
				validateMultiPlatformVersionWithRepo(ctx, ScyllaDBImageRepository),
			),
		},
		{
			name:        "scyllaDBEnterpriseVersionNeedingConsistentClusterManagementOverride",
			configField: Project.Operator.ScyllaDBEnterpriseVersionNeedingConsistentClusterManagementOverride,
			testFn: composeValidators(
				validateRequired,
				validateMultiPlatformVersionWithRepo(ctx, ScyllaDBEnterpriseImageRepository),
			),
		},
		{
			name:        "scyllaDBUtilsImage",
			configField: Project.Operator.ScyllaDBUtilsImage,
			testFn: composeValidators(
				validateRequired,
				validateMultiPlatformImage(ctx),
			),
		},
		{
			name:        "scyllaDBManagerVersion",
			configField: Project.Operator.ScyllaDBManagerVersion,
			testFn: composeValidators(
				validateRequired,
				validateMultiPlatformVersionWithRepo(ctx, ScyllaDBManagerImageRepository),
			),
		},
		{
			name:        "scyllaDBManagerAgentVersion",
			configField: Project.Operator.ScyllaDBManagerAgentVersion,
			testFn: composeValidators(
				validateRequired,
				validateMultiPlatformVersionWithRepo(ctx, ScyllaDBManagerAgentImageRepository),
			),
		},
		{
			name:        "bashToolsImage",
			configField: Project.Operator.BashToolsImage,
			testFn: composeValidators(
				validateRequired,
				validateMultiPlatformImage(ctx),
			),
		},
		{
			name:        "grafanaImage",
			configField: Project.Operator.GrafanaImage,
			testFn: composeValidators(
				validateRequired,
				validateMultiPlatformImage(ctx),
			),
		},
		{
			name:        "grafanaDefaultPlatformDashboard",
			configField: Project.Operator.GrafanaDefaultPlatformDashboard,
			testFn: composeValidators(
				validateRequired,
				validateDashboardPath,
			),
		},
		{
			name:        "prometheusVersion",
			configField: Project.Operator.PrometheusVersion,
			testFn: composeValidators(
				validateRequired,
				validateSemanticVersion,
			),
		},
		{
			name:        "scyllaDBVersions.UpdateFrom",
			configField: Project.OperatorTests.ScyllaDBVersions.UpdateFrom,
			testFn: composeValidators(
				validateRequired,
				validateMultiPlatformVersionWithRepo(ctx, ScyllaDBImageRepository),
			),
		},
		{
			name:        "scyllaDBVersions.UpgradeFrom",
			configField: Project.OperatorTests.ScyllaDBVersions.UpgradeFrom,
			testFn: composeValidators(
				validateRequired,
				validateMultiPlatformVersionWithRepo(ctx, ScyllaDBImageRepository),
			),
		},
		{
			name:        "nodeSetupImage",
			configField: Project.OperatorTests.NodeSetupImage,
			testFn: composeValidators(
				validateRequired,
				validateMultiPlatformImage(ctx),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := tc.testFn(tc.configField); err != nil {
				t.Errorf("validation failed for %s: %v", tc.name, err)
			}
		})
	}
}

func TestScyllaMonitoringVersions(t *testing.T) {
	t.Parallel()

	// Parse versions.sh from scylla-monitoring submodule.
	versionsShPath := "../../submodules/github.com/scylladb/scylla-monitoring/versions.sh"
	file, err := os.Open(versionsShPath)
	if err != nil {
		t.Fatalf("failed to open versions.sh: %v", err)
	}
	defer file.Close()

	var prometheusVersion, grafanaVersion string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "PROMETHEUS_VERSION=") {
			prometheusVersion = strings.TrimPrefix(line, "PROMETHEUS_VERSION=")
		} else if strings.HasPrefix(line, "GRAFANA_VERSION=") {
			grafanaVersion = strings.Trim(strings.TrimPrefix(line, "GRAFANA_VERSION="), `"`)
		}
	}

	if err = scanner.Err(); err != nil {
		t.Fatalf("failed to read versions.sh: %v", err)
	}

	t.Run("prometheusVersion matches PROMETHEUS_VERSION", func(t *testing.T) {
		if Project.Operator.PrometheusVersion != prometheusVersion {
			t.Errorf("prometheusVersion %q does not match PROMETHEUS_VERSION in versions.sh %q", Project.Operator.PrometheusVersion, prometheusVersion)
		}
	})

	t.Run("grafanaImage tag matches GRAFANA_VERSION", func(t *testing.T) {
		grafanaImageTag, err := naming.ImageToVersion(Project.Operator.GrafanaImage)
		if err != nil {
			t.Fatalf("failed to extract tag from grafanaImage: %v", err)
		}

		if grafanaImageTag != grafanaVersion {
			t.Errorf("grafanaImage tag %q does not match GRAFANA_VERSION in versions.sh %q", grafanaImageTag, grafanaVersion)
		}
	})

	// Verify that grafanaDefaultPlatformDashboard matches scyllaDBVersion (up to minor).
	t.Run("grafanaDefaultPlatformDashboard matches scyllaDBVersion", func(t *testing.T) {
		scyllaDBVersion, err := semver.Parse(Project.Operator.ScyllaDBVersion)
		if err != nil {
			t.Fatalf("failed to parse scyllaDBVersion: %v", err)
		}

		expectedMajorMinor := fmt.Sprintf("%d.%d", scyllaDBVersion.Major, scyllaDBVersion.Minor)
		expectedDefaultPlatformDashboard := fmt.Sprintf("scylladb-%s/scylla-overview.%s.json", expectedMajorMinor, expectedMajorMinor)

		if Project.Operator.GrafanaDefaultPlatformDashboard != expectedDefaultPlatformDashboard {
			t.Errorf("grafanaDefaultPlatformDashboard %q does not match expected %q based on scyllaDBVersion %q", Project.Operator.GrafanaDefaultPlatformDashboard, expectedDefaultPlatformDashboard, Project.Operator.ScyllaDBVersion)
		}
	})

	t.Run("grafanaDefaultPlatformDashboard exists", func(t *testing.T) {
		dashboardPath := "../../assets/monitoring/grafana/v1alpha1/dashboards/platform/" + Project.Operator.GrafanaDefaultPlatformDashboard
		if _, err := os.Stat(dashboardPath); os.IsNotExist(err) {
			t.Errorf("grafanaDefaultPlatformDashboard %q does not exist at path %q", Project.Operator.GrafanaDefaultPlatformDashboard, dashboardPath)
		}
	})
}
