package main

import (
	"net/http"
	"os/exec"
	"strings"
	"testing"
)

func TestResizeApiResponse(t *testing.T) {
	runCmd := exec.Command(dockerBinary, "run", "-d", "busybox", "top")
	out, _, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatalf(out, err)
	}
	defer deleteAllContainers()
	cleanedContainerID := strings.TrimSpace(out)

	endpoint := "/containers/" + cleanedContainerID + "/resize?h=40&w=40"
	_, _, err = sockRequest("POST", endpoint, nil)
	if err != nil {
		t.Fatalf("resize Request failed %v", err)
	}

	logDone("container resize - when started")
}

func TestResizeApiHeightWidthNoInt(t *testing.T) {
	runCmd := exec.Command(dockerBinary, "run", "-d", "busybox", "top")
	out, _, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatalf(out, err)
	}
	defer deleteAllContainers()
	cleanedContainerID := strings.TrimSpace(out)

	endpoint := "/containers/" + cleanedContainerID + "/resize?h=foo&w=bar"
	status, _, err := sockRequest("POST", endpoint, nil)
	if err == nil {
		t.Fatal("Expected resize Request to fail")
	}
	if status != http.StatusInternalServerError {
		t.Fatalf("Status expected %d, got %d", http.StatusInternalServerError, status)
	}

	logDone("container resize - height, width no int fail")
}

func TestResizeApiResponseWhenContainerNotStarted(t *testing.T) {
	runCmd := exec.Command(dockerBinary, "run", "-d", "busybox", "true")
	out, _, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatalf(out, err)
	}
	defer deleteAllContainers()
	cleanedContainerID := strings.TrimSpace(out)

	// make sure the exited container is not running
	runCmd = exec.Command(dockerBinary, "wait", cleanedContainerID)
	out, _, err = runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatalf(out, err)
	}

	endpoint := "/containers/" + cleanedContainerID + "/resize?h=40&w=40"
	_, body, err := sockRequest("POST", endpoint, nil)
	if err == nil {
		t.Fatalf("resize should fail when container is not started")
	}
	if !strings.Contains(string(body), "Cannot resize container") && !strings.Contains(string(body), cleanedContainerID) {
		t.Fatalf("resize should fail with message 'Cannot resize container' but instead received %s", string(body))
	}

	logDone("container resize - when not started should not resize")
}
