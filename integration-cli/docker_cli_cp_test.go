package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

const (
	cpTestPathParent = "/some"
	cpTestPath       = "/some/path"
	cpTestName       = "test"
	cpFullPath       = "/some/path/test"

	cpContainerContents = "holla, i am the container"
	cpHostContents      = "hello, i am the host"
)

// Test for #5656
// Check that garbage paths don't escape the container's rootfs
func TestCpGarbagePath(t *testing.T) {
	out, exitCode, err := dockerCmd(t, "run", "-d", "busybox", "/bin/sh", "-c", "mkdir -p '"+cpTestPath+"' && echo -n '"+cpContainerContents+"' > "+cpFullPath)
	if err != nil || exitCode != 0 {
		t.Fatal("failed to create a container", out, err)
	}

	cleanedContainerID := strings.TrimSpace(out)
	defer deleteContainer(cleanedContainerID)

	out, _, err = dockerCmd(t, "wait", cleanedContainerID)
	if err != nil || strings.TrimSpace(out) != "0" {
		t.Fatal("failed to set up container", out, err)
	}

	if err := os.MkdirAll(cpTestPath, os.ModeDir); err != nil {
		t.Fatal(err)
	}

	hostFile, err := os.Create(cpFullPath)
	if err != nil {
		t.Fatal(err)
	}
	defer hostFile.Close()
	defer os.RemoveAll(cpTestPathParent)

	fmt.Fprintf(hostFile, "%s", cpHostContents)

	tmpdir, err := ioutil.TempDir("", "docker-integration")
	if err != nil {
		t.Fatal(err)
	}

	tmpname := filepath.Join(tmpdir, cpTestName)
	defer os.RemoveAll(tmpdir)

	path := path.Join("../../../../../../../../../../../../", cpFullPath)

	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":"+path, tmpdir)
	if err != nil {
		t.Fatalf("couldn't copy from garbage path: %s:%s %s", cleanedContainerID, path, err)
	}

	file, _ := os.Open(tmpname)
	defer file.Close()

	test, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	if string(test) == cpHostContents {
		t.Errorf("output matched host file -- garbage path can escape container rootfs")
	}

	if string(test) != cpContainerContents {
		t.Errorf("output doesn't match the input for garbage path")
	}

	logDone("cp - garbage paths relative to container's rootfs")
}

// Check that relative paths are relative to the container's rootfs
func TestCpRelativePath(t *testing.T) {
	out, exitCode, err := dockerCmd(t, "run", "-d", "busybox", "/bin/sh", "-c", "mkdir -p '"+cpTestPath+"' && echo -n '"+cpContainerContents+"' > "+cpFullPath)
	if err != nil || exitCode != 0 {
		t.Fatal("failed to create a container", out, err)
	}

	cleanedContainerID := strings.TrimSpace(out)
	defer deleteContainer(cleanedContainerID)

	out, _, err = dockerCmd(t, "wait", cleanedContainerID)
	if err != nil || strings.TrimSpace(out) != "0" {
		t.Fatal("failed to set up container", out, err)
	}

	if err := os.MkdirAll(cpTestPath, os.ModeDir); err != nil {
		t.Fatal(err)
	}

	hostFile, err := os.Create(cpFullPath)
	if err != nil {
		t.Fatal(err)
	}
	defer hostFile.Close()
	defer os.RemoveAll(cpTestPathParent)

	fmt.Fprintf(hostFile, "%s", cpHostContents)

	tmpdir, err := ioutil.TempDir("", "docker-integration")

	if err != nil {
		t.Fatal(err)
	}

	tmpname := filepath.Join(tmpdir, cpTestName)
	defer os.RemoveAll(tmpdir)

	var relPath string
	if path.IsAbs(cpFullPath) {
		// normally this is `filepath.Rel("/", cpFullPath)` but we cannot
		// get this unix-path manipulation on windows with filepath.
		relPath = cpFullPath[1:]
	} else {
		t.Fatalf("path %s was assumed to be an absolute path", cpFullPath)
	}

	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":"+relPath, tmpdir)
	if err != nil {
		t.Fatalf("couldn't copy from relative path: %s:%s %s", cleanedContainerID, relPath, err)
	}

	file, _ := os.Open(tmpname)
	defer file.Close()

	test, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	if string(test) == cpHostContents {
		t.Errorf("output matched host file -- relative path can escape container rootfs")
	}

	if string(test) != cpContainerContents {
		t.Errorf("output doesn't match the input for relative path")
	}

	logDone("cp - relative paths relative to container's rootfs")
}

