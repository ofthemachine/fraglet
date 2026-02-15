#!/usr/bin/env -S fragletc --vein=apl
⍝ FizzBuzz 1–15: index 1–4 from divisibility (3,5,15), pick (⍕n) or Fizz/Buzz/FizzBuzz
{((⍕⍵) 'Fizz' 'Buzz' 'FizzBuzz')[1+((0=15|⍵)×3)+((0=3|⍵)×(0≠15|⍵))+(2×(0=5|⍵)×(0≠15|⍵))]}¨⍳15
