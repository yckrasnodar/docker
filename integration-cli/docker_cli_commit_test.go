package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCommitAfterContainerIsDone(t *testing.T) {
	runCmd := exec.Command(dockerBinary, "run", "-i", "-a", "stdin", "busybox", "echo", "foo")
	out, _, _, err := runCommandWithStdoutStderr(runCmd)
	if err != nil {
		t.Fatalf("failed to run container: %s, %v", out, err)
	}

	cleanedContainerID := strings.TrimSpace(out)

	waitCmd := exec.Command(dockerBinary, "wait", cleanedContainerID)
	if _, _, err = runCommandWithOutput(waitCmd); err != nil {
		t.Fatalf("error thrown while waiting for container: %s, %v", out, err)
	}

	commitCmd := exec.Command(dockerBinary, "commit", cleanedContainerID)
	out, _, err = runCommandWithOutput(commitCmd)
	if err != nil {
		t.Fatalf("failed to commit container to image: %s, %v", out, err)
	}

	cleanedImageID := strings.TrimSpace(out)

	inspectCmd := exec.Command(dockerBinary, "inspect", cleanedImageID)
	if out, _, err = runCommandWithOutput(inspectCmd); err != nil {
		t.Fatalf("failed to inspect image: %s, %v", out, err)
	}

	deleteContainer(cleanedContainerID)
	deleteImages(cleanedImageID)

	logDone("commit - echo foo and commit the image")
}

func TestCommitWithoutPause(t *testing.T) {
	runCmd := exec.Command(dockerBinary, "run", "-i", "-a", "stdin", "busybox", "echo", "foo")
	out, _, _, err := runCommandWithStdoutStderr(runCmd)
	if err != nil {
		t.Fatalf("failed to run container: %s, %v", out, err)
	}

	cleanedContainerID := strings.TrimSpace(out)

	waitCmd := exec.Command(dockerBinary, "wait", cleanedContainerID)
	if _, _, err = runCommandWithOutput(waitCmd); err != nil {
		t.Fatalf("error thrown while waiting for container: %s, %v", out, err)
	}

	commitCmd := exec.Command(dockerBinary, "commit", "-p=false", cleanedContainerID)
	out, _, err = runCommandWithOutput(commitCmd)
	if err != nil {
		t.Fatalf("failed to commit container to image: %s, %v", out, err)
	}

	cleanedImageID := strings.TrimSpace(out)

	inspectCmd := exec.Command(dockerBinary, "inspect", cleanedImageID)
	if out, _, err = runCommandWithOutput(inspectCmd); err != nil {
		t.Fatalf("failed to inspect image: %s, %v", out, err)
	}

	deleteContainer(cleanedContainerID)
	deleteImages(cleanedImageID)

	logDone("commit - echo foo and commit the image with --pause=false")
}

//test commit a paused container should not unpause it after commit
func TestCommitPausedContainer(t *testing.T) {
	defer deleteAllContainers()
	defer unpauseAllContainers()
	cmd := exec.Command(dockerBinary, "run", "-i", "-d", "busybox")
	out, _, _, err := runCommandWithStdoutStderr(cmd)
	if err != nil {
		t.Fatalf("failed to run container: %v, output: %q", err, out)
	}

	cleanedContainerID := strings.TrimSpace(out)
	cmd = exec.Command(dockerBinary, "pause", cleanedContainerID)
	out, _, _, err = runCommandWithStdoutStderr(cmd)
	if err != nil {
		t.Fatalf("failed to pause container: %v, output: %q", err, out)
	}

	commitCmd := exec.Command(dockerBinary, "commit", cleanedContainerID)
	out, _, err = runCommandWithOutput(commitCmd)
	if err != nil {
		t.Fatalf("failed to commit container to image: %s, %v", out, err)
	}
	cleanedImageID := strings.TrimSpace(out)
	defer deleteImages(cleanedImageID)

	cmd = exec.Command(dockerBinary, "inspect", "-f", "{{.State.Paused}}", cleanedContainerID)
	out, _, _, err = runCommandWithStdoutStderr(cmd)
	if err != nil {
		t.Fatalf("failed to inspect container: %v, output: %q", err, out)
	}

	if !strings.Contains(out, "true") {
		t.Fatalf("commit should not unpause a paused container")
	}

	logDone("commit - commit a paused container will not unpause it")
}

func TestCommitNewFile(t *testing.T) {
	defer deleteAllContainers()

	cmd := exec.Command(dockerBinary, "run", "--name", "commiter", "busybox", "/bin/sh", "-c", "echo koye > /foo")
	if _, err := runCommand(cmd); err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command(dockerBinary, "commit", "commiter")
	imageID, _, err := runCommandWithOutput(cmd)
	if err != nil {
		t.Fatal(err)
	}
	imageID = strings.Trim(imageID, "\r\n")
	defer deleteImages(imageID)

	cmd = exec.Command(dockerBinary, "run", imageID, "cat", "/foo")

	out, _, err := runCommandWithOutput(cmd)
	if err != nil {
		t.Fatal(err, out)
	}
	if actual := strings.Trim(out, "\r\n"); actual != "koye" {
		t.Fatalf("expected output koye received %q", actual)
	}

	logDone("commit - commit file and read")
}

