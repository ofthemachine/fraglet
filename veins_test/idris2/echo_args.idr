#!/usr/bin/env -S fragletc --vein=idris2
import System
import Data.String

main : IO ()
main = do
    args <- getArgs
    putStrLn ("Args: " ++ joinBy " " (drop 1 args))