// Check that absolute paths are relative to the container's rootfs
func TestCpAbsolutePath(t *testing.T) {
	out, exitCode, err := dockerCmd(t, "run", "-d", "busybox", "/bin/sh", "-c", "mkdir -p '"+cpTestPath+"' && echo -n '"+cpContainerContents+"' > "+cpFullPath)
	if err != nil || exitCode != 0 {
		t.Fatal("failed to create a container", out, err)
	}

	cleanedContainerID := strings.TrimSpace(out)
	defer deleteContainer(cleanedContainerID)

	out, _, err = dockerCmd(t, "wait", cleanedContainerID)
	if err != nil || strings.TrimSpace(out) != "0" {
		t.Fatal("failed to set up container", out, err)
	}

	if err := os.MkdirAll(cpTestPath, os.ModeDir); err != nil {
		t.Fatal(err)
	}

	hostFile, err := os.Create(cpFullPath)
	if err != nil {
		t.Fatal(err)
	}
	defer hostFile.Close()
	defer os.RemoveAll(cpTestPathParent)

	fmt.Fprintf(hostFile, "%s", cpHostContents)

	tmpdir, err := ioutil.TempDir("", "docker-integration")

	if err != nil {
		t.Fatal(err)
	}

	tmpname := filepath.Join(tmpdir, cpTestName)
	defer os.RemoveAll(tmpdir)

	path := cpFullPath

	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":"+path, tmpdir)
	if err != nil {
		t.Fatalf("couldn't copy from absolute path: %s:%s %s", cleanedContainerID, path, err)
	}

	file, _ := os.Open(tmpname)
	defer file.Close()

	test, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	if string(test) == cpHostContents {
		t.Errorf("output matched host file -- absolute path can escape container rootfs")
	}

	if string(test) != cpContainerContents {
		t.Errorf("output doesn't match the input for absolute path")
	}

	logDone("cp - absolute paths relative to container's rootfs")
}

// Test for #5619
// Check that absolute symlinks are still relative to the container's rootfs
func TestCpAbsoluteSymlink(t *testing.T) {
	out, exitCode, err := dockerCmd(t, "run", "-d", "busybox", "/bin/sh", "-c", "mkdir -p '"+cpTestPath+"' && echo -n '"+cpContainerContents+"' > "+cpFullPath+" && ln -s "+cpFullPath+" container_path")
	if err != nil || exitCode != 0 {
		t.Fatal("failed to create a container", out, err)
	}

	cleanedContainerID := strings.TrimSpace(out)
	defer deleteContainer(cleanedContainerID)

	out, _, err = dockerCmd(t, "wait", cleanedContainerID)
	if err != nil || strings.TrimSpace(out) != "0" {
		t.Fatal("failed to set up container", out, err)
	}

	if err := os.MkdirAll(cpTestPath, os.ModeDir); err != nil {
		t.Fatal(err)
	}

	hostFile, err := os.Create(cpFullPath)
	if err != nil {
		t.Fatal(err)
	}
	defer hostFile.Close()
	defer os.RemoveAll(cpTestPathParent)

	fmt.Fprintf(hostFile, "%s", cpHostContents)

	tmpdir, err := ioutil.TempDir("", "docker-integration")

	if err != nil {
		t.Fatal(err)
	}

	tmpname := filepath.Join(tmpdir, cpTestName)
	defer os.RemoveAll(tmpdir)

	path := path.Join("/", "container_path")

	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":"+path, tmpdir)
	if err != nil {
		t.Fatalf("couldn't copy from absolute path: %s:%s %s", cleanedContainerID, path, err)
	}

	file, _ := os.Open(tmpname)
	defer file.Close()

	test, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	if string(test) == cpHostContents {
		t.Errorf("output matched host file -- absolute symlink can escape container rootfs")
	}

	if string(test) != cpContainerContents {
		t.Errorf("output doesn't match the input for absolute symlink")
	}

	logDone("cp - absolute symlink relative to container's rootfs")
}

