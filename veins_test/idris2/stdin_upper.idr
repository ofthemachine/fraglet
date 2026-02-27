#!/usr/bin/env -S fragletc --vein=idris2
import Data.String

main : IO ()
main = do
  l <- getLine
  putStrLn (toUpper l)
