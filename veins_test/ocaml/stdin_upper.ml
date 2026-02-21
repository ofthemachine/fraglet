#!/usr/bin/env -S fragletc --vein=ocaml
let () =
  try
    while true do
      let line = input_line stdin in
      print_endline (String.uppercase_ascii line)
    done
  with End_of_file -> ()
