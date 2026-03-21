# Opening .py (and other) files with fragletc on Windows

You can make Windows run fragletc when you double-click or "Open" a script file (e.g. `.py`). fragletc infers the vein from the file extension, so `fragletc script.py` already uses the python vein — no need to pass `--vein=python` unless you want to override.

## 1. Get fragletc

- Download `fragletc-windows-amd64.exe` from a [release](https://github.com/ofthemachine/fraglet/releases).
- Either put it on your PATH (e.g. `C:\bin`), or keep it in the same folder as `fragletc-open.bat` — the launcher finds fragletc in its own directory first, so PATH is optional.

## 2. Associate the extension with fragletc

**Option A — Open with (one-time per extension)**

1. Right-click a `.py` file → **Open with** → **Choose another app**.
2. **More apps** → **Look for another app on this PC**.
3. Browse to `fragletc.exe` (or your renamed `fragletc.exe`).
4. Check **Always use this app to open .py files** → OK.

Windows will run `fragletc.exe "C:\path\to\script.py"`. fragletc gets the path as the first argument and infers the vein from `.py`.

**Option B — Launcher that keeps the console open (recommended for double-click)**

If you double-click a script, the console often closes before you can see output. Use the `fragletc-open.bat` from the repo (or the same folder as your download):

1. Put `fragletc-open.bat` in the **same folder** as `fragletc-windows-amd64.exe` (or `fragletc.exe`). The launcher looks for fragletc in its own directory first, so you don’t need to add anything to PATH.
2. In **Open with** → **Choose another app**, select **`fragletc-open.bat`** and check **Always use this app to open .py files**.

Double-clicking a `.py` file will then run it with fragletc and leave the window open so you can read output.

## Other extensions

Repeat the same steps for other extensions (e.g. `.rb` for Ruby, `.js` for JavaScript) if you want "Open" to run them with fragletc. The vein is inferred from the extension; use `--vein=name` in the launcher if you need to override.

## Requirements

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (or another Docker-compatible engine) must be installed and running; fragletc invokes `docker run` to execute the script in the container.
