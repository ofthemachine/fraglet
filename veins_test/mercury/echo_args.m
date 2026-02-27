#!/usr/bin/env -S fragletc --vein=mercury
main(!IO) :-
    io.command_line_arguments(Args, !IO),
    string.join_list(" ", Args, Joined),
    io.format("Args: %s\n", [s(Joined)], !IO).
