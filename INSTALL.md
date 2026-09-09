# Installing fragletc

**fragletc** runs code fragments in isolated containers across 90+ languages. It ships as a single binary.

## Prerequisites

**Docker** is required. fragletc uses Docker to run code in isolated containers.

| Platform | Install Docker |
|----------|---------------|
| macOS | [Docker Desktop for Mac](https://docs.docker.com/desktop/install/mac-install/) |
| Linux | [Docker Engine](https://docs.docker.com/engine/install/) |
| Windows | [Docker Desktop for Windows](https://docs.docker.com/desktop/install/windows-install/) (WSL 2 backend) |

Verify Docker is working:

```sh
docker run --rm hello-world
```

## Quick Install

```sh
curl -fsSL https://raw.githubusercontent.com/ofthemachine/fraglet/main/install.sh | sh
```

This downloads the latest release for your platform, verifies the checksum, and installs to `~/.local/bin`.

To install to a different location:

```sh
FRAGLETC_INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/ofthemachine/fraglet/main/install.sh | sh
```

## Verify

```sh
fragletc --help
fragletc --vein=python -c 'print("hello from fraglet")'
```

## Manual Install

Download the binary for your platform from [GitHub Releases](https://github.com/ofthemachine/fraglet/releases), then:

```sh
chmod +x fragletc-*
mv fragletc-* /usr/local/bin/fragletc
```

### Build from Source

Requires Go 1.24+:

```sh
git clone https://github.com/ofthemachine/fraglet.git
cd fraglet
make install
```

Or directly:

```sh
go install github.com/ofthemachine/fraglet/cmd/fragletc@latest
```

## Troubleshooting

### `command not found: fragletc`

The install directory is not in your PATH. Add it:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

Add this line to your `~/.zshrc`, `~/.bashrc`, or equivalent.

### Docker connection errors

Ensure Docker is running:

```sh
docker info
```

On Linux, your user may need to be in the `docker` group:

```sh
sudo usermod -aG docker $USER
# Then log out and back in
```

### Container image pull failures

If you're behind a firewall or proxy, Docker may not be able to pull images. Pre-pull a vein to test:

```sh
fragletc refresh python
```
