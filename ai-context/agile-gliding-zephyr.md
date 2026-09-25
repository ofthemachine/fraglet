# Fraglet Parameterization — Design Plan

## Context

Operon Phase 3 introduces fraglet execution (`operon run`). Phase 5's solver executes fraglets with varying inputs for memoization. Both require fraglets to accept dynamic parameters. Today fraglets have no parameterization — code is static, only CLI args and stdin (neither available via MCP) carry data.

String templating (Go templates, mustache) is a dead end — escaping across 70+ languages is intractable. Instead: **environment variable injection with encoding annotations**, leveraging the fact that fraglet owns the entire execution path from CLI through entrypoint.

## Decision: Env Var Injection with Bare Names

### Two layers

| Layer | Prefix | Where | Purpose |
|-------|--------|-------|---------|
| **Transport** | `FRAGLET_PARAM_` | Runner → Docker `-e` | Allow-list: only env vars with this prefix pass into the container |
| **Program-facing** | *(bare, case-preserved)* | Inside container | Entrypoint strips `FRAGLET_PARAM_` prefix, decodes value, injects bare env var |

### Entrypoint coercion — dumb and schema-free

The entrypoint does NOT read fraglet-meta or the fraglet source code for param coercion. It operates blindly:

1. Scan container env for all `FRAGLET_PARAM_*` vars
2. For each: strip `FRAGLET_PARAM_` prefix → bare env var name (case-preserved)
3. Decode value (parse `type:value`, apply decoding)
4. Check no-shadow: if bare env var already exists, skip (existing wins)
5. Set bare env var, unset `FRAGLET_PARAM_*` var
6. Proceed to injection + execution as normal

This means ANY `FRAGLET_PARAM_*` env var passed to the container gets coerced. The entrypoint doesn't validate, doesn't check declarations, doesn't parse code. It's a pure transport-to-runtime bridge.

### Param declarations via fraglet-meta — documentation, aliases, preflight

Fraglet source files declare accepted params using `param=` tokens in `fraglet-meta:` annotations. These declarations are **not used by the entrypoint** — they serve as:
1. **Aliases**: User-facing param names that map to env vars (resolved host-side by the runner)
2. **Documentation**: What params a fraglet accepts, surfaced via `--fraglet-help`
3. **Preflight validation** (future): Runner can validate params before launching the container
4. **Tooling**: MCP/CLI can use declarations for help text, auto-complete, etc.

#### Token syntax: `param=<alias>[:<modifier>...]`

```
param=<alias>                           — declares a param, env var = ALIAS (uppercased)
param=<alias>:<modifier>                — with one modifier
param=<alias>:<modifier>:<modifier>     — with multiple modifiers
```

- `=` separates the `param` keyword from the alias (user-facing name)
- `:` separates modifiers
- Modifiers are bare words (`required`, `optional`) or `key=value` pairs (`default=metric`, `envvar=HURL_VARIABLE_host`)
- Default env var name: alias uppercased. Override with `envvar=` modifier.

#### Alias-to-envvar mapping

By default, `param=city` maps to env var `CITY` (uppercased). For languages with non-standard env var conventions, use the `envvar=` modifier:

```
param=city                              → --param city=london → CITY=london
param=city:required                     → same, must be provided
param=host:envvar=HURL_VARIABLE_host    → --param host=localhost → HURL_VARIABLE_host=localhost
param=port:envvar=HURL_VARIABLE_port:default=8080
```

The alias is what users type with `--param`. The envvar is what the program reads. The mapping is resolved **host-side by the runner** — the entrypoint never sees aliases.

**Resolution flow:**
1. User passes `--param host=localhost`
2. Runner reads fraglet-meta, finds `param=host:envvar=HURL_VARIABLE_host`
3. Runner creates transport: `FRAGLET_PARAM_HURL_VARIABLE_host=raw:localhost`
4. Entrypoint strips prefix → `HURL_VARIABLE_host=localhost`
5. Canonical form (ledger/memo): `HURL_VARIABLE_host=raw:localhost` (real envvar, not alias)

**Examples:**

