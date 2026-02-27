#!/usr/bin/env -S fragletc --vein=sml
fun loop () =
    case TextIO.inputLine TextIO.stdIn of
        NONE => ()
      | SOME s => (print (String.map Char.toUpper s); loop ());
val () = loop ();
