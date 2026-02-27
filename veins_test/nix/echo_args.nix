#!/usr/bin/env -S fragletc --vein=nix
builtins.concatStringsSep " " (builtins.tail builtins.getEnv "ARGS")