```python
# fraglet-meta: determinism:deterministic
# fraglet-meta: param=city:required
# fraglet-meta: param=date:required
# fraglet-meta: param=units:optional:default=metric
import os
city = os.environ['CITY']
date = os.environ['DATE']
units = os.environ.get('UNITS', 'metric')
print(f"Weather in {city} on {date} ({units})")
```

**Hurl (language-specific env var convention):**
```hurl
# fraglet-meta: param=host:envvar=HURL_VARIABLE_host:required
# fraglet-meta: param=port:envvar=HURL_VARIABLE_port:default=8080
GET http://{{host}}:{{port}}/health
HTTP 200
```
User passes `--param host=localhost`. Clean, domain-oriented. The `HURL_VARIABLE_` prefix is an implementation detail of the fraglet.

**Bash:**
```bash
#!/usr/bin/env -S fragletc --vein=bash
# fraglet-meta: param=host param=port
curl -s "http://$HOST:$PORT/health"
```

#### `--fraglet-help` flag

The CLI provides `--fraglet-help` to display param declarations from a fraglet's metadata:

```bash
$ fragletc --fraglet-help test.hurl
Parameters for test.hurl (vein: hurl):
  host     (required)                    → HURL_VARIABLE_host
  port     (optional, default: 8080)     → HURL_VARIABLE_port
```

- **Host-side only** — reads the fraglet source, parses `param=` tokens, no container needed
- **Does not collide with `--help`** — fragletc's own `--help` shows fragletc usage; `--fraglet-help` shows the fraglet's param contract
- **MCP equivalent**: param declarations included in `language_help` tool output and in run tool description when code contains fraglet-meta

#### Token parsing grammar

```
fraglet-meta token → annotation | param_decl
annotation         → key:value                    (e.g., determinism:deterministic)
param_decl         → param=ALIAS[:MODIFIER]*       (e.g., param=city:required:default=london)

MODIFIER           → bare_word                     (e.g., required, optional)
                   | key=value                     (e.g., default=metric, envvar=HURL_VARIABLE_host)
```

Disambiguation: tokens starting with `param=` are param declarations; all others are `key:value` annotations. No ambiguity because annotations use `:` as their primary separator and params use `=`.

### Multi-line fraglet-meta

Multiple `# fraglet-meta:` lines are collected and merged. Params sorted lexically by env var name for deterministic ordering.

```python
# fraglet-meta: determinism:deterministic
# fraglet-meta: param=CITY:required
# fraglet-meta: param=DATE:required
# fraglet-meta: param=UNITS:optional:default=metric
```

**Parsing**: Scan all lines for the sentinel `fraglet-meta:` (language-agnostic — works with `#`, `//`, `--`, `%`, etc.). Collect all tokens from all matching lines.

**Serialization** (save.go): Annotations on one `# fraglet-meta:` line, params on subsequent lines, both lexically sorted.

### No-Shadow Rule

The entrypoint checks if the bare env var already exists before injecting. If it does, skip — existing value wins. This naturally protects system env vars (`PATH`, `HOME`, etc.) without maintaining a blocklist.

### Encoding Annotations

Param values carry an optional type prefix: `type:value`. Default type is `raw`.

```
--param CITY=london              → CITY=london           (raw, default)
--param CITY=raw:london          → CITY=london           (explicit raw)
--param PAYLOAD=b64:SGVsbG8=     → PAYLOAD=Hello         (base64 decoded)
--param DATA=cb64:H4sIAAAA...    → DATA=<decompressed>   (zlib+base64 decoded)
```

Transport layer always carries the encoded form (`FRAGLET_PARAM_CITY=raw:london`).
The entrypoint decodes before injecting — programs always see clean decoded values.
The **encoded form** is what's stored in the ledger INPUTS field and used for memoization.

Supported types:
- `raw` (default) — pass-through string
- `b64` — base64 decode to string
- `cb64` — base64 decode → zlib decompress to string

### Env var size limits

Each Docker `-e` flag is an `execve` argument, subject to `MAX_ARG_STRLEN` = **128KB per argument** on Linux. This limits per-param transport size:

| Encoding | Max value per param |
|----------|-------------------|
| `raw` | ~128KB |
| `b64` | ~96KB source data (33% b64 expansion) |
| `cb64` | Depends on compressibility; compressed+b64 must be < 128KB |