func TestCommitHardlink(t *testing.T) {
	defer deleteAllContainers()

	cmd := exec.Command(dockerBinary, "run", "-t", "--name", "hardlinks", "busybox", "sh", "-c", "touch file1 && ln file1 file2 && ls -di file1 file2")
	firstOuput, _, err := runCommandWithOutput(cmd)
	if err != nil {
		t.Fatal(err)
	}

	chunks := strings.Split(strings.TrimSpace(firstOuput), " ")
	inode := chunks[0]
	found := false
	for _, chunk := range chunks[1:] {
		if chunk == inode {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Failed to create hardlink in a container. Expected to find %q in %q", inode, chunks[1:])
	}

	cmd = exec.Command(dockerBinary, "commit", "hardlinks", "hardlinks")
	imageID, _, err := runCommandWithOutput(cmd)
	if err != nil {
		t.Fatal(imageID, err)
	}
	imageID = strings.Trim(imageID, "\r\n")
	defer deleteImages(imageID)

	cmd = exec.Command(dockerBinary, "run", "-t", "hardlinks", "ls", "-di", "file1", "file2")
	secondOuput, _, err := runCommandWithOutput(cmd)
	if err != nil {
		t.Fatal(err)
	}

	chunks = strings.Split(strings.TrimSpace(secondOuput), " ")
	inode = chunks[0]
	found = false
	for _, chunk := range chunks[1:] {
		if chunk == inode {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Failed to create hardlink in a container. Expected to find %q in %q", inode, chunks[1:])
	}

	logDone("commit - commit hardlinks")
}

func TestCommitTTY(t *testing.T) {
	defer deleteImages("ttytest")
	defer deleteAllContainers()

	cmd := exec.Command(dockerBinary, "run", "-t", "--name", "tty", "busybox", "/bin/ls")
	if _, err := runCommand(cmd); err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command(dockerBinary, "commit", "tty", "ttytest")
	imageID, _, err := runCommandWithOutput(cmd)
	if err != nil {
		t.Fatal(err)
	}
	imageID = strings.Trim(imageID, "\r\n")

	cmd = exec.Command(dockerBinary, "run", "ttytest", "/bin/ls")
	if _, err := runCommand(cmd); err != nil {
		t.Fatal(err)
	}

	logDone("commit - commit tty")
}

func TestCommitWithHostBindMount(t *testing.T) {
	defer deleteAllContainers()

	cmd := exec.Command(dockerBinary, "run", "--name", "bind-commit", "-v", "/dev/null:/winning", "busybox", "true")
	if _, err := runCommand(cmd); err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command(dockerBinary, "commit", "bind-commit", "bindtest")
	imageID, _, err := runCommandWithOutput(cmd)
	if err != nil {
		t.Fatal(imageID, err)
	}

	imageID = strings.Trim(imageID, "\r\n")
	defer deleteImages(imageID)

	cmd = exec.Command(dockerBinary, "run", "bindtest", "true")

	if _, err := runCommand(cmd); err != nil {
		t.Fatal(err)
	}

	logDone("commit - commit bind mounted file")
}

func TestCommitChange(t *testing.T) {
	defer deleteAllContainers()

	cmd := exec.Command(dockerBinary, "run", "--name", "test", "busybox", "true")
	if _, err := runCommand(cmd); err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command(dockerBinary, "commit",
		"--change", "EXPOSE 8080",
		"--change", "ENV DEBUG true",
		"--change", "ENV test 1",
		"--change", "ENV PATH /foo",
		"test", "test-commit")
	imageId, _, err := runCommandWithOutput(cmd)
	if err != nil {
		t.Fatal(imageId, err)
	}
	imageId = strings.Trim(imageId, "\r\n")
	defer deleteImages(imageId)

	expected := map[string]string{
		"Config.ExposedPorts": "map[8080/tcp:map[]]",
		"Config.Env":          "[DEBUG=true test=1 PATH=/foo]",
	}

	for conf, value := range expected {
		res, err := inspectField(imageId, conf)
		if err != nil {
			t.Errorf("failed to get value %s, error: %s", conf, err)
		}
		if res != value {
			t.Errorf("%s('%s'), expected %s", conf, res, value)
		}
	}

	logDone("commit - commit --change")
}

// TODO: commit --run is deprecated, remove this once --run is removed
func TestCommitMergeConfigRun(t *testing.T) {
	defer deleteAllContainers()
	name := "commit-test"
	out, _, _ := dockerCmd(t, "run", "-d", "-e=FOO=bar", "busybox", "/bin/sh", "-c", "echo testing > /tmp/foo")
	id := strings.TrimSpace(out)

	dockerCmd(t, "commit", `--run={"Cmd": ["cat", "/tmp/foo"]}`, id, "commit-test")
	defer deleteImages("commit-test")

	out, _, _ = dockerCmd(t, "run", "--name", name, "commit-test")
	if strings.TrimSpace(out) != "testing" {
		t.Fatal("run config in commited container was not merged")
	}

	type cfg struct {
		Env []string
		Cmd []string
	}
	config1 := cfg{}
	if err := inspectFieldAndMarshall(id, "Config", &config1); err != nil {
		t.Fatal(err)
	}
	config2 := cfg{}
	if err := inspectFieldAndMarshall(name, "Config", &config2); err != nil {
		t.Fatal(err)
	}

	// Env has at least PATH loaded as well here, so let's just grab the FOO one
	var env1, env2 string
	for _, e := range config1.Env {
		if strings.HasPrefix(e, "FOO") {
			env1 = e
			break
		}
	}
	for _, e := range config2.Env {
		if strings.HasPrefix(e, "FOO") {
			env2 = e
			break
		}
	}

	if len(config1.Env) != len(config2.Env) || env1 != env2 && env2 != "" {
		t.Fatalf("expected envs to match: %v - %v", config1.Env, config2.Env)
	}

	logDone("commit - configs are merged with --run")
}
