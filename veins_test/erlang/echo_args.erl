#!/usr/bin/env -S fragletc --vein=erlang
-export([main/0]).

main() ->
    Args = init:get_plain_arguments(),
    io:format("Args: ~s~n", [string:join(Args, " ")]).
