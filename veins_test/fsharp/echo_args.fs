#!/usr/bin/env -S fragletc --vein=fsharp
open System
let args = Environment.GetCommandLineArgs()[2..]  // [0]=runtime [1]=script
printfn "Args: %s" (String.Join(" ", args))
