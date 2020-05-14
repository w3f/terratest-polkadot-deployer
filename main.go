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

    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/assert"
)

// Test that the Node count matches the Terraform specification
func ValidateNodeCount(t *testing.T, terraformDir string) {
    kubectlOptions := test_structure.LoadKubectlOptions(t, terraformDir)
    nodeCount := test_structure.LoadInt(t, terraformDir, "nodeCount")

    k8s.WaitUntilAllNodesReady(t, kubectlOptions, 40, 10*time.Second)
    nodes := k8s.GetNodes(t, kubectlOptions)
    assert.Equal(t, len(nodes), nodeCount)
}

// Test service deployment and verify it's availability on configured nodePort
func ValidateServiceAvailability(t *testing.T, terraformDir string) {
    kubectlOptions := test_structure.LoadKubectlOptions(t, terraformDir)
    nodePort := test_structure.LoadInt(t, terraformDir, "nodePort")

    helmOptions := &helm.Options{
        KubectlOptions: kubectlOptions,
        SetValues: map[string]string{
            "image.repo": "nginx",
            "image.tag":  "1.8",
            "nodePort":   fmt.Sprintf("%d", nodePort),
        },
    }

    helmChartPath, err := filepath.Abs("fixtures/nginx-chart")
    require.NoError(t, err)

    helmReleaseName := fmt.Sprintf("polkadot-nginx-%s", strings.ToLower(random.UniqueId()))
    defer helm.Delete(t, helmOptions, helmReleaseName, true)
    helm.Install(t, helmOptions, helmChartPath, helmReleaseName)

    // Validate service availability
    k8s.WaitUntilServiceAvailable(t, kubectlOptions, helmReleaseName, 20, 5*time.Second)
    service := k8s.GetService(t, kubectlOptions, helmReleaseName)
    require.Equal(t, service.Name, helmReleaseName)

    // Check external connectivity
    tlsConfig := tls.Config{}
    url := fmt.Sprintf("http://%s:%d", GetNodeAddress(t, kubectlOptions, "ExternalIP"), nodePort)
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
