#!/usr/bin/env -S fragletc --vein=haskell
import System.Environment (getArgs)

main = do
  args <- getArgs
  putStrLn $ "Args: " ++ unwords args
