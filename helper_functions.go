package terratest_polkadot_deployer

import (
    "fmt"
    "io/ioutil"
    "os"
    "testing"

    corev1 "k8s.io/api/core/v1"

    "github.com/gruntwork-io/terratest/modules/k8s"
    "github.com/gruntwork-io/terratest/modules/random"

    "github.com/stretchr/testify/require"
)

// Create temporary file with passed content
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

// Get publicly accessible service endpoint
func GetPublicServiceEndpoint(t *testing.T, kubectlOptions *k8s.KubectlOptions, service *corev1.Service, servicePort int) string {
    switch service.Spec.Type {
    case corev1.ServiceTypeNodePort:
        return GetNodePortServiceEndpoint(t, kubectlOptions, "ExternalIP", servicePort)
    default:
        return k8s.GetServiceEndpoint(t, kubectlOptions, service, servicePort)
    }
    return ""
}

// Get NodePort service endpoint
func GetNodePortServiceEndpoint(t *testing.T, kubectlOptions *k8s.KubectlOptions, addrType string, servicePort int) string {
    address := GetRandomNodeAddress(t, kubectlOptions, addrType)
    return fmt.Sprintf("%s:%d", address, servicePort)
}

// Get k8s random node's IP address of a giving type
func GetRandomNodeAddress(t *testing.T, kubectlOptions *k8s.KubectlOptions, addrType string) string {
    nodes := k8s.GetNodes(t, kubectlOptions)
    require.NotEqual(t, 0, len(nodes), "Cluster has no nodes")

    index := random.Random(0, len(nodes)-1)
    node := nodes[index]

    address := ""
    for _, addr := range node.Status.Addresses {
        if string(addr.Type) == addrType {
            address = addr.Address
        }
    }
    require.NotEqualf(t, "", address, "No available address of type: %s", addrType)

    return address
}