**Inside the container, there is no limit.** `os.Setenv()` in an already-running process is not subject to `execve` constraints. So `cb64` values that decompress to >128KB work fine — only the wire form must fit.

Params are for string values. Not files, not large blobs.

### Deterministic Hashing (for operon memoization)

```
memo_key = sha256(fragl_sha + "\n" + sorted encoded "key=type:value" lines)
```
The encoded form is the canonical representation. Same code + same encoded params = cache hit.

### Secrets — separate concept, not params

Secrets (credentials, API keys, passwords) are **not params**. They use a different declaration, different transport, and different injection mechanism.

```python
# fraglet-meta: secret=DB_PASSWORD
# fraglet-meta: secret=API_KEY
# fraglet-meta: param=CITY:required
```

**Why separate:**
- Env vars are visible in `docker inspect`, `/proc/*/environ`, `ps aux`, shell history
- Params use env var transport — fundamentally not secret-safe
- Secrets need a different channel: mounted files, not env vars

**Implemented (fragletc, 2026-09-24):** `#: secret=NAME:d=<description>`, one per header line.
- The value comes from the caller's env var `NAME` (wrap with `op run`, `security find-generic-password`, etc. for vaults); a missing one exits 2 host-side.
- Delivery follows Docker Compose's environment-sourced secrets: `docker create --rm`, then `docker cp -` of an in-memory tar putting a 0444 file at `/run/fraglet/secrets/NAME`, then `docker start -a`. `NAME_FILE` holds the path. The value is never in argv, `docker inspect`, the container env, a host file, or the receipt; SIGINT/SIGTERM force-remove the container.
- Declarations are in the script, so `procedure_hash` covers them and the run stays conformant (memo key intact). No source or transport options: those are caller concerns, not tool contract.

---

## Full Lifecycle Examples

### 1. Parameterized Fraglet Creation

**Simple — default mapping (alias uppercased → env var):**
```python
# fraglet-meta: param=city param=units
import os
city = os.environ['CITY']
units = os.environ.get('UNITS', 'metric')
print(f"Weather in {city} ({units})")
```

**Multi-line with modifiers:**
```python
# fraglet-meta: determinism:deterministic
# fraglet-meta: param=city:required
# fraglet-meta: param=date:required
# fraglet-meta: param=units:optional:default=metric
import os
print(f"Weather in {os.environ['CITY']} on {os.environ['DATE']} ({os.environ.get('UNITS', 'metric')})")
```

**Hurl (envvar mapping):**
```hurl
# fraglet-meta: param=host:envvar=HURL_VARIABLE_host:required
# fraglet-meta: param=port:envvar=HURL_VARIABLE_port:default=8080
GET http://{{host}}:{{port}}/health
HTTP 200
```

**Go (different comment syntax, same sentinel):**
```go
// fraglet-meta: param=name
package main
import ("fmt"; "os")
func main() { fmt.Println("Hello, " + os.Getenv("NAME")) }
```

### 2. CLI Invocation

**Basic (aliases, default encoding):**
```bash
fragletc --vein=python -p city=london -p units=metric script.py
```
- Runner resolves: `city` → `CITY` (default uppercase), `units` → `UNITS`
- Transport env: `FRAGLET_PARAM_CITY=raw:london`, `FRAGLET_PARAM_UNITS=raw:metric`
- Container env (after entrypoint): `CITY=london`, `UNITS=metric`

**With explicit encoding:**
```bash
fragletc --vein=python -p city=raw:london -p payload=b64:eyJrZXkiOiJ2YWx1ZSJ9 script.py
```
- Transport: `FRAGLET_PARAM_CITY=raw:london`, `FRAGLET_PARAM_PAYLOAD=b64:eyJrZXkiOiJ2YWx1ZSJ9`
- Container: `CITY=london`, `PAYLOAD={"key":"value"}`

