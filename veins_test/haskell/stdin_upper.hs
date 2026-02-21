#!/usr/bin/env -S fragletc --vein=haskell
import Data.Char (toUpper)

main = interact (map toUpper)
