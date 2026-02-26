# Installing fragletc

**fragletc** runs code fragments in isolated containers across 90+ languages. It ships as a single binary with built-in MCP server support for AI tool integration.

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

## MCP Server Setup

fragletc includes a built-in MCP (Model Context Protocol) server. Run it with:

```sh
fragletc mcp
```

This starts the server over stdio, compatible with any MCP client.

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "fraglet": {
      "command": "fragletc",
      "args": ["mcp"]
    }
  }
}
```

Restart Claude Desktop after saving.

### Cursor

**One-click install:**

[Install fraglet in Cursor](cursor://anysphere.cursor-deeplink/mcp/install?name=fraglet&config=eyJjb21tYW5kIjoiZnJhZ2xldGMiLCJhcmdzIjpbIm1jcCJdfQ==)

**Manual setup** — add to `.cursor/mcp.json` (project-level) or `~/.cursor/mcp.json` (global):

```json
{
  "mcpServers": {
    "fraglet": {
      "command": "fragletc",
      "args": ["mcp"]
    }
  }
}
```

### MCP Tools

The MCP server exposes two tools:

| Tool | Description |
|------|-------------|
| `run` | Execute code snippets in any supported language container |
| `language_help` | Get the authoring guide for a language (syntax, libraries, patterns) |

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
go install github.com/ofthemachine/fraglet/cli@latest
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

### MCP server not connecting

1. Verify the binary path is absolute or in PATH
2. Check the JSON syntax in your MCP config file
3. Restart your MCP client (Claude Desktop / Cursor) after config changes
4. Test the server manually: `echo '{}' | fragletc mcp` should produce JSON-RPC output
