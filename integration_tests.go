package terratest_polkadot_deployer

import (
    "crypto/tls"
    "fmt"
    "path/filepath"
    "testing"
    "time"
    "strings"

    "github.com/gruntwork-io/terratest/modules/helm"
    "github.com/gruntwork-io/terratest/modules/k8s"
    "github.com/gruntwork-io/terratest/modules/random"
    "github.com/gruntwork-io/terratest/modules/test-structure"
    "github.com/gruntwork-io/terratest/modules/http-helper"

    "github.com/gruntwork-io/terratest/modules/terraform"

    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/assert"
)

// Test that the Node count matches the Terraform specification
func TestNodeCount(t *testing.T, terraformDir string) {
    kubectlOptions := test_structure.LoadKubectlOptions(t, terraformDir)
    nodeCount := test_structure.LoadInt(t, terraformDir, "nodeCount")

    k8s.WaitUntilAllNodesReady(t, kubectlOptions, 40, 10*time.Second)
    nodes := k8s.GetNodes(t, kubectlOptions)
    assert.Equal(t, len(nodes), nodeCount)
}

// Test service deployment and verify it's availability on configured nodePort
func TestServiceAvailability(t *testing.T, terraformDir string) {
    kubectlOptions := test_structure.LoadKubectlOptions(t, terraformDir)
    nodePort := test_structure.LoadInt(t, terraformDir, "nodePort")

    // Deploy Helm chart
    var helmValues map[string]string
    helmValuesFile := test_structure.LoadString(t, terraformDir, "helmValuesFile")
    test_structure.LoadTestData(t, helmValuesFile, &helmValues)

    helmOptions := &helm.Options{
        KubectlOptions: kubectlOptions,
        SetValues: helmValues,
    }

    helmChartPath, err := filepath.Abs("fixtures/nginx-chart")
    require.NoError(t, err)

    helmReleaseName := fmt.Sprintf("polkadot-nginx-%s", strings.ToLower(random.UniqueId()))
    defer helm.Delete(t, helmOptions, helmReleaseName, true)
    helm.Install(t, helmOptions, helmChartPath, helmReleaseName)

    // Validate service availability
    k8s.WaitUntilServiceAvailable(t, kubectlOptions, helmReleaseName, 60, 5*time.Second)
    service := k8s.GetService(t, kubectlOptions, helmReleaseName)
    require.Equal(t, service.Name, helmReleaseName)

    // Check external connectivity
    tlsConfig := tls.Config{}
    url := fmt.Sprintf("http://%s", GetPublicServiceEndpoint(t, kubectlOptions, service, nodePort))
    http_helper.HttpGetWithRetryWithCustomValidation(
        t,
        url,
        &tlsConfig,
        30,
        5*time.Second,
        func(statusCode int, body string) bool {
            return statusCode == 200
        },
    )
}

// Validate no resources will change on subsequent terraform executions
func TestResourceChanges(t *testing.T, terraformDir string) {
    terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
    planResult := terraform.InitAndPlan(t, terraformOptions)
    resourceCount := terraform.GetResourceCount(t, planResult)
    assert.Equal(t, 0, resourceCount.Change)
    assert.Equal(t, 0, resourceCount.Add)
}
