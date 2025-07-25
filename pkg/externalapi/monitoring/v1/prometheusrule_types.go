// Copyright 2018 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apimachineryutilintstr "k8s.io/apimachinery/pkg/util/intstr"
)

const (
	PrometheusRuleKind    = "PrometheusRule"
	PrometheusRuleName    = "prometheusrules"
	PrometheusRuleKindKey = "prometheusrule"
)

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="promrule"

// The `PrometheusRule` custom resource definition (CRD) defines [alerting](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/) and [recording](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/) rules to be evaluated by `Prometheus` or `ThanosRuler` objects.
//
// `Prometheus` and `ThanosRuler` objects select `PrometheusRule` objects using label and namespace selectors.
type PrometheusRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of desired alerting rule definitions for Prometheus.
	Spec PrometheusRuleSpec `json:"spec"`
}

// DeepCopyObject implements the runtime.Object interface.
func (f *PrometheusRule) DeepCopyObject() runtime.Object {
	return f.DeepCopy()
}

// PrometheusRuleSpec contains specification parameters for a Rule.
// +k8s:openapi-gen=true
type PrometheusRuleSpec struct {
	// Content of Prometheus rule file
	// +listType=map
	// +listMapKey=name
	Groups []RuleGroup `json:"groups,omitempty"`
}

// RuleGroup and Rule are copied instead of vendored because the
// upstream Prometheus struct definitions don't have json struct tags.

// RuleGroup is a list of sequentially evaluated recording and alerting rules.
// +k8s:openapi-gen=true
type RuleGroup struct {
	// Name of the rule group.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// Labels to add or overwrite before storing the result for its rules.
	// The labels defined at the rule level take precedence.
	//
	// It requires Prometheus >= 3.0.0.
	// The field is ignored for Thanos Ruler.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Interval determines how often rules in the group are evaluated.
	// +optional
	Interval *Duration `json:"interval,omitempty"`
	// Defines the offset the rule evaluation timestamp of this particular group by the specified duration into the past.
	//
	// It requires Prometheus >= v2.53.0.
	// It is not supported for ThanosRuler.
	// +optional
	QueryOffset *Duration `json:"query_offset,omitempty"`
	// List of alerting and recording rules.
	// +optional
	Rules []Rule `json:"rules,omitempty"`
	// PartialResponseStrategy is only used by ThanosRuler and will
	// be ignored by Prometheus instances.
	// More info: https://github.com/thanos-io/thanos/blob/main/docs/components/rule.md#partial-response
	// +kubebuilder:validation:Pattern="^(?i)(abort|warn)?$"
	PartialResponseStrategy string `json:"partial_response_strategy,omitempty"`
	// Limit the number of alerts an alerting rule and series a recording
	// rule can produce.
	// Limit is supported starting with Prometheus >= 2.31 and Thanos Ruler >= 0.24.
	// +optional
	Limit *int `json:"limit,omitempty"`
}

// Rule describes an alerting or recording rule
// See Prometheus documentation: [alerting](https://www.prometheus.io/docs/prometheus/latest/configuration/alerting_rules/) or [recording](https://www.prometheus.io/docs/prometheus/latest/configuration/recording_rules/#recording-rules) rule
// +k8s:openapi-gen=true
type Rule struct {
	// Name of the time series to output to. Must be a valid metric name.
	// Only one of `record` and `alert` must be set.
	Record string `json:"record,omitempty"`
	// Name of the alert. Must be a valid label value.
	// Only one of `record` and `alert` must be set.
	Alert string `json:"alert,omitempty"`
	// PromQL expression to evaluate.
	Expr apimachineryutilintstr.IntOrString `json:"expr"`
	// Alerts are considered firing once they have been returned for this long.
	// +optional
	For *Duration `json:"for,omitempty"`
	// KeepFiringFor defines how long an alert will continue firing after the condition that triggered it has cleared.
	// +optional
	KeepFiringFor *NonEmptyDuration `json:"keep_firing_for,omitempty"`
	// Labels to add or overwrite.
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations to add to each alert.
	// Only valid for alerting rules.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// PrometheusRuleList is a list of PrometheusRules.
// +k8s:openapi-gen=true
type PrometheusRuleList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of Rules
	Items []PrometheusRule `json:"items"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *PrometheusRuleList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}
