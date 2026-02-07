# Stdin Passthrough for Fraglet Scripts

## Goal

Enable piping data into fraglet scripts:

```bash
cat file.txt | ./foo.java          # finite pipe
echo "hello" | ./foo.py arg1 arg2  # stdin + args together
tail -f log.txt | ./foo.rb         # non-terminating stream
```

This would make fragletc-shebang'd scripts true drop-in replacements for native executables — they accept args, they accept stdin, they produce stdout/stderr with correct exit codes.

---

## Feasibility Verdict: Completely Feasible

There are **no fundamental limitations** preventing this. The Unix stdin pipeline model works cleanly across the Docker boundary. The existing architecture is ~90% there — three specific gaps need closing.

---

## Interface Change: `-c` for Inline Code

The current stdin-as-code-delivery path (`fragletc --vein=python -` / reading from stdin when no file given) is being **replaced** by a `-c` flag, mirroring Python's `python -c 'print("hello")'` convention.

### New interface

```bash
# Inline code via -c (new)
fragletc --vein=python -c 'print("hello")'

# File (unchanged)
fragletc --vein=python script.py

# Shebang (unchanged)
./script.py
```

### What this means for stdin

Stdin is **always** available for data piping. There is no longer a "code via stdin" path competing for the file descriptor. The old `-` positional argument and the bare-stdin fallback are removed.

```bash
# All of these work — stdin is always data
echo "data" | fragletc --vein=python -c 'import sys; print(sys.stdin.read())'
echo "data" | fragletc --vein=python script.py
echo "data" | ./script.py
```

### Implementation of `-c`

**File:** `cli/fragletc.go`

Add a `-c` / `--code` flag:

```go
code := flag.String("c", "", "Program passed in as string")
flag.StringVar(code, "code", "", "Program passed in as string")
```

Logic change:
- If `-c` is set: use its value as the code, stdin is free for data
- If a script file is given: read code from file, stdin is free for data
- If neither: error — no implicit stdin reading for code

Remove the old stdin-code-reading block (`io.ReadAll(os.Stdin)` on lines 129-137) entirely.

---

## Current State: Where Stdin Dies

Tracing `echo "hello" | ./foo.py` through the three layers reveals stdin is dropped at **every** level:

### Layer 1: fragletc CLI (`cli/fragletc.go`)

When reading code from a **file** (shebang case), `os.Stdin` is completely unused — the code comes from `os.ReadFile()`. But the CLI never forwards host stdin into the `RunSpec`. It just... ignores it.

### Layer 2: Docker Runner (`pkg/runner/docker.go`)

The `-i` flag IS present on `docker run` (line 63) — Docker is willing to forward stdin. But `dockerCmd.Stdin` is only set when `spec.Stdin != ""` (a finite string buffer). When no buffer is provided, Go's `exec.Cmd` defaults `Stdin` to `nil`, which maps to `/dev/null`. The container gets nothing.

The local runner (`pkg/runner/local.go`) has the identical gap.

### Layer 3: Entrypoint Executor (`entrypoint/internal/executor/executor.go`)

```go
cmd := exec.Command(cmdPath, cmdArgs...)
cmd.Stdout = os.Stdout  // forwarded
cmd.Stderr = os.Stderr  // forwarded
// cmd.Stdin — NOT SET → reads from /dev/null
```

Stdout and stderr are forwarded. Stdin is silently dropped. Even if Docker somehow delivered stdin to the container, the executor would throw it away.

### Summary

```
Host stdin → [DROPPED by fragletc] → docker run -i → [DROPPED by runner] → container → [DROPPED by executor] → user code
```

Three independent drops. All three must be fixed.

---

## The Fix: Three Layers + Interface Change

### Change 0: Replace stdin-code-reading with `-c` flag

**File:** `cli/fragletc.go`

1. Add `-c` / `--code` flag for inline code as a string
2. Remove the bare stdin fallback (`io.ReadAll(os.Stdin)`)
3. Remove the `-` positional argument meaning "read code from stdin"
4. Error if no code source is provided (no `-c`, no file)
5. **Always** set `StdinReader: os.Stdin` on the RunSpec — since stdin is never consumed for code anymore, it's always available for data

### Change 1: Entrypoint Executor (1 line)

**File:** `entrypoint/internal/executor/executor.go`

Add `cmd.Stdin = os.Stdin` alongside the existing stdout/stderr forwarding:

```go
cmd := exec.Command(cmdPath, cmdArgs...)
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
cmd.Stdin = os.Stdin     // <-- ADD THIS
```

**Risk:** Zero. If no stdin is available, the child reads from the container's stdin (which is `/dev/null` if nothing is piped). If stdin IS available, the child gets it. This is exactly how stdout/stderr already work. Programs that don't read stdin are unaffected.

**This change is required regardless of the other two** — even if Docker somehow delivered stdin, the executor would still eat it.

### Change 2: RunSpec + Runners (add `io.Reader` support)

**Files:** `pkg/runner/runner.go`, `pkg/runner/docker.go`, `pkg/runner/local.go`

