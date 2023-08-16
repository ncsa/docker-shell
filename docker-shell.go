package main

import (
    "bytes"
    "fmt"
    "log"
    "os"
    "os/exec"
    "strings"
    "encoding/json"
)

const version = "0.1.0"

func main() {
    for _, arg := range os.Args[1:] {
        switch arg {
            case "--version":
                fmt.Println(version)
                os.Exit(0)
            case "--help":
                printHelp()
                os.Exit(0)
        }
    }

    if len(os.Args) > 1 && os.Args[1] == "docker-cli-plugin-metadata" {
        outputPluginMetadata()
        os.Exit(0)
    }

    if len(os.Args) < 2 {
        printHelp()
        os.Exit(1)
    }

    // Check for --version and --help flags
    switch os.Args[1] {
    case "--version":
        fmt.Println(version)
        os.Exit(0)
    case "--help":
        printHelp()
        os.Exit(0)
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
    cmdArgs := constructDockerRunCommand(volumeArgs, customHostname, os.Args[1:])

    // Run the docker command
    runDockerCommand(cmdArgs)
}

// printHelp displays the help message
func printHelp() {
    fmt.Println(`
Usage: docker-shell [additional docker run options] <DOCKER_IMAGE>

Options:
  --version    Display the version of the script
  --help       Display this help message

This tool is designed to provide an easy way to run a Docker container with system volumes mounted.
`)
}

// constructDockerRunCommand constructs the arguments for the docker run command.
func constructDockerRunCommand(volumeArgs []string, customHostname string, additionalArgs []string) []string {
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
    cmdArgs = append(cmdArgs, additionalArgs...)
    cmdArgs = append(cmdArgs, "/bin/bash")
    return cmdArgs
}

// runDockerCommand executes the docker command with given arguments.
func runDockerCommand(args []string) {
    cmd := exec.Command("docker", args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin
    if err := cmd.Run(); err != nil {
        log.Fatalf("Failed to run docker command: %v", err)
    }
}

// outputPluginMetadata outputs the metadata for the Docker CLI plugin.
func outputPluginMetadata() {
    metadata := struct {
        SchemaVersion    string `json:"SchemaVersion"`
        Vendor           string `json:"Vendor"`
        Version          string `json:"Version"`
        ShortDescription string `json:"ShortDescription"`
        URL              string `json:"URL"`
    }{
        SchemaVersion:   "0.1.0",
        Vendor:          "Matthew Krafczyk",
        Version:         version,
        ShortDescription: "A utility designed to mimic the `singularity shell` command.",
        URL:              "https://github.com/ncsa/docker-shell",
    }

    jsonOutput, err := json.MarshallIndent(metadata, "", "    ")
    if err != nil {
        log.Fatalf("Failed to generate plugin metadata: %v", err)
    }
    fmt.Println(string(jsonOutput))
}

// isBlockDevice checks if a given device is a block device.
func isBlockDevice(device string) bool {
    cmd := exec.Command("ls", "-l", device)
    var out bytes.Buffer
    cmd.Stdout = &out
    if err := cmd.Run(); err != nil {
        return false
    }
    return strings.HasPrefix(out.String(), "b")
}

// isLoopDevice checks if a given device is a loop device.
func isLoopDevice(device string) bool {
    cmd := exec.Command("losetup", device)
    return cmd.Run() == nil
}

// getDfOutputMap creates a mapping of device names to their corresponding mount points.
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

// getMountPoints retrieves the mount points of non-loopback block devices.
func getMountPoints(dfMap map[string]string) []string {
    var mountPoints []string
    for device := range dfMap {
        if isBlockDevice(device) && !isLoopDevice(device) {
            mountPoints = append(mountPoints, device)
        }
    }
	return mountPoints
}

// generateVolumeArgs generates the volume arguments for the docker command.
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
