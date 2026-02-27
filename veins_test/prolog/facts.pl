#!/usr/bin/env -S fragletc --vein=prolog
assertz(likes(alice, chocolate)).
assertz(likes(bob, ice_cream)).
likes(alice, What), write("Alice likes: "), write(What), nl.
halt.
