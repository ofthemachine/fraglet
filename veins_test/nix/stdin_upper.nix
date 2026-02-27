#!/usr/bin/env -S fragletc --vein=nix
builtins.getEnv "STDIN_LINE" or ""
