@echo off
REM Launcher for opening script files with fragletc (e.g. associate .py with this file).
REM Usage: fragletc-open.bat "path\to\script.py" [args...]
REM Looks for fragletc in the same folder as this .bat first, then PATH.

set "FRAGLETC=%~dp0fragletc.exe"
if not exist "%FRAGLETC%" set "FRAGLETC=%~dp0fragletc-windows-amd64.exe"
if not exist "%FRAGLETC%" set "FRAGLETC=fragletc"

"%FRAGLETC%" "%1" %*
pause
