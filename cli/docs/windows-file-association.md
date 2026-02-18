# Opening .py (and other) files with fragletc on Windows

You can make Windows run fragletc when you double-click or "Open" a script file (e.g. `.py`). fragletc infers the vein from the file extension, so `fragletc script.py` already uses the python vein — no need to pass `--vein=python` unless you want to override.

## 1. Put fragletc on your PATH

- Download `fragletc-windows-amd64.exe` from a [release](https://github.com/ofthemachine/fraglet/releases).
- Put it in a folder that’s on your PATH (e.g. `C:\bin` or your user profile), and optionally rename to `fragletc.exe`.

## 2. Associate the extension with fragletc

**Option A — Open with (one-time per extension)**

1. Right-click a `.py` file → **Open with** → **Choose another app**.
2. **More apps** → **Look for another app on this PC**.
3. Browse to `fragletc.exe` (or your renamed `fragletc.exe`).
4. Check **Always use this app to open .py files** → OK.

Windows will run `fragletc.exe "C:\path\to\script.py"`. fragletc gets the path as the first argument and infers the vein from `.py`.

**Option B — Launcher that keeps the console open (recommended for double-click)**

If you double-click a script, the console often closes before you can see output. Use a small launcher that runs fragletc and then pauses:

1. Save this as e.g. `fragletc-open.bat` next to `fragletc.exe` (or anywhere on PATH):

   ```batch
   @echo off
   fragletc "%1" %*
   pause
   ```

2. In **Open with** → **Choose another app**, select this **`.bat` file** (not fragletc.exe) and check **Always use this app to open .py files**.

Double-clicking a `.py` file will then run it with fragletc and leave the window open so you can read output.

## Other extensions

Repeat the same steps for other extensions (e.g. `.rb` for Ruby, `.js` for JavaScript) if you want "Open" to run them with fragletc. The vein is inferred from the extension; use `--vein=name` in the launcher if you need to override.

## Requirements

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (or another Docker-compatible engine) must be installed and running; fragletc invokes `docker run` to execute the script in the container.
