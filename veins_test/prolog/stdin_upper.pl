#!/usr/bin/env -S fragletc --vein=prolog
:- read_line_to_string(user_input, Line),
   upcase_atom(Line, Upper),
   write(Upper), nl.
