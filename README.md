# docker-shell
A seamless Docker container launcher that auto-configures mounts, user credentials, and working directory, making it easier for developers and system administrators to quickly dive into containerized environments without the hassles of manual setup.

# Docker Shell

**Docker Shell** is a simple utility designed to launch a Docker container with the current user's credentials, the current working directory as the container's working directory, and mounts for the system's available block devices. It is specifically tailored for providing a quick and seamless experience for developers and system administrators.

## Summary

Docker Shell makes it easier for developers and administrators to launch Docker containers by auto-configuring various parameters. Key features include:

- Automatically mapping block devices from the host to the container.
- Setting the current user and group for the container.
- Preserving the current working directory inside the container.
- Providing a custom hostname that encapsulates both the host's hostname and the container's name.

## Examples

Launch a Docker container using the ubuntu image:

```bash
$ docker-shell ubuntu
```

Launch a Docker container with additional Docker arguments:

```bash
$ docker-shell -p 8080:80 nginx
```

## Security Warning

Using **`docker-shell`** allows for quick configuration and launching of Docker containers. However, there are a few points to consider:

- **Mounts**: The utility will automatically mount block devices available on the host to the container. Ensure you're aware of the mounts being made, as they might expose sensitive data to the container.

- **User Credentials**: The utility will run the Docker container with the same user and group ID as the current user. This is a safer alternative to running as root, but you should still exercise caution and ensure that the Docker image you're using is trusted.

## Building and Installing from Source

### Prerequisites:

- Ensure you have Go installed on your machine. If not, download and install it from here.

### Steps:

1. Clone the repository:

```bash
$ git clone https://github.com/your-username/docker-shell.git
$ cd docker-shell
```

2. Build the project:

```bash
$ go build
```

3. Install the binary:

```bash
$ sudo make install
```

If you want to install the binary to a custom location, use:

```bash
$ sudo make install PREFIX=/path/to/your/bin
```
