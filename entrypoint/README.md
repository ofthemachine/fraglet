# fraglet-entrypoint

The `fraglet-entrypoint` binary orchestrates fraglet injection and execution within containers.

## Configuration

Configuration is provided via `fraglet-entrypoint.yaml` (default: `/fraglet-entrypoint.yaml`) or the `FRAGLET_CONFIG` environment variable.

### Key Fields

- **`fragletTempPath`**: Temporary location where fraglet code is written before injection
- **`injection.codePath`**: Target file where fraglet code is injected
- **`injection.match`**: String marker that identifies the injection point (line replacement)
- **`agentHelp`**: Path to agent-help markdown file (served via `agent-help` command)
- **`howTo`**: Path to how-to markdown file (served via `how-to` command)
- **`execution.path`**: Command/path to execute after injection
- **`execution.makeExecutable`**: Whether to make the execution path executable (default: `true`)

### Execution Notes

- If `execution.path` is a file path (not an interpreter command), `makeExecutable` must be `true` or execution will fail
- If `execution.path` is omitted, the entrypoint passes through command-line arguments
- Documentation files are checked first at the configured absolute path, then in the code directory

See `files/fraglet-entrypoint.yaml` for a fully documented example.

