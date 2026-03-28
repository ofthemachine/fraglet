# fraglet-entrypoint

The `fraglet-entrypoint` binary orchestrates fraglet injection and execution within containers.

## Configuration

Configuration is provided via `fraglet.yaml` or `fraglet.yml` (default: `/fraglet.yaml`) or the `FRAGLET_CONFIG` environment variable.

### Key Fields

- **`fragletTempPath`**: Temporary location where fraglet code is written before injection
- **`injection.codePath`**: Target file where fraglet code is injected
- **`injection.match`**: String marker that identifies the injection point (line replacement)
- **`guide`**: Path to guide markdown file (served via `guide` command)
- **`execution.path`**: Command/path to execute after injection
- **`execution.makeExecutable`**: Whether to make the execution path executable (default: `true`)

### Commands

- **`usage`**: Displays dynamic container usage documentation (generated from config)
- **`guide`**: Displays static authoring guide for writing fraglets (from `guide.md` file)

### Execution Notes

- If `execution.path` is a file path (not an interpreter command), `makeExecutable` must be `true` or execution will fail
- If `execution.path` is omitted, the entrypoint passes through command-line arguments
- Guide files are checked first at the configured absolute path, then in the code directory

See `fraglet.yaml` for a fully documented example.

