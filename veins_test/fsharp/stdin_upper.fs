#!/usr/bin/env -S fragletc --vein=fsharp
open System
let rec readLines () =
    match Console.ReadLine() with
    | null -> ()
    | line -> printfn "%s" (line.ToUpper()); readLines ()
readLines ()
