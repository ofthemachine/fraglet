# Fraglet testing

This document describes the testing contract for fraglet support and how it is validated across the fraglet repo and 100hellos.

## Verification contract

A language has **complete fraglet support** when all of the following hold:

1. **Default execution**  
   Running the container without a fraglet (no `/FRAGLET` mount) produces the expected default output (e.g. "Hello World!").

2. **Guide examples**  
   Every example from the language's `fraglet/guide.md` (Examples section) runs correctly when executed as a fraglet and produces the expected output.

3. **Stdin**  
   Piping data into the container is forwarded to the fraglet (e.g. `echo "x" | fragletc --image <image> script.ext` produces output that reflects the piped input).  
   **N/A** for languages where the injection scope cannot read stdin (e.g. missing imports or runtime support).

4. **Arguments**  
   Positional arguments passed after the script reach the fraglet (e.g. `fragletc --image <image> script.ext arg1 arg2` and the program can read `argv`).  
   **N/A** where the language or injection model does not support arguments.

Reference implementations that satisfy this contract: **the-c-programming-language**, **python**, **ruby** (see 100hellos `*/fraglet/verify.sh` and fraglet entrypoint stdin tests).

## Where the contract is tested

| Layer | What is tested | How |
|-------|----------------|-----|
| **Entrypoint** | Stdin and stdin+args for 14 languages | `make test-entrypoint` — builds `fraglet-entrypoint`, mounts it into 100hellos images (`stdin_passthrough`, `stdin_languages`, `stdin_languages_extended`). |
| **veins_test** | Vein-by-name execution, args, optional stdin | `make test-veins` — builds `fragletc`, runs per-vein `act.sh` (e.g. basic run, `echo_args`). |
| **100hellos verify.sh** | Per-language image: default, guide examples, stdin, args | Run from fraglet repo: `make verify-100hellos LANGUAGE=<lang>` (installs fragletc, runs `$HELLOS_ROOT/<lang>/fraglet/verify.sh`). |

The same contract is enforced at different layers: entrypoint tests prove the binary and image work with stdin/args; veins_test proves the vein registry and CLI; verify.sh proves the full image and guide for each language.

## Running 100hellos verify.sh

verify.sh requires `fragletc` on PATH and Docker. From the **fraglet** repo:

```bash
make install-cli
export HELLOS_ROOT=/path/to/100hellos   # or leave unset to use default
make verify-100hellos LANGUAGE=ats
```

Or run a single verify script (fragletc must be on PATH, e.g. after `make install-cli`):

```bash
cd /path/to/100hellos && ats/fraglet/verify.sh
```

**Using locally built images:** Set `FRAGLET_VEINS_FORCE_TAG=local` so vein images use the `:local` tag instead of `:latest` (e.g. `100hellos/python:local`). Useful when testing with images built via `make <lang>` in 100hellos.

See [veins_test/README.md](veins_test/README.md) for veins_test structure and [entrypoint/releases/](entrypoint/releases/) for entrypoint changelog (e.g. v0.4.0 stdin passthrough).
