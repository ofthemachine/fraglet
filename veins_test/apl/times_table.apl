#!/usr/bin/env -S fragletc --vein=apl
⍝ A small multiplication table (many-line fraglet)
⍝ Build rows and cols, then outer product.
n ← 5
rows ← ⍳n
cols ← ⍳n
table ← rows ∘.× cols
'5×5 times table:'
table
'Done.'
