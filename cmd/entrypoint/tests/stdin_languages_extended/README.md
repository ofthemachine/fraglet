# stdin_languages_extended

Docker-based stdin checks for eight 100hellos images. Requires Docker and pulled `100hellos/*:latest` images.

**Full-suite backups (restore after local bisection):**

- `act.full.sh` — copy over `act.sh` to reset (`cp act.full.sh act.sh`)
- `assert.full.txt` — copy over `assert.txt` to reset

To find a hang: trim matching chunks out of `act.sh` and `assert.txt`, run `make test-entrypoint` or:

```bash
cd entrypoint
ENTRYPOINT_TEST_SUITE_DIR="$(pwd)/tests/stdin_languages_extended" go test -tags=integration -v ./tests -run 'TestCLI'
```

When done, restore from `act.full.sh` / `assert.full.txt`.

**Wall-clock when bisecting:** pick a cap from the last *passing* run—e.g. **2–3×** that duration or **~30–45s** if you’re warm on images. If nothing prints for that long after the previous case, you’ve found the stuck `docker run`. Reserve multi-minute caps for **cold pull / first image** only, not for “did this step hang?” on a machine that already passed the same suite recently.