**Hurl with aliased params:**
```bash
fragletc --vein=hurl -p host=localhost -p port=8080 test.hurl
```
- Runner reads fraglet-meta: `host` → `HURL_VARIABLE_host`, `port` → `HURL_VARIABLE_port`
- Transport: `FRAGLET_PARAM_HURL_VARIABLE_host=raw:localhost`, `FRAGLET_PARAM_HURL_VARIABLE_port=raw:8080`
- Container: `HURL_VARIABLE_host=localhost`, `HURL_VARIABLE_port=8080`
- Hurl accesses: `{{host}}`, `{{port}}`
- User never types `HURL_VARIABLE_` — the alias hides the plumbing

**Fraglet help:**
```bash
$ fragletc --fraglet-help test.hurl
Parameters for test.hurl (vein: hurl):
  host     (required)                    → HURL_VARIABLE_host
  port     (optional, default: 8080)     → HURL_VARIABLE_port
```

**Shebang execution:**
```bash
chmod +x weather.py
./weather.py -p city=london -p date=2024-01-15
```

### 3. MCP Invocation

**Basic MCP call (uses aliases):**
```json
{
  "tool": "run",
  "arguments": {
    "lang": "python",
    "code": "# fraglet-meta: param=city param=units\nimport os\nprint(os.environ['CITY'], os.environ.get('UNITS', 'metric'))",
    "params": {
      "city": "london",
      "units": "metric"
    }
  }
}
```
- `params` keys are aliases — runner resolves to env var names
- `params` values default to `raw:` encoding

**MCP with explicit encoding:**
```json
{
  "tool": "run",
  "arguments": {
    "lang": "python",
    "code": "# fraglet-meta: param=payload\nimport os, json\nprint(json.loads(os.environ['PAYLOAD']))",
    "params": {
      "payload": "b64:eyJrZXkiOiJ2YWx1ZSJ9"
    }
  }
}
```

### 4. Encoding Type Summary Across Phases

| Encoding | User provides (alias) | Transport (`FRAGLET_PARAM_*`) | Container (bare) | Canonical (ledger/memo) |
|----------|----------------------|-------------------------------|-------------------|-------------------------|
| `raw` (default) | `city=london` | `FRAGLET_PARAM_CITY=raw:london` | `CITY=london` | `CITY=raw:london` |
| `raw` (explicit) | `city=raw:london` | `FRAGLET_PARAM_CITY=raw:london` | `CITY=london` | `CITY=raw:london` |
| `b64` | `payload=b64:SGVsbG8=` | `FRAGLET_PARAM_PAYLOAD=b64:SGVsbG8=` | `PAYLOAD=Hello` | `PAYLOAD=b64:SGVsbG8=` |
| `cb64` | `data=cb64:H4sI...` | `FRAGLET_PARAM_DATA=cb64:H4sI...` | `DATA=<decompressed>` | `DATA=cb64:H4sI...` |

---

## Implementation — Fraglet Repo First

Work lands in the fraglet repo first, then operon Phase 3 ports the parameterized runner.

### Step 1: Param types and encoding (`fraglet/pkg/fraglet/param.go` — new)

```go
type Param struct {
    EnvVar   string  // resolved env var name: "CITY", "HURL_VARIABLE_host"
    Encoding string  // "raw", "b64", "cb64"
    Value    string  // encoded value as provided: "london" or "SGVsbG8="
}

// ParseParam parses "city=london" or "city=b64:SGVsbG8=" into a Param.
// The name is treated as the env var (uppercased by default).
// Use ResolveAlias to apply fraglet-meta envvar= mappings.
func ParseParam(s string) (Param, error)

// Decode returns the decoded value (applies b64/cb64 decoding)
func (p Param) Decode() (string, error)

// TransportEnvName returns "FRAGLET_PARAM_" + EnvVar
func (p Param) TransportEnvName() string  // "CITY" → "FRAGLET_PARAM_CITY"

// TransportEnvValue returns the encoded form: "raw:london"
func (p Param) TransportEnvValue() string

// Canonical returns the deterministic form for hashing/ledger: "CITY=raw:london"
// Always uses the resolved env var name, not the alias.
func (p Param) Canonical() string
```

