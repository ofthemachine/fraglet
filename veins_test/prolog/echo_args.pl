#!/usr/bin/env -S fragletc --vein=prolog
:- current_prolog_flag(argv, Args),
   atomic_list_concat(Args, ' ', Joined),
   format("Args: ~w~n", [Joined]).