The existing `Stdin string` field is a buffered snapshot — it works for the MCP/programmatic use case where you have a known string to inject. For piping, we need a streaming `io.Reader`.

**In `RunSpec`** (`pkg/runner/runner.go`):

```go
type RunSpec struct {
    // ... existing fields ...
    Stdin       string     // Optional stdin input (buffered string)
    StdinReader io.Reader  // Optional stdin stream (takes precedence over Stdin)
    // ...
}
```

**In both runners** (docker.go and local.go), update the stdin wiring:

```go
// Priority: StdinReader > Stdin string > nil
if spec.StdinReader != nil {
    cmd.Stdin = spec.StdinReader
} else if spec.Stdin != "" {
    cmd.Stdin = bytes.NewBufferString(spec.Stdin)
}
```

**Risk:** Low. Existing callers don't set `StdinReader`, so they get identical behavior. The `Stdin string` path is untouched.

### Change 3: fragletc CLI (always forward host stdin)

**File:** `cli/fragletc.go`

Since stdin is never consumed for code anymore (thanks to Change 0), **always** forward `os.Stdin`:

```go
spec := runner.RunSpec{
    Container:   v.Container,
    Env:         envVars,
    Args:        scriptArgs,
    StdinReader: os.Stdin,   // <-- ALWAYS SET (stdin is never used for code)
    Volumes: []runner.VolumeMount{
        {
            HostPath:      tmpFile,
            ContainerPath: defaultFragletPath,
            ReadOnly:      true,
        },
    },
}
```

**Note:** There are two places where `RunSpec` is constructed in fragletc (vein path and direct image path). Both get the same treatment. No conditional logic needed.

**Risk:** Low. When no one is piping data, `os.Stdin` is a terminal — the child process won't block unless it explicitly calls a read on stdin. Programs that don't touch stdin are unaffected.

---

## The `tail -f` Case: Yes, It Works

For `tail -f somefile.txt | ./foo.java`:

1. Host stdin is a pipe from `tail -f` — it never sends EOF until killed
2. `fragletc` forwards `os.Stdin` (the pipe) as `StdinReader` — always, unconditionally
3. Docker `-i` keeps the container's stdin pipe open
4. The entrypoint forwards container stdin to the child process
5. The child process reads stdin in a loop, getting new lines as `tail -f` emits them
6. When the user hits Ctrl+C (or `tail -f` is killed), the pipe closes, the child gets EOF, Docker exits, fragletc exits

The `-i` flag (without `-t`) is correct — no TTY processing, clean byte stream. The entire pipeline is non-buffered at the kernel level (Unix pipes). Line buffering in Docker is configurable but defaults to pass-through for `-i`.

**The only requirement on the user code**: it must read stdin incrementally (line by line or chunk by chunk), not via "read everything then process." Every language supports this:

- **Python:** `for line in sys.stdin:`
- **Ruby:** `ARGF.each_line` or `$stdin.each_line`
- **Java:** `Scanner(System.in).hasNextLine()`
- **C#:** `Console.ReadLine()` in a loop
- **C++:** `while(std::getline(std::cin, line))`
- **C:** `while(fgets(buf, sizeof(buf), stdin))`

---

## Implementation Order

1. **Change 1 first** (executor) — it's inside the container image, so it needs to be built and shipped. It's also the simplest change and independently valuable (fixes stdin for anyone running the entrypoint directly).

2. **Change 2 second** (RunSpec + runners) — adds the plumbing for streaming stdin. Pure library change, no behavior change for existing callers.

3. **Change 0 + Change 3 together** (fragletc CLI) — the `-c` flag and stdin forwarding land at the same time. This is where the feature becomes user-visible.

---

## Testing Strategy

### Candidate Languages

Per the project notes, these veins support arg passing and are good candidates:

- `python`
- `ruby`
- `the-c-programming-language`
- `java`
- `csharp`
- `cpp`

### Test Cases

For each language, create a test script that demonstrates stdin reading:

**Test 1: Basic stdin pipe**
```bash
echo "hello world" | ./stdin_test.py
# Expected output: HELLO WORLD (or similar transformation proving stdin was read)
```

**Test 2: Stdin + args together**
```bash
echo "hello world" | ./stdin_test.py --upper
# Expected output: HELLO WORLD
# Proves args and stdin coexist
```

**Test 3: Multi-line stdin**
```bash
printf "line1\nline2\nline3" | ./stdin_test.py
# Expected output: 3 lines processed
```

**Test 4: Empty stdin (no pipe)**
```bash
./stdin_test.py arg1
# Expected: program runs normally, doesn't block waiting for stdin
# This is critical — programs must handle "no stdin" gracefully
```

**Test 5: Inline code with stdin**
```bash
echo "hello" | fragletc --vein=python -c 'import sys; print(sys.stdin.read().upper())'
# Expected output: HELLO
```

**Test 6: Non-terminating pipe (stretch goal)**
```bash
(echo "line1"; sleep 1; echo "line2"; sleep 1; echo "line3") | ./streaming_test.py
# Expected: each line appears ~1 second apart, proving stream processing
```

