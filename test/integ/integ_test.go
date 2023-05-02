//go:build integration
// +build integration

package integ

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func currentContext(t *testing.T) string {
	val, err := k8s.RunKubectlAndGetOutputE(t, k8s.NewKubectlOptions("", "", ""), "config", "current-context")
	require.Nil(t, err)
	return val
}

func TestIntegrationAgentWithOperator(t *testing.T) {
	t.Log("Checking current context:", currentContext(t))
	if strings.Contains(currentContext(t), "staging") ||
		strings.Contains(currentContext(t), "prod") ||
		strings.Contains(currentContext(t), "dog") {
		t.Error("Make sure context is pointing to local cluster")
	}

	helmChartPath, err := filepath.Abs("../../charts/datadog-operator")
	require.NoError(t, err)

	namespaceName := fmt.Sprintf("datadog-agent-%s", strings.ToLower(random.UniqueId()))

	kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)

	k8s.CreateNamespace(t, kubectlOptions, namespaceName)
	defer k8s.DeleteNamespace(t, kubectlOptions, namespaceName)

	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		SetValues:      map[string]string{
			// "kube_version":                       "119",
		},
		ValuesFiles: []string{
			// "../charts/.common_lint_values.yaml",
		},
	}

	releaseName := fmt.Sprintf(
		"datadog-operator-%s",
		strings.ToLower(random.UniqueId()),
	)
	defer helm.Delete(t, options, releaseName, true)

	helm.Install(t, options, helmChartPath, releaseName)
	verifyNumPodsForSelector(t, kubectlOptions, 1, "app.kubernetes.io/name=datadog-operator")

	// Create secret from env
	k8s.RunKubectl(t, kubectlOptions, "create", "secret", "generic", "datadog-secret",
		"--from-literal",
		"api-key="+os.Getenv("LEVAN_M_TEST_API_KEY"),
		"--from-literal",
		"app-key="+os.Getenv("LEVAN_M_TEST_APP_KEY"))
	// defer k8s.KubectlDelete(t, kubectlOptions, )

	// Install DatadogAgent
	k8s.KubectlApply(t, kubectlOptions, "default.yaml")
	defer k8s.KubectlDelete(t, kubectlOptions, "default.yaml")

	verifyNumPodsForSelector(t, kubectlOptions, 2, "agent.datadoghq.com/component=agent")
	verifyNumPodsForSelector(t, kubectlOptions, 1, "agent.datadoghq.com/component=cluster-agent")
	verifyNumPodsForSelector(t, kubectlOptions, 2, "agent.datadoghq.com/component=cluster-checks-runner")
}

func verifyNumPodsForSelector(t *testing.T, kubectlOptions *k8s.KubectlOptions, numPods int, selector string) {
	t.Log("Waiting for number of pods created", "number", numPods, "selector", selector)
	k8s.WaitUntilNumPodsCreated(t, kubectlOptions, v1.ListOptions{
		LabelSelector: selector,
	}, numPods, 10, 5*time.Second)

	pods := k8s.ListPods(t, kubectlOptions, v1.ListOptions{
		LabelSelector: selector,
	})
	t.Log("Created pods", "number", len(pods), "selector", selector, "list", pods)
}
