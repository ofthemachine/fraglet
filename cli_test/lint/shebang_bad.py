#!/usr/bin/env -S fragletc --image python:3.12 --bogus x -v --mode main
#: d=A shebang fragletc would misread.
#: when=Use when testing the shebang lint.
#: network=none
print("hi")
