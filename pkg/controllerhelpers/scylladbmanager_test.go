// Copyright (C) 2025 ScyllaDB

package controllerhelpers

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_getScyllaDBManagerAgentAuthTokenConfig(t *testing.T) {
	t.Parallel()

	const mockAuthToken = "mock-auth-token"
	newMockAuthToken := func() string {
		return mockAuthToken
	}

	getNilSecret := func() ([]metav1.Condition, *corev1.Secret, error) {
		return nil, nil, nil
	}

	newDefaultOptions := func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
		return GetScyllaDBManagerAgentAuthTokenConfigOptions{
			GetOptionalCustomAgentConfigSecret: getNilSecret,
			GetOptionalExistingAuthTokenSecret: getNilSecret,
			ContinueOnCustomAgentConfigError:   false,
		}
	}

	tt := []struct {
		name                                               string
		options                                            GetScyllaDBManagerAgentAuthTokenConfigOptions
		expectedProgressingConditions                      []metav1.Condition
		expected                                           []byte
		expectedErr                                        error
		expectedErrAsScyllaDBManagerAgentCustomConfigError bool
	}{
		{
			name:                          "no custom agent config, no existing auth token",
			options:                       newDefaultOptions(),
			expectedProgressingConditions: nil,
			expected:                      []byte(`auth_token: ` + mockAuthToken + "\n"),
			expectedErr:                   nil,
			expectedErrAsScyllaDBManagerAgentCustomConfigError: false,
		},
		{
			name: "custom agent config, no existing auth token",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalCustomAgentConfigSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "custom-agent-config",
							Namespace: "test",
						},
						Data: map[string][]byte{
							"scylla-manager-agent.yaml": []byte("auth_token: custom-auth-token\n"),
						},
					}, nil
				}

				return opts
			}(),
			expectedProgressingConditions: nil,
			expected:                      []byte(`auth_token: custom-auth-token` + "\n"),
			expectedErr:                   nil,
			expectedErrAsScyllaDBManagerAgentCustomConfigError: false,
		},
		{
			name: "progressing conditions from custom agent config func, no existing auth token",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalCustomAgentConfigSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return []metav1.Condition{
						{
							Type:               "SecretControllerProgressing",
							Status:             metav1.ConditionTrue,
							Reason:             "WaitingForSecret",
							Message:            `Waiting for Secret "test/custom-agent-config" to exist.`,
							ObservedGeneration: 1,
						},
					}, nil, nil
				}

				return opts
			}(),
			expectedProgressingConditions: []metav1.Condition{
				{
					Type:               "SecretControllerProgressing",
					Status:             metav1.ConditionTrue,
					Reason:             "WaitingForSecret",
					Message:            `Waiting for Secret "test/custom-agent-config" to exist.`,
					ObservedGeneration: 1,
				},
			},
			expected:    nil,
			expectedErr: nil,
			expectedErrAsScyllaDBManagerAgentCustomConfigError: false,
		},
		{
			name: "error from custom agent config func, no existing auth token, no continue on custom agent config error",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalCustomAgentConfigSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, nil, errors.New("can't get custom agent config secret")
				}

				return opts
			}(),
			expectedProgressingConditions: nil,
			expected:                      nil,
			expectedErr:                   fmt.Errorf("can't get ScyllaDB Manager agent auth token: %w", NewScyllaDBManagerAgentCustomConfigError(fmt.Errorf("can't get ScyllaDB Manager agent auth token from custom config secret: %w", fmt.Errorf("can't get ScyllaDB Manager agent auth token secret: %w", errors.New("can't get custom agent config secret"))))),
			expectedErrAsScyllaDBManagerAgentCustomConfigError: true,
		},
		{
			name: "error from custom agent config func, no existing auth token, continue on custom agent config error",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalCustomAgentConfigSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, nil, errors.New("can't get custom agent config secret")
				}
				opts.ContinueOnCustomAgentConfigError = true

				return opts
			}(),
			expectedProgressingConditions: nil,
			expected:                      []byte(`auth_token: ` + mockAuthToken + "\n"),
			expectedErr:                   fmt.Errorf("can't get ScyllaDB Manager agent auth token: %w", NewScyllaDBManagerAgentCustomConfigError(fmt.Errorf("can't get ScyllaDB Manager agent auth token from custom config secret: %w", fmt.Errorf("can't get ScyllaDB Manager agent auth token secret: %w", errors.New("can't get custom agent config secret"))))),
			expectedErrAsScyllaDBManagerAgentCustomConfigError: true,
		},
		{
			name: "error from custom agent config extraction, no existing auth token, no continue on custom agent config error",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalCustomAgentConfigSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "custom-agent-config",
							Namespace: "test",
						},
						Data: map[string][]byte{
							// scylla-manager-agent.yaml is missing.
						},
					}, nil
				}

				return opts
			}(),
			expectedProgressingConditions: nil,
			expected:                      nil,
			expectedErr:                   fmt.Errorf("can't get ScyllaDB Manager agent auth token: %w", NewScyllaDBManagerAgentCustomConfigError(fmt.Errorf("can't get ScyllaDB Manager agent auth token from custom config secret: %w", fmt.Errorf("can't extract ScyllaDB Manager agent auth token from Secret \"test/custom-agent-config\": %w", errors.New("secret \"test/custom-agent-config\" is missing \"scylla-manager-agent.yaml\" data"))))),
			expectedErrAsScyllaDBManagerAgentCustomConfigError: true,
		},
		{
			name: "error from custom agent config extraction, no existing auth token, continue on custom agent config error",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalCustomAgentConfigSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "custom-agent-config",
							Namespace: "test",
						},
						Data: map[string][]byte{
							// scylla-manager-agent.yaml is missing.
						},
					}, nil
				}
				opts.ContinueOnCustomAgentConfigError = true

				return opts
			}(),
			expectedProgressingConditions: nil,
			expected:                      []byte(`auth_token: ` + mockAuthToken + "\n"),
			expectedErr:                   fmt.Errorf("can't get ScyllaDB Manager agent auth token: %w", NewScyllaDBManagerAgentCustomConfigError(fmt.Errorf("can't get ScyllaDB Manager agent auth token from custom config secret: %w", fmt.Errorf("can't extract ScyllaDB Manager agent auth token from Secret \"test/custom-agent-config\": %w", errors.New("secret \"test/custom-agent-config\" is missing \"scylla-manager-agent.yaml\" data"))))),
			expectedErrAsScyllaDBManagerAgentCustomConfigError: true,
		},
		{
			name: "no custom agent config, existing auth token",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalExistingAuthTokenSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "existing-auth-token",
							Namespace: "test",
						},
						Data: map[string][]byte{
							"auth-token.yaml": []byte("auth_token: existing-auth-token\n"),
						},
					}, nil
				}

				return opts
			}(),
			expectedProgressingConditions: nil,
			expected:                      []byte(`auth_token: existing-auth-token` + "\n"),
			expectedErr:                   nil,
			expectedErrAsScyllaDBManagerAgentCustomConfigError: false,
		},
		{
			name: "no custom agent config, progressing conditions from existing auth token func",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalExistingAuthTokenSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return []metav1.Condition{
						{
							Type:               "SecretControllerProgressing",
							Status:             metav1.ConditionTrue,
							Reason:             "WaitingForSecret",
							Message:            `Waiting for Secret "test/existing-auth-token" to exist.`,
							ObservedGeneration: 1,
						},
					}, nil, nil
				}

				return opts
			}(),
			expectedProgressingConditions: []metav1.Condition{
				{
					Type:               "SecretControllerProgressing",
					Status:             metav1.ConditionTrue,
					Reason:             "WaitingForSecret",
					Message:            `Waiting for Secret "test/existing-auth-token" to exist.`,
					ObservedGeneration: 1,
				},
			},
			expected:    nil,
			expectedErr: nil,
			expectedErrAsScyllaDBManagerAgentCustomConfigError: false,
		},
		{
			name: "no custom agent config, error from existing auth token func",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalExistingAuthTokenSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, nil, errors.New("can't get existing auth token secret")
				}

				return opts
			}(),
			expectedProgressingConditions: nil,
			expected:                      nil,
			expectedErr:                   fmt.Errorf("can't get ScyllaDB Manager agent auth token: %w", fmt.Errorf("can't get ScyllaDB Manager agent auth token from existing secret: %w", fmt.Errorf("can't get ScyllaDB Manager agent auth token secret: %w", errors.New("can't get existing auth token secret")))),
			expectedErrAsScyllaDBManagerAgentCustomConfigError: false,
		},
		{
			name: "no custom agent config, error from existing auth token extraction",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalExistingAuthTokenSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "existing-auth-token",
							Namespace: "test",
						},
						Data: map[string][]byte{
							// auth-token.yaml is missing.
						},
					}, nil
				}

				return opts
			}(),
			expectedProgressingConditions: nil,
			expected:                      nil,
			expectedErr:                   fmt.Errorf("can't get ScyllaDB Manager agent auth token: %w", fmt.Errorf("can't get ScyllaDB Manager agent auth token from existing secret: %w", fmt.Errorf("can't extract ScyllaDB Manager agent auth token from Secret \"test/existing-auth-token\": %w", errors.New("secret \"test/existing-auth-token\" is missing \"auth-token.yaml\" data")))),
			expectedErrAsScyllaDBManagerAgentCustomConfigError: false,
		},
		{
			name: "custom agent config, existing auth token",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalCustomAgentConfigSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "custom-agent-config",
							Namespace: "test",
						},
						Data: map[string][]byte{
							"scylla-manager-agent.yaml": []byte("auth_token: custom-auth-token\n"),
						},
					}, nil
				}
				opts.GetOptionalExistingAuthTokenSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "existing-auth-token",
							Namespace: "test",
						},
						Data: map[string][]byte{
							"auth-token.yaml": []byte("auth_token: existing-auth-token\n"),
						},
					}, nil
				}

				return opts
			}(),
			expectedProgressingConditions: nil,
			expected:                      []byte(`auth_token: custom-auth-token` + "\n"),
			expectedErr:                   nil,
			expectedErrAsScyllaDBManagerAgentCustomConfigError: false,
		},
		{
			name: "progressing conditions from custom agent config func, existing auth token",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalCustomAgentConfigSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return []metav1.Condition{
						{
							Type:               "SecretControllerProgressing",
							Status:             metav1.ConditionTrue,
							Reason:             "WaitingForSecret",
							Message:            `Waiting for Secret "test/custom-agent-config" to exist.`,
							ObservedGeneration: 1,
						},
					}, nil, nil
				}
				opts.GetOptionalExistingAuthTokenSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "existing-auth-token",
							Namespace: "test",
						},
						Data: map[string][]byte{
							"auth-token.yaml": []byte("auth_token: existing-auth-token\n"),
						},
					}, nil
				}

				return opts
			}(),
			expectedProgressingConditions: []metav1.Condition{
				{
					Type:               "SecretControllerProgressing",
					Status:             metav1.ConditionTrue,
					Reason:             "WaitingForSecret",
					Message:            `Waiting for Secret "test/custom-agent-config" to exist.`,
					ObservedGeneration: 1,
				},
			},
			expected:    nil,
			expectedErr: nil,
			expectedErrAsScyllaDBManagerAgentCustomConfigError: false,
		},
		{
			name: "error from custom agent config func, existing auth token, no continue on custom agent config error",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalCustomAgentConfigSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, nil, errors.New("can't get custom agent config secret")
				}
				opts.GetOptionalExistingAuthTokenSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "existing-auth-token",
							Namespace: "test",
						},
						Data: map[string][]byte{
							"auth-token.yaml": []byte("auth_token: existing-auth-token\n"),
						},
					}, nil
				}

				return opts
			}(),
			expectedProgressingConditions: nil,
			expected:                      nil,
			expectedErr:                   fmt.Errorf("can't get ScyllaDB Manager agent auth token: %w", NewScyllaDBManagerAgentCustomConfigError(fmt.Errorf("can't get ScyllaDB Manager agent auth token from custom config secret: %w", fmt.Errorf("can't get ScyllaDB Manager agent auth token secret: %w", errors.New("can't get custom agent config secret"))))),
			expectedErrAsScyllaDBManagerAgentCustomConfigError: true,
		},
		{
			name: "error from custom agent config func, existing auth token, continue on custom agent config error",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalCustomAgentConfigSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, nil, errors.New("can't get custom agent config secret")
				}
				opts.GetOptionalExistingAuthTokenSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "existing-auth-token",
							Namespace: "test",
						},
						Data: map[string][]byte{
							"auth-token.yaml": []byte("auth_token: existing-auth-token\n"),
						},
					}, nil
				}
				opts.ContinueOnCustomAgentConfigError = true

				return opts
			}(),
			expectedProgressingConditions: nil,
			expected:                      []byte(`auth_token: existing-auth-token` + "\n"),
			expectedErr:                   fmt.Errorf("can't get ScyllaDB Manager agent auth token: %w", NewScyllaDBManagerAgentCustomConfigError(fmt.Errorf("can't get ScyllaDB Manager agent auth token from custom config secret: %w", fmt.Errorf("can't get ScyllaDB Manager agent auth token secret: %w", errors.New("can't get custom agent config secret"))))),
			expectedErrAsScyllaDBManagerAgentCustomConfigError: true,
		},
		{
			name: "error from custom agent config func, progressing conditions from existing auth token func, continue on custom agent config error",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalCustomAgentConfigSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, nil, errors.New("can't get custom agent config secret")
				}
				opts.GetOptionalExistingAuthTokenSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return []metav1.Condition{
						{
							Type:               "SecretControllerProgressing",
							Status:             metav1.ConditionTrue,
							Reason:             "WaitingForSecret",
							Message:            `Waiting for Secret "test/existing-auth-token" to exist.`,
							ObservedGeneration: 1,
						},
					}, nil, nil
				}
				opts.ContinueOnCustomAgentConfigError = true

				return opts
			}(),
			expectedProgressingConditions: []metav1.Condition{
				{
					Type:               "SecretControllerProgressing",
					Status:             metav1.ConditionTrue,
					Reason:             "WaitingForSecret",
					Message:            `Waiting for Secret "test/existing-auth-token" to exist.`,
					ObservedGeneration: 1,
				},
			},
			expected:    nil,
			expectedErr: fmt.Errorf("can't get ScyllaDB Manager agent auth token: %w", NewScyllaDBManagerAgentCustomConfigError(fmt.Errorf("can't get ScyllaDB Manager agent auth token from custom config secret: %w", fmt.Errorf("can't get ScyllaDB Manager agent auth token secret: %w", errors.New("can't get custom agent config secret"))))),
			expectedErrAsScyllaDBManagerAgentCustomConfigError: true,
		},
		{
			name: "error from custom agent config func, error from existing auth token func, continue on custom agent config error",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalCustomAgentConfigSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, nil, errors.New("can't get custom agent config secret")
				}
				opts.GetOptionalExistingAuthTokenSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, nil, errors.New("can't get existing auth token secret")
				}
				opts.ContinueOnCustomAgentConfigError = true

				return opts
			}(),
			expectedProgressingConditions: nil,
			expected:                      nil,
			expectedErr:                   fmt.Errorf("can't get ScyllaDB Manager agent auth token: %w", fmt.Errorf("can't get ScyllaDB Manager agent auth token from existing secret: %w", fmt.Errorf("can't get ScyllaDB Manager agent auth token secret: %w", errors.New("can't get existing auth token secret")))),
			// This error is not a ScyllaDBManagerAgentCustomConfigError because a subsequent error occurred.
			expectedErrAsScyllaDBManagerAgentCustomConfigError: false,
		},
		{
			name: "error from custom agent config func, error from existing auth token func, continue on custom agent config error",
			options: func() GetScyllaDBManagerAgentAuthTokenConfigOptions {
				opts := newDefaultOptions()

				opts.GetOptionalCustomAgentConfigSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, nil, errors.New("can't get custom agent config secret")
				}
				opts.GetOptionalExistingAuthTokenSecret = func() ([]metav1.Condition, *corev1.Secret, error) {
					return nil, &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "existing-auth-token",
							Namespace: "test",
						},
						Data: map[string][]byte{
							// auth-token.yaml is missing.
						},
					}, nil
				}
				opts.ContinueOnCustomAgentConfigError = true

				return opts
			}(),
			expectedProgressingConditions: nil,
			expected:                      nil,
			expectedErr:                   fmt.Errorf("can't get ScyllaDB Manager agent auth token: %w", fmt.Errorf("can't get ScyllaDB Manager agent auth token from existing secret: %w", fmt.Errorf("can't extract ScyllaDB Manager agent auth token from Secret \"test/existing-auth-token\": %w", errors.New("secret \"test/existing-auth-token\" is missing \"auth-token.yaml\" data")))),
			// This error is not a ScyllaDBManagerAgentCustomConfigError because a subsequent error occurred.
			expectedErrAsScyllaDBManagerAgentCustomConfigError: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			progressingConditions, got, err := getScyllaDBManagerAgentAuthTokenConfig(
				newMockAuthToken,
				tc.options,
			)

			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s\n", cmp.Diff(tc.expectedErr, err, cmpopts.EquateErrors()))
			}

			gotErrAsScyllaDBManagerAgentCustomConfigError := errors.As(err, &ScyllaDBManagerAgentCustomConfigError{})
			if gotErrAsScyllaDBManagerAgentCustomConfigError != tc.expectedErrAsScyllaDBManagerAgentCustomConfigError {
				t.Errorf("expected error as ScyllaDBManagerAgentCustomConfigError to evaluate to: %t, got: %t", tc.expectedErrAsScyllaDBManagerAgentCustomConfigError, gotErrAsScyllaDBManagerAgentCustomConfigError)
			}

			if !equality.Semantic.DeepEqual(progressingConditions, tc.expectedProgressingConditions) {
				t.Errorf("expected and got progressing conditions differ:\n%s\n", cmp.Diff(tc.expectedProgressingConditions, progressingConditions))
			}

			if !bytes.Equal(got, tc.expected) {
				t.Errorf("expected auth token config: %q, got %q", tc.expected, got)
			}
		})
	}
}
