@echo off
REM Launcher for opening script files with fragletc (e.g. associate .py with this file).
REM Usage: fragletc-open.bat "path\to\script.py" [args...]
REM Place next to fragletc.exe or ensure fragletc is on PATH.

fragletc "%1" %*
pause
