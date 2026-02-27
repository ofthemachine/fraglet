#!/usr/bin/env -S fragletc --vein=erlang
-export([main/0]).

main() ->
    case io:get_line("") of
        eof -> ok;
        Line -> io:format("~s", [string:to_upper(Line)]), main()
    end.