Also:
```go
type Params []Param

// ToTransportEnv returns sorted "FRAGLET_PARAM_X=type:value" pairs for docker -e
func (ps Params) ToTransportEnv() []string

// ToCanonical returns sorted "key=type:value" pairs for hashing/ledger
func (ps Params) ToCanonical() []string

// ResolveAliases applies fraglet-meta param declarations to map aliases → env var names.
// Called host-side by the runner before building transport env vars.
func (ps Params) ResolveAliases(decls []ParamDecl) (Params, error)

// reserved param names that cannot be used
var reserved = map[string]bool{"CONFIG": true}
```

Tests: `pkg/fraglet/param_test.go`
- Parse round-trips for raw, b64, cb64
- Default alias resolution: `city` → `CITY` (uppercased)
- Explicit envvar mapping: alias `host` + `envvar=HURL_VARIABLE_host` → `HURL_VARIABLE_host`
- Canonical form uses resolved env var name, not alias
- Canonical form is deterministic regardless of input order
- TransportEnvName prefixes correctly
- Reserved names rejected
- Empty params produce no env vars
- ResolveAliases errors on unknown alias (param not declared in fraglet-meta)

### Step 2: Entrypoint param coercion (`fraglet/entrypoint/`)

The entrypoint does NOT read fraglet-meta. It coerces all `FRAGLET_PARAM_*` env vars blindly.

**Flow:**
1. Scan `os.Environ()` for all vars starting with `FRAGLET_PARAM_`
2. For each: strip prefix → bare name (case-preserved)
3. Parse value as `type:value` (default `raw`), decode
4. Check no-shadow: if bare name already in env → skip
5. `os.Setenv(bareName, decodedValue)`
6. `os.Unsetenv(transportName)`
7. Proceed to injection + execution as normal

**Files:**
- `entrypoint/cmd/main.go` — add param coercion step before injection
- `entrypoint/internal/params/coerce.go` — new: scan env, strip prefix, decode, no-shadow, inject
- `entrypoint/internal/params/coerce_test.go` — new: unit tests

### Step 3: fraglet-meta param parsing (`fraglet/pkg/fraglet/meta.go` — new)

Host-side parsing of `param=` tokens from fraglet-meta. Used by runner for alias resolution, `--fraglet-help`, preflight validation (future), and save.go for artifact serialization.

```go
// ParamDecl represents a declared parameter from fraglet-meta
type ParamDecl struct {
    Alias     string            // user-facing name: "city", "host"
    EnvVar    string            // resolved env var: "CITY", "HURL_VARIABLE_host"
    Modifiers map[string]string // "required" → "", "default" → "metric"
}

// ParseParamDecls extracts param= tokens from code string.
// Scans all lines for "fraglet-meta:" sentinel, collects param= tokens.
// For each: alias = token value, envvar = envvar= modifier or alias uppercased.
func ParseParamDecls(code string) []ParamDecl
```

Tests: `pkg/fraglet/meta_test.go`
- Single-line and multi-line fraglet-meta
- Mixed annotations + params (annotations ignored, params extracted)
- Various comment prefixes (`#`, `//`, `--`, `%`)
- Default envvar: `param=city` → Alias="city", EnvVar="CITY"
- Explicit envvar: `param=host:envvar=HURL_VARIABLE_host` → Alias="host", EnvVar="HURL_VARIABLE_host"
- Modifier parsing: `param=city:required:default=london`
- Deduplication of repeated param declarations

### Step 4: Runner integration (`fraglet/pkg/runner/`)

Modify `RunSpec` to accept params:

**`pkg/runner/runner.go`:**
```go
type RunSpec struct {
    // ... existing fields ...
    Params fraglet.Params  // typed parameters, injected as env vars
}
```

**`pkg/runner/docker.go`:**
In `RunStreaming`, after building base env vars, append param transport env vars:
```go
if len(spec.Params) > 0 {
    allEnv = append(allEnv, spec.Params.ToTransportEnv()...)
}
```

### Step 5: CLI + MCP integration (`fraglet/cli/`, `fraglet/mcp/tools/`)

**CLI (`cli/fragletc.go`):**
Add `--param` / `-p` repeatable flag and `--fraglet-help`:
```bash
fragletc --vein=python -p city=london -p date=2024-01-15 script.py
fragletc --fraglet-help script.py   # prints param declarations from fraglet-meta
```

