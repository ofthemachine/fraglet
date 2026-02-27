#!/usr/bin/env -S fragletc --vein=ats
implement main0() =
  let
    val lines = streamize_fileref_line(stdin_ref)
    val () = lines.foreach()(lam x => println!(x))
  in () end