// Test for #5619
// Check that symlinks which are part of the resource path are still relative to the container's rootfs
func TestCpSymlinkComponent(t *testing.T) {
	out, exitCode, err := dockerCmd(t, "run", "-d", "busybox", "/bin/sh", "-c", "mkdir -p '"+cpTestPath+"' && echo -n '"+cpContainerContents+"' > "+cpFullPath+" && ln -s "+cpTestPath+" container_path")
	if err != nil || exitCode != 0 {
		t.Fatal("failed to create a container", out, err)
	}

	cleanedContainerID := strings.TrimSpace(out)
	defer deleteContainer(cleanedContainerID)

	out, _, err = dockerCmd(t, "wait", cleanedContainerID)
	if err != nil || strings.TrimSpace(out) != "0" {
		t.Fatal("failed to set up container", out, err)
	}

	if err := os.MkdirAll(cpTestPath, os.ModeDir); err != nil {
		t.Fatal(err)
	}

	hostFile, err := os.Create(cpFullPath)
	if err != nil {
		t.Fatal(err)
	}
	defer hostFile.Close()
	defer os.RemoveAll(cpTestPathParent)

	fmt.Fprintf(hostFile, "%s", cpHostContents)

	tmpdir, err := ioutil.TempDir("", "docker-integration")

	if err != nil {
		t.Fatal(err)
	}

	tmpname := filepath.Join(tmpdir, cpTestName)
	defer os.RemoveAll(tmpdir)

	path := path.Join("/", "container_path", cpTestName)

	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":"+path, tmpdir)
	if err != nil {
		t.Fatalf("couldn't copy from symlink path component: %s:%s %s", cleanedContainerID, path, err)
	}

	file, _ := os.Open(tmpname)
	defer file.Close()

	test, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	if string(test) == cpHostContents {
		t.Errorf("output matched host file -- symlink path component can escape container rootfs")
	}

	if string(test) != cpContainerContents {
		t.Errorf("output doesn't match the input for symlink path component")
	}

	logDone("cp - symlink path components relative to container's rootfs")
}

// Check that cp with unprivileged user doesn't return any error
func TestCpUnprivilegedUser(t *testing.T) {
	testRequires(t, UnixCli) // uses chmod/su: not available on windows

	out, exitCode, err := dockerCmd(t, "run", "-d", "busybox", "/bin/sh", "-c", "touch "+cpTestName)
	if err != nil || exitCode != 0 {
		t.Fatal("failed to create a container", out, err)
	}

	cleanedContainerID := strings.TrimSpace(out)
	defer deleteContainer(cleanedContainerID)

	out, _, err = dockerCmd(t, "wait", cleanedContainerID)
	if err != nil || strings.TrimSpace(out) != "0" {
		t.Fatal("failed to set up container", out, err)
	}

	tmpdir, err := ioutil.TempDir("", "docker-integration")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmpdir)

	if err = os.Chmod(tmpdir, 0777); err != nil {
		t.Fatal(err)
	}

	path := cpTestName

	_, _, err = runCommandWithOutput(exec.Command("su", "unprivilegeduser", "-c", dockerBinary+" cp "+cleanedContainerID+":"+path+" "+tmpdir))
	if err != nil {
		t.Fatalf("couldn't copy with unprivileged user: %s:%s %s", cleanedContainerID, path, err)
	}

	logDone("cp - unprivileged user")
}