`--fraglet-help` flow:
1. Read the fraglet source file
2. Parse fraglet-meta with `ParseParamDecls()`
3. Print formatted table: alias, modifiers, resolved env var name
4. Exit (no container launched)

**MCP run tool (`mcp/tools/run.go`):**
Add `Params` to `RunInput`:
```go
type RunInput struct {
    // ... existing fields ...
    Params map[string]string `json:"params,omitempty" jsonschema:"alias=value or alias=type:value parameters injected as env vars"`
}
```

MCP run flow with params:
1. Parse `Params` map into `fraglet.Params` (keys are aliases)
2. Parse code for `ParamDecls` via `ParseParamDecls()`
3. Call `params.ResolveAliases(decls)` to map aliases → env var names
4. Pass resolved `Params` to `RunSpec`

### Step 6: Integration tests

- `run_with_params/` — `-p city=london` → fraglet-meta `param=city`, program reads `$CITY`, prints "london"
- `run_params_alias/` — `-p host=localhost` → fraglet-meta `param=host:envvar=HURL_VARIABLE_host`, program reads `$HURL_VARIABLE_host`
- `run_params_no_shadow/` — container pre-sets `CITY`, param injection skips it
- `run_params_b64/` — `-p msg=b64:SGVsbG8=` → `$MSG` = "Hello"
- `run_params_deterministic/` — same params in different order produce same canonical form
- `run_params_transport_cleaned/` — `FRAGLET_PARAM_*` vars unset after coercion
- `run_params_undeclared_works/` — params without fraglet-meta declarations still coerce (entrypoint is schema-free)
- `run_fraglet_help/` — `--fraglet-help` prints param table with aliases, env vars, modifiers

---

## Operon Phase 3 Porting

Once fraglet has parameterized runner, operon Phase 3 ports it:

1. Port `pkg/runner/` from fraglet (now includes Params support)
2. Port `pkg/fraglet/param.go` types into `pkg/runner/params.go`
3. `operon run --param CITY=london` → params flow into RunSpec → container → bare `CITY=london`
4. Params recorded as typed inputs in FRAGL/FRAGR ledger entries: `CITY=raw:london DATE=raw:2024-01-15`
5. Memoization key computed from `sha256(fragl_sha + sorted canonical params)`

The entrypoint binary lives inside the container images (100hellos). The coercion logic from Step 2 becomes available once the containers are rebuilt with the updated entrypoint.

---

## Key Files Summary

| File | Repo | Action |
|------|------|--------|
| `pkg/fraglet/param.go` | fraglet | New: Param type, Parse, Decode, TransportEnvName, Canonical |
| `pkg/fraglet/param_test.go` | fraglet | New: unit tests |
| `pkg/fraglet/meta.go` | fraglet | New: ParseParamDecls (host-side fraglet-meta parsing) |
| `pkg/fraglet/meta_test.go` | fraglet | New: unit tests |
| `entrypoint/internal/params/coerce.go` | fraglet | New: scan FRAGLET_PARAM_*, decode, no-shadow, inject |
| `entrypoint/internal/params/coerce_test.go` | fraglet | New: unit tests |
| `entrypoint/cmd/main.go` | fraglet | Modified: add param coercion before injection |
| `pkg/runner/runner.go` | fraglet | Modified: Params field on RunSpec |
| `pkg/runner/docker.go` | fraglet | Modified: append Params.ToTransportEnv() to docker -e |
| `cli/fragletc.go` | fraglet | Modified: --param / -p repeatable flag |
| `mcp/tools/run.go` | fraglet | Modified: Params field in RunInput |

## Verification

1. `make test` in fraglet — all existing tests still pass
2. `make test-entrypoint` — param coercion tests pass (no-shadow, encoding, FRAGLET_PARAM_* stripping)
3. New CLI integration tests: alias resolution, envvar mapping, no-shadow, b64 encoding, transport cleanup, `--fraglet-help`
4. Manual MCP test: `"params": {"city": "london"}` → alias resolved → `$CITY` reads correctly inside container
5. Manual MCP test (Hurl): `"params": {"host": "localhost"}` → `$HURL_VARIABLE_host` reads correctly
6. After operon port: `make test-integration` — run_with_params test passes
7. Ledger entries contain typed inputs using resolved env var names in canonical form
