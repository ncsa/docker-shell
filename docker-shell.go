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

    // Get the df output
    dfMap, err := getDfOutputMap()
    if err != nil {
        log.Fatalf("Error getting df output: %v", err)
    }

    // Detect the non-loopback block device mount points
    mountPoints := getMountPoints(dfMap)

    // Generate the volume arguments
    volumeArgs := generateVolumeArgs(dfMap, mountPoints)

    // Generate the custom hostname
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

    // Run the docker command
    cmd := exec.Command("docker", cmdArgs...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        log.Fatalf("Failed to run docker command: %v", err)
    }
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

func getDfOutputMap() (map[string]string, error) {
    mapping := make(map[string]string)

    cmd := exec.Command("df", "--output=source,target")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        fields := strings.Fields(line)
        if len(fields) < 2 {
            continue
        }
        device := fields[0]
        mount := fields[1]

        if strings.HasPrefix(device, "/dev/") {
            mapping[device] = mount
        }
    }

    return mapping, nil
}

func getMountPoints(dfMap map[string]string) []string {
    var mountPoints []string
    for device := range dfMap {
        if isBlockDevice(device) && !isLoopDevice(device) {
            mountPoints = append(mountPoints, device)
        }
    }
	return mountPoints
}

func generateVolumeArgs(dfMap map[string]string, mountPoints []string) []string {
    var volumeArgs []string
    for _, device := range mountPoints {
        mount := dfMap[device]
        if mount != "/" {
            volumeArgs = append(volumeArgs, "--volume", fmt.Sprintf("%s:%s", mount, mount))
        }
    }
    return volumeArgs
}
