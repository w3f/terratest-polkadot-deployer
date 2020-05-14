package terratest_polkadot_deployer

import (
    "io/ioutil"
    "os"
    "testing"

    "github.com/gruntwork-io/terratest/modules/k8s"
    "github.com/gruntwork-io/terratest/modules/random"
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

// Create temporary file
func CreateTempFile(t *testing.T, content []byte) *os.File {
    tempFile, err := ioutil.TempFile(os.TempDir(), random.UniqueId())
    if err != nil {
        t.Fatal("Cannot create temporary file", err)
    }

    if _, err = tempFile.Write(content); err != nil {
        t.Fatal("Failed to write to temporary file", err)
    }
    if err := tempFile.Close(); err != nil {
        t.Fatal(err)
    }

    return tempFile
}

// Write kubeconfig file from terraform output and configure to use it kubectl
func SetupKubeconfig(t *testing.T, terraformDir string) {
    terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
    kubeconfig := terraform.Output(t, terraformOptions, "kubeconfig")
    kubeconfigFile := CreateTempFile(t, []byte(kubeconfig))
    kubectlOptions := k8s.NewKubectlOptions("", kubeconfigFile.Name(), "default")
    test_structure.SaveKubectlOptions(t, terraformDir, kubectlOptions)
}

// Get k8s node's IP address of a giving type
func GetNodeAddress(t *testing.T, kubectlOptions *k8s.KubectlOptions, addrType string) string {
    nodes := k8s.GetNodes(t, kubectlOptions)
    for _, addr := range nodes[0].Status.Addresses {
        if string(addr.Type) == addrType {
            return addr.Address
        }
    }
    t.Fatalf("%s address is not available", addrType)
    return ""
}
