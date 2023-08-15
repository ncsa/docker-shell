package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: docker-shell [additional docker run options] <DOCKER_IMAGE>")
		os.Exit(1)
	}

	mountPoints, err := getMountPoints()
	if err != nil {
		log.Fatalf("Error getting mount points: %v", err)
	}

	volumeArgs := generateVolumeArgs(mountPoints)

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Error getting hostname: %v", err)
	}
	containerName := os.Args[len(os.Args)-1]
	customHostname := fmt.Sprintf("%s{docker|%s}", hostname, containerName)

	fmt.Println(customHostname)

	// Constructing the docker run command
	cmdArgs := []string{
		"run", "-it",
		"--hostname", customHostname,
		"--user", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		"--volume", fmt.Sprintf("%s:%s", os.Getenv("HOME"), os.Getenv("HOME")),
		"--volume", "/etc/passwd:/etc/passwd:ro",
		"--volume", "/etc/group:/etc/group:ro",
		"--workdir", os.Getenv("PWD"),
		"--env", fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
		"--env", fmt.Sprintf("USER=%s", os.Getenv("USER")),
	}
	cmdArgs = append(cmdArgs, volumeArgs...)
	cmdArgs = append(cmdArgs, os.Args[1:]...)
	cmdArgs = append(cmdArgs, "/bin/bash")

	cmd := exec.Command("docker", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run docker command: %v", err)
	}
}

func getMountPoints() ([]string, error) {
	var mountPoints []string

	cmd := exec.Command("df", "--output=source")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	devices := strings.Split(string(output), "\n")
	for _, device := range devices {
		if strings.HasPrefix(device, "/dev/") {
			if isBlockDevice(device) && !isLoopDevice(device) {
				mountPoints = append(mountPoints, device)
			}
		}
	}

	return mountPoints, nil
}

func isBlockDevice(device string) bool {
	cmd := exec.Command("ls", "-l", device)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false
	}
	return strings.HasPrefix(out.String(), "b")
}

func isLoopDevice(device string) bool {
	cmd := exec.Command("losetup", device)
	return cmd.Run() == nil
}

func generateVolumeArgs(mountPoints []string) []string {
	var volumeArgs []string
	for _, device := range mountPoints {
		cmd := exec.Command("df", "--output=target,source")
		output, err := cmd.Output()
		if err != nil {
			continue
		}
		mount := strings.Split(string(output), "\n")[0]
		if mount != "/" {
			volumeArgs = append(volumeArgs, "--volume", fmt.Sprintf("%s:%s", mount, mount))
		}
	}
	return volumeArgs
}
