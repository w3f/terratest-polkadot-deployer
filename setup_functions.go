package terratest_polkadot_deployer

import (
    "os"
    "testing"

    "github.com/gruntwork-io/terratest/modules/k8s"
    "github.com/gruntwork-io/terratest/modules/terraform"
    "github.com/gruntwork-io/terratest/modules/test-structure"

    "github.com/stretchr/testify/require"
)

// Deploy terraform module to cloud provider
func CreateTerraformStack(t *testing.T, terraformDir string) {
    terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
    terraform.InitAndApply(t, terraformOptions)
}

// Destroy previously created terraform stack
func DestroyTerraformStack(t *testing.T, terraformDir string) {
    terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
    terraform.Destroy(t, terraformOptions)

    kubectlOptions := test_structure.LoadKubectlOptions(t, terraformDir)
    err := os.Remove(kubectlOptions.ConfigPath)
    require.NoError(t, err)
}

// Write kubeconfig file from terraform output and configure to use it kubectl
func SetupKubeconfig(t *testing.T, terraformDir string) {
    terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
    kubeconfig := terraform.Output(t, terraformOptions, "kubeconfig")
    kubeconfigFile := CreateTempFile(t, []byte(kubeconfig))
    kubectlOptions := k8s.NewKubectlOptions("", kubeconfigFile.Name(), "default")
    test_structure.SaveKubectlOptions(t, terraformDir, kubectlOptions)
}