func TestCpSpecialFiles(t *testing.T) {
	testRequires(t, SameHostDaemon)

	outDir, err := ioutil.TempDir("", "cp-test-special-files")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outDir)

	out, exitCode, err := dockerCmd(t, "run", "-d", "busybox", "/bin/sh", "-c", "touch /foo")
	if err != nil || exitCode != 0 {
		t.Fatal("failed to create a container", out, err)
	}

	cleanedContainerID := strings.TrimSpace(out)
	defer deleteContainer(cleanedContainerID)

	out, _, err = dockerCmd(t, "wait", cleanedContainerID)
	if err != nil || strings.TrimSpace(out) != "0" {
		t.Fatal("failed to set up container", out, err)
	}

	// Copy actual /etc/resolv.conf
	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":/etc/resolv.conf", outDir)
	if err != nil {
		t.Fatalf("couldn't copy from container: %s:%s %v", cleanedContainerID, "/etc/resolv.conf", err)
	}

	expected, err := ioutil.ReadFile("/var/lib/docker/containers/" + cleanedContainerID + "/resolv.conf")
	actual, err := ioutil.ReadFile(outDir + "/resolv.conf")

	if !bytes.Equal(actual, expected) {
		t.Fatalf("Expected copied file to be duplicate of the container resolvconf")
	}

	// Copy actual /etc/hosts
	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":/etc/hosts", outDir)
	if err != nil {
		t.Fatalf("couldn't copy from container: %s:%s %v", cleanedContainerID, "/etc/hosts", err)
	}

	expected, err = ioutil.ReadFile("/var/lib/docker/containers/" + cleanedContainerID + "/hosts")
	actual, err = ioutil.ReadFile(outDir + "/hosts")

	if !bytes.Equal(actual, expected) {
		t.Fatalf("Expected copied file to be duplicate of the container hosts")
	}

	// Copy actual /etc/resolv.conf
	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":/etc/hostname", outDir)
	if err != nil {
		t.Fatalf("couldn't copy from container: %s:%s %v", cleanedContainerID, "/etc/hostname", err)
	}

	expected, err = ioutil.ReadFile("/var/lib/docker/containers/" + cleanedContainerID + "/hostname")
	actual, err = ioutil.ReadFile(outDir + "/hostname")

	if !bytes.Equal(actual, expected) {
		t.Fatalf("Expected copied file to be duplicate of the container resolvconf")
	}

	logDone("cp - special files (resolv.conf, hosts, hostname)")
}

func TestCpVolumePath(t *testing.T) {
	testRequires(t, SameHostDaemon)

	tmpDir, err := ioutil.TempDir("", "cp-test-volumepath")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	outDir, err := ioutil.TempDir("", "cp-test-volumepath-out")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outDir)
	_, err = os.Create(tmpDir + "/test")
	if err != nil {
		t.Fatal(err)
	}

	out, exitCode, err := dockerCmd(t, "run", "-d", "-v", "/foo", "-v", tmpDir+"/test:/test", "-v", tmpDir+":/baz", "busybox", "/bin/sh", "-c", "touch /foo/bar")
	if err != nil || exitCode != 0 {
		t.Fatal("failed to create a container", out, err)
	}

	cleanedContainerID := strings.TrimSpace(out)
	defer dockerCmd(t, "rm", "-fv", cleanedContainerID)

	out, _, err = dockerCmd(t, "wait", cleanedContainerID)
	if err != nil || strings.TrimSpace(out) != "0" {
		t.Fatal("failed to set up container", out, err)
	}

	// Copy actual volume path
	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":/foo", outDir)
	if err != nil {
		t.Fatalf("couldn't copy from volume path: %s:%s %v", cleanedContainerID, "/foo", err)
	}
	stat, err := os.Stat(outDir + "/foo")
	if err != nil {
		t.Fatal(err)
	}
	if !stat.IsDir() {
		t.Fatal("expected copied content to be dir")
	}
	stat, err = os.Stat(outDir + "/foo/bar")
	if err != nil {
		t.Fatal(err)
	}
	if stat.IsDir() {
		t.Fatal("Expected file `bar` to be a file")
	}

	// Copy file nested in volume
	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":/foo/bar", outDir)
	if err != nil {
		t.Fatalf("couldn't copy from volume path: %s:%s %v", cleanedContainerID, "/foo", err)
	}
	stat, err = os.Stat(outDir + "/bar")
	if err != nil {
		t.Fatal(err)
	}
	if stat.IsDir() {
		t.Fatal("Expected file `bar` to be a file")
	}

	// Copy Bind-mounted dir
	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":/baz", outDir)
	if err != nil {
		t.Fatalf("couldn't copy from bind-mounted volume path: %s:%s %v", cleanedContainerID, "/baz", err)
	}
	stat, err = os.Stat(outDir + "/baz")
	if err != nil {
		t.Fatal(err)
	}
	if !stat.IsDir() {
		t.Fatal("Expected `baz` to be a dir")
	}

	// Copy file nested in bind-mounted dir
	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":/baz/test", outDir)
	fb, err := ioutil.ReadFile(outDir + "/baz/test")
	if err != nil {
		t.Fatal(err)
	}
	fb2, err := ioutil.ReadFile(tmpDir + "/test")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fb, fb2) {
		t.Fatalf("Expected copied file to be duplicate of bind-mounted file")
	}

	// Copy bind-mounted file
	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":/test", outDir)
	fb, err = ioutil.ReadFile(outDir + "/test")
	if err != nil {
		t.Fatal(err)
	}
	fb2, err = ioutil.ReadFile(tmpDir + "/test")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fb, fb2) {
		t.Fatalf("Expected copied file to be duplicate of bind-mounted file")
	}

	logDone("cp - volume path")
}

