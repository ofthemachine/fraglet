# fraglet-entrypoint

The `fraglet-entrypoint` binary orchestrates fraglet injection and execution within containers.

## Configuration

Configuration is provided via `fraglet.yaml` or `fraglet.yml` (default: `/fraglet.yaml`) or the `FRAGLET_CONFIG` environment variable.

### Key Fields

- **`fragletTempPath`**: Temporary location where fraglet code is written before injection
- **`injection.codePath`**: Target file where fraglet code is injected
- **`injection.match`**: String marker that identifies the injection point (line replacement)
- **`guide`**: Path to guide markdown file (served via `guide` command)
- **`essence`**: Path to essence markdown file (served via `essence` command)
- **`execution.path`**: Command/path to execute after injection
- **`execution.makeExecutable`**: Whether to make the execution path executable (default: `true`)
- **`execution.argv`**: Skip file injection entirely and exec `execution.path` directly against a literal argv (see "Argv mode" below)

### Commands

- **`usage`**: Displays dynamic container usage documentation (generated from config)
- **`guide`**: Displays static authoring guide for writing fraglets (from `guide.md` file)
- **`essence`**: Displays short capability summary for the active mode (from essence markdown file)

### Execution Notes

- If `execution.path` is a file path (not an interpreter command), `makeExecutable` must be `true` or execution will fail
- If `execution.path` is omitted, the entrypoint passes through command-line arguments
- Guide and essence files are checked first at the configured absolute path, then in the code directory

### Argv mode

For containers with no interpreter at all — `FROM scratch`, a single static binary plus `fraglet-entrypoint` — there is no scaffold file to inject into and no shell to run one in. `execution.argv: true` skips injection entirely:

- **No fraglet mounted** (a plain `docker run image <args>`): args pass straight through to `execution.path`, unmodified.
- **A fraglet is mounted** (a `fragletc` invocation): the body must be exactly one command line (fraglet-meta header stripped). It's expanded to a literal argv with `pkg/fraglet.ExpandArgv` — whitespace words, `'single'`/`"double"` quoting, `$VAR`/`${VAR}`/`$@`/`$1..$9` — and exec'd with no shell involved. The body's first word must name `execution.path` (by base name or full path); a mismatch is a hard error rather than a confusing silent exec.
- Deliberately unsupported: pipes, redirects, `;`/`&&`/`||`, command substitution, globbing — anything that needs a real shell. Use script/injection mode instead if you need those.

See `containers/meme` (sibling `ofthemachine/containers` repo) for a working example.

See `fraglet.yaml` for a fully documented example.
