package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// export an image and try to import it into a new one
func TestExportContainerAndImportImage(t *testing.T) {
	containerID := "testexportcontainerandimportimage"

	defer deleteImages("repo/testexp:v1")
	defer deleteContainer(containerID)

	runCmd := exec.Command(dockerBinary, "run", "-d", "--name", containerID, "busybox", "true")
	out, _, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatal("failed to create a container", out, err)
	}

	inspectCmd := exec.Command(dockerBinary, "inspect", containerID)
	out, _, err = runCommandWithOutput(inspectCmd)
	if err != nil {
		t.Fatalf("output should've been a container id: %s %s ", containerID, err)
	}

	exportCmd := exec.Command(dockerBinary, "export", containerID)
	if out, _, err = runCommandWithOutput(exportCmd); err != nil {
		t.Fatalf("failed to export container: %s, %v", out, err)
	}

	importCmd := exec.Command(dockerBinary, "import", "-", "repo/testexp:v1")
	importCmd.Stdin = strings.NewReader(out)
	out, _, err = runCommandWithOutput(importCmd)
	if err != nil {
		t.Fatalf("failed to import image: %s, %v", out, err)
	}

	cleanedImageID := strings.TrimSpace(out)

	inspectCmd = exec.Command(dockerBinary, "inspect", cleanedImageID)
	if out, _, err = runCommandWithOutput(inspectCmd); err != nil {
		t.Fatalf("output should've been an image id: %s, %v", out, err)
	}

	logDone("export - export/import a container/image")
}

// Used to test output flag in the export command
func TestExportContainerWithOutputAndImportImage(t *testing.T) {
	containerID := "testexportcontainerwithoutputandimportimage"

	defer deleteImages("repo/testexp:v1")
	defer deleteContainer(containerID)

	runCmd := exec.Command(dockerBinary, "run", "-d", "--name", containerID, "busybox", "true")
	out, _, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatal("failed to create a container", out, err)
	}

	inspectCmd := exec.Command(dockerBinary, "inspect", containerID)
	out, _, err = runCommandWithOutput(inspectCmd)
	if err != nil {
		t.Fatalf("output should've been a container id: %s %s ", containerID, err)
	}

	defer os.Remove("testexp.tar")

	exportCmd := exec.Command(dockerBinary, "export", "--output=testexp.tar", containerID)
	if out, _, err = runCommandWithOutput(exportCmd); err != nil {
		t.Fatalf("failed to export container: %s, %v", out, err)
	}

	out, _, err = runCommandWithOutput(exec.Command("cat", "testexp.tar"))
	if err != nil {
		t.Fatal(out, err)
	}

	importCmd := exec.Command(dockerBinary, "import", "-", "repo/testexp:v1")
	importCmd.Stdin = strings.NewReader(out)
	out, _, err = runCommandWithOutput(importCmd)
	if err != nil {
		t.Fatalf("failed to import image: %s, %v", out, err)
	}

	cleanedImageID := strings.TrimSpace(out)

	inspectCmd = exec.Command(dockerBinary, "inspect", cleanedImageID)
	if out, _, err = runCommandWithOutput(inspectCmd); err != nil {
		t.Fatalf("output should've been an image id: %s, %v", out, err)
	}

	logDone("export - export/import a container/image with output flag")
}
