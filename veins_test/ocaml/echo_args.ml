#!/usr/bin/env -S fragletc --vein=ocaml
let () =
  let args = Array.to_list Sys.argv in
  let rest = List.tl args in
  Printf.printf "Args: %s\n" (String.concat " " rest)
