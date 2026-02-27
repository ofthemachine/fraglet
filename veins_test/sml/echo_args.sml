#!/usr/bin/env -S fragletc --vein=sml
val allArgs = CommandLine.arguments ();
val args = List.drop (allArgs, 3);
print ("Args: " ^ String.concatWith " " args ^ "\n");
