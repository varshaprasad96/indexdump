package main

import "fmt"

//import "github.com/containers/libpod/pkg/domain/entities"
import "encoding/json"
import "os"

import "strings"
import "os/exec"
import "bytes"

type ImageSummary struct {
	ID          string            `json:"Id"`
	ParentId    string            `json:",omitempty"` // nolint
	RepoTags    []string          `json:",omitempty"`
	Created     string            `json:",omitempty"`
	Size        int64             `json:",omitempty"`
	SharedSize  int               `json:",omitempty"`
	VirtualSize int64             `json:",omitempty"`
	Labels      map[string]string `json:",omitempty"`
	Containers  int               `json:",omitempty"`
	ReadOnly    bool              `json:",omitempty"`
	Dangling    bool              `json:",omitempty"`

	// Podman extensions
	Names        []string `json:",omitempty"`
	Digest       string   `json:",omitempty"`
	Digests      []string `json:",omitempty"`
	ConfigDigest string   `json:",omitempty"`
	//	History      []string `json:",omitempty"`
}

func main() {
	//bundleImage := "registry.redhat.io/amq7/amqstreams-rhel7-operator-metadata@sha256:0b98ed968b943454b4424ed4dd35b6c9bd8e4d958eaf1efeedfb605e1fe6eabd"
	//	bundleImage := "registry.redhat.io/container-native-virtualization/hco-bundle-registry@sha256:a3c97ad23758c377884be9ca53b4136b256952a660872b38d6ac2e9a6b449eaa"
	//	bundleImage := "registry.redhat.io/integration/service-registry-rhel8-operator-metadata@sha256:31c7dd275bc4c2b0d9ead9c2002963485791b279a90174191826236ac624a44b"
	bundleImage := "registry.redhat.io/rhmtc/openshift-migration-operator-bundle@sha256:f53faabf65c9a610cc2c840db941a284ccc4a9665f9f16ae226a2ec8262dd17e"
	sha, err := pullBundleImage(bundleImage)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("sha %s\n", sha)

	var inspectOutput string
	inspectOutput, err = inspectImage(strings.TrimSpace(sha))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(inspectOutput)
	operatorType, sdkVersion, err := printLabels(inspectOutput)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("operator type [%s] sdk version [%s]\n", operatorType, sdkVersion)

}

func printLabels(inspectOutput string) (operatorType string, sdkversion string, err error) {
	//convert string into object
	var i []ImageSummary
	err = json.Unmarshal([]byte(inspectOutput), &i)
	if err != nil {
		fmt.Println(err.Error())
		return "", "", err
	}
	//	fmt.Printf("images len %d\n", len(i))
	if i[0].Labels == nil {
		fmt.Println("labels are nil")
		return "", "", err
	}
	//	fmt.Printf("labels are %+v\n", i[0].Labels)
	for k, v := range i[0].Labels {
		if k == "operators.operatorframework.io.metrics.builder" {
			sdkversion = v
		}
		if k == "operators.operatorframework.io.metrics.project_layout" {

			fmt.Printf("[%s][%s]\n", k, v)
			if strings.Contains(v, "ansible") {
				operatorType = "ansible"
			}
			if strings.Contains(v, "helm") {
				operatorType = "helm"
			}
			if strings.Contains(v, "go") {
				operatorType = "golang"
			}
		}
	}
	return operatorType, sdkversion, nil

}

func pullBundleImage(bundlePath string) (sha string, err error) {

	var stdout bytes.Buffer
	cmd := &exec.Cmd{
		Path:   "/usr/bin/podman",
		Args:   []string{"/usr/bin/podman", "pull", bundlePath, "--quiet"},
		Stdout: &stdout,
		Stderr: os.Stderr,
	}

	err = cmd.Run()
	return stdout.String(), err

}

func inspectImage(bundlePath string) (imageOutput string, err error) {
	var stdout bytes.Buffer
	//var stderr bytes.Buffer
	cmd := &exec.Cmd{
		Path:   "/usr/bin/podman",
		Args:   []string{"/usr/bin/podman", "inspect", bundlePath, "--format", "json"},
		Stdout: &stdout,
		Stderr: os.Stderr,
	}

	err = cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return stdout.String(), err

}
