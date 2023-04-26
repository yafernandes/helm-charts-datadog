package datadog_operator

import (
	"testing"

	"github.com/DataDog/datadog-operator/apis/datadoghq/v2alpha1"
	"github.com/DataDog/helm-chart/test/common"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
)

func Test_baseline_manifests(t *testing.T) {
	tests := []struct {
		name                 string
		command              common.HelmCommand
		baselineManifestPath string
		assertions           func(t *testing.T, baselineManifestPath, manifest string)
		skipTest             bool
	}{
		{
			name: "Operator Deployment default",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      "../../charts/datadog-operator/values.yaml",
				Overrides:   map[string]string{},
			},
			baselineManifestPath: "./baseline/default_Operator_Deployment.yaml",
			assertions:           verifyOperatorDeployment,
			skipTest:             SkipTest,
		},
		{
			name: "DatadogAgent CRD default",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				// datadogCRDs is an alias defined in the chart dependency
				ShowOnly:  []string{"charts/datadogCRDs/templates/datadoghq.com_datadogagents_v1.yaml"},
				Values:    "../../charts/datadog-operator/values.yaml",
				Overrides: map[string]string{},
			},
			baselineManifestPath: "./baseline/default_DatadogAgent.yaml",
			assertions:           verifyDatadogAgent,
			skipTest:             SkipTest,
		},
		{
			name: "Operator Deployment with cert manager enabled",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      "../../charts/datadog-operator/values.yaml",
				Overrides: map[string]string{
					"datadogCRDs.migration.datadogAgents.useCertManager":            "true",
					"datadogCRDs.migration.datadogAgents.conversionWebhook.enabled": "true",
				},
			},
			baselineManifestPath: "./baseline/Operator_Deployment_with_certManager.yaml",
			assertions:           verifyOperatorDeployment,
			skipTest:             SkipTest,
		},
		{
			name: "DatadogAgent CRD with cert manager enabled",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				// datadogCRDs is an alias defined in the chart dependency
				ShowOnly: []string{"charts/datadogCRDs/templates/datadoghq.com_datadogagents_v1.yaml"},
				Values:   "../../charts/datadog-operator/values.yaml",
				Overrides: map[string]string{
					"datadogCRDs.migration.datadogAgents.useCertManager":            "true",
					"datadogCRDs.migration.datadogAgents.conversionWebhook.enabled": "true",
				},
			},
			baselineManifestPath: "./baseline/DatadogAgent_with_certManager.yaml",
			assertions:           verifyDatadogAgent,
			skipTest:             SkipTest,
		},
	}

	for _, tt := range tests {
		if tt.skipTest {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			manifest, err1 := common.RenderChart(t, tt.command)
			// fmt.Println("manifest", manifest)
			assert.Nil(t, err1, "cound't render template")
			tt.assertions(t, tt.baselineManifestPath, manifest)
		})
	}
}

func verifyOperatorDeployment(t *testing.T, baselineManifestPath, manifest string) {
	verifyBaseline(t, baselineManifestPath, manifest, appsv1.Deployment{}, appsv1.Deployment{})
}

func verifyDatadogAgent(t *testing.T, baselineManifestPath, manifest string) {
	verifyBaseline(t, baselineManifestPath, manifest, v2alpha1.DatadogAgent{}, v2alpha1.DatadogAgent{})
}

func verifyBaseline[T any](t *testing.T, baselineManifestPath, manifest string, baseline, actual T) {
	common.Unmarshal(t, manifest, &actual)
	common.LoadFromFile(t, baselineManifestPath, &baseline)
	assert.True(t, cmp.Equal(baseline, actual), cmp.Diff(baseline, actual))
}