### Test Script Templates

Each test script should:
1. Have a fragletc shebang
2. Accept optional args
3. Read stdin if available (non-blocking check where possible)
4. Transform and output the result

**Important caveat for Test 4**: Some languages (C, Java) will block on stdin read if there's no data and stdin is a TTY. The test scripts should ideally check if stdin is a pipe before attempting to read. This is language-specific:

- **Python:** `not sys.stdin.isatty()`
- **Ruby:** `!$stdin.tty?`
- **C:** `!isatty(fileno(stdin))`
- **Java:** `System.console() == null` (rough heuristic)
- **C#:** `Console.IsInputRedirected`
- **C++:** `!isatty(fileno(stdin))`

---

## Edge Cases to Consider

### Binary data through stdin
Docker `-i` without `-t` passes raw bytes. No TTY escaping. Binary data (images, compressed streams) should work fine. Worth a test but not critical for v1.

### Large stdin payloads
Docker and Unix pipes buffer at the kernel level (~64KB default pipe buffer). For large files, the pipeline handles backpressure naturally — writers block when the buffer is full. No special handling needed.

### Stdin EOF propagation
When the pipe writer closes, EOF propagates: pipe → docker → container → entrypoint → child process. Clean and predictable. When the child process exits, fragletc exits, which breaks the pipe (SIGPIPE to the writer). Also clean.

### The `collectStreamingResults` concern
The current `fragletc` CLI uses `r.Run()` which calls `collectStreamingResults()` — this collects ALL stdout/stderr into strings before returning. For the `tail -f` case (non-terminating), this would block forever. The fix: for streaming stdin scenarios, `fragletc` may need to use `RunStreaming()` instead and forward output incrementally. **This is a secondary concern** — the basic finite pipe case works with `Run()`, and the streaming case is a stretch goal.

### Docker `-t` (TTY) flag
Do NOT add `-t`. TTY processing corrupts binary data, adds carriage returns, and interprets control characters. The current `-i` without `-t` is correct for piping.

---

## Proven Results

All changes implemented and tested. Stdin passes through the full Docker boundary for **14 languages**:

### Test Suite: `stdin_passthrough` (entrypoint-level, bash in Alpine container)
- Basic stdin pipe, multi-line stdin, stdin + args, no-stdin (non-blocking), byte counting

### Test Suite: `stdin_languages` (original 6 target languages)
- **Python** — `for line in sys.stdin: print(line.strip().upper())`
- **Ruby** — `$stdin.each_line { |line| puts line.strip.upcase }`
- **C** — `while ((c = getchar()) != EOF) putchar(toupper(c));`
- **C++** — `while (std::getline(std::cin, line)) { ... }`
- **Java** — `Scanner(System.in).hasNextLine()` loop
- **C#** — `while ((line = Console.ReadLine()) != null)`
- **Python stdin+args** — args and stdin coexist correctly

### Test Suite: `stdin_languages_extended` (8 more languages)
- **Bash** — `while IFS= read -r line; do ... done`
- **Go** — `fmt.Scan(&s)` (single token, `fmt` only import available)
- **Rust** — `io::stdin().lock().lines()` with `to_uppercase()`
- **Haskell** — `interact (map toUpper)`
- **Perl** — `while (<STDIN>) { print uc($_); }`
- **Lua** — `for line in io.lines() do print(string.upper(line)) end`
- **Kotlin** — `readLine()` + `uppercase()`
- **Scala** — `scala.io.StdIn.readLine().toUpperCase`

### Testing Approach
The new entrypoint binary was mounted into existing 100hellos containers via `-v`, overriding the baked-in binary. This proves the fix works without rebuilding any container images. The 100hellos update path is to bump the `FRAGLET_VERSION` in `000-base/Dockerfile` once a new release is cut.

### Languages Not Yet Tested (but expected to work)
- **Dart** — Needs `import 'dart:io'` but single-line match injection can't add top-level imports. Template would need region-based injection to support stdin.
- Any language where the fraglet region doesn't include imports and stdin requires an import not in the base template.

---

## Delivery Path to 100hellos

1. Cut a new fraglet release (bump version, build binaries, publish to GitHub releases)
2. Update `FRAGLET_VERSION` in `100hellos/000-base/Dockerfile`
3. Rebuild base images: `make base`
4. Rebuild all language images: `make build`
5. The stdin tests in this repo validate against the published images

No changes needed to any 100hellos language Dockerfiles or fraglet configs — the fix is entirely in the entrypoint binary.

---

## Stretch Goals (Not Required for v1)

1. **Streaming output mode for fragletc** — Use `RunStreaming()` to forward output in real-time, enabling the `tail -f` case fully.

2. **Stdin detection flag** — `fragletc --no-stdin` to explicitly disable stdin forwarding (for edge cases where you don't want the child to inherit stdin).

3. **Bidirectional interactive mode** — `fragletc --interactive` with `-it` Docker flags for REPL-like behavior. Different use case but adjacent.
