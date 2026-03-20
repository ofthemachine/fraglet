#!/usr/bin/env -S fragletc --vein=python
# fraglet-meta: param=alpha:required
# fraglet-meta: param=beta:optional:default=x
# fraglet-meta: param=gamma:envvar=CUSTOM_GAMMA
print("multiline meta")
