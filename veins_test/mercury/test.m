#!/usr/bin/env -S fragletc --vein=mercury
main(!IO) :-
    io.write_string("Hello from fragment!\n", !IO).
