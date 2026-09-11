#!/usr/bin/env -S fragletc --vein=python
#: d=One header, many mistakes.
#: network=bridge
#: param=engine:d=TeX engine:default=pdflatex
#: param=n:required:default=1:d=Count
#: param=count:optional:default=5:d=How many
#: param=Input-File:file:d=Bad alias
#: param=url:requried:d=Typo modifier
#: param=url:d=Declared twice
#: param=ghost:d=Never read
#: param=bare
#: output=/abs.png
#: output=a.png
#: output=a.png
import os
count = int(os.environ.get("COUNT", "5"))
print(os.environ["ENGINE"], os.environ["N"], count, os.environ["BARE"], os.environ["URL"], os.environ["INPUT-FILE"])