func TestCpToDot(t *testing.T) {
	out, exitCode, err := dockerCmd(t, "run", "-d", "busybox", "/bin/sh", "-c", "echo lololol > /test")
	if err != nil || exitCode != 0 {
		t.Fatal("failed to create a container", out, err)
	}

	cleanedContainerID := strings.TrimSpace(out)
	defer deleteContainer(cleanedContainerID)

	out, _, err = dockerCmd(t, "wait", cleanedContainerID)
	if err != nil || strings.TrimSpace(out) != "0" {
		t.Fatal("failed to set up container", out, err)
	}

	tmpdir, err := ioutil.TempDir("", "docker-integration")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	if err := os.Chdir(tmpdir); err != nil {
		t.Fatal(err)
	}
	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":/test", ".")
	if err != nil {
		t.Fatalf("couldn't docker cp to \".\" path: %s", err)
	}
	content, err := ioutil.ReadFile("./test")
	if string(content) != "lololol\n" {
		t.Fatalf("Wrong content in copied file %q, should be %q", content, "lololol\n")
	}
	logDone("cp - to dot path")
}

func TestCpToStdout(t *testing.T) {
	out, exitCode, err := dockerCmd(t, "run", "-d", "busybox", "/bin/sh", "-c", "echo lololol > /test")
	if err != nil || exitCode != 0 {
		t.Fatalf("failed to create a container:%s\n%s", out, err)
	}

	cID := strings.TrimSpace(out)
	defer deleteContainer(cID)

	out, _, err = dockerCmd(t, "wait", cID)
	if err != nil || strings.TrimSpace(out) != "0" {
		t.Fatalf("failed to set up container:%s\n%s", out, err)
	}

	out, _, err = runCommandPipelineWithOutput(
		exec.Command(dockerBinary, "cp", cID+":/test", "-"),
		exec.Command("tar", "-vtf", "-"))
	if err != nil {
		t.Fatalf("Failed to run commands: %s", err)
	}

	if !strings.Contains(out, "test") || !strings.Contains(out, "-rw") {
		t.Fatalf("Missing file from tar TOC:\n%s", out)
	}
	logDone("cp - to stdout")
}

func TestCpNameHasColon(t *testing.T) {
	testRequires(t, SameHostDaemon)

	out, exitCode, err := dockerCmd(t, "run", "-d", "busybox", "/bin/sh", "-c", "echo lololol > /te:s:t")
	if err != nil || exitCode != 0 {
		t.Fatal("failed to create a container", out, err)
	}

	cleanedContainerID := strings.TrimSpace(out)
	defer deleteContainer(cleanedContainerID)

	out, _, err = dockerCmd(t, "wait", cleanedContainerID)
	if err != nil || strings.TrimSpace(out) != "0" {
		t.Fatal("failed to set up container", out, err)
	}

	tmpdir, err := ioutil.TempDir("", "docker-integration")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	_, _, err = dockerCmd(t, "cp", cleanedContainerID+":/te:s:t", tmpdir)
	if err != nil {
		t.Fatalf("couldn't docker cp to %s: %s", tmpdir, err)
	}
	content, err := ioutil.ReadFile(tmpdir + "/te:s:t")
	if string(content) != "lololol\n" {
		t.Fatalf("Wrong content in copied file %q, should be %q", content, "lololol\n")
	}
	logDone("cp - copy filename has ':'")
}
