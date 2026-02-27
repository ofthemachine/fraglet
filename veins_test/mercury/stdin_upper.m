#!/usr/bin/env -S fragletc --vein=mercury
main(!IO) :-
    io.read_line(Res, !IO),
    (
        Res = ok(Line),
        string.to_upper(Line, Upper),
        io.write_string(Upper, !IO),
        main(!IO)
    ;
        Res = eof
    ;
        Res = error(_)
    ).
