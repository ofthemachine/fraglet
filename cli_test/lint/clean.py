#!/usr/bin/env -S fragletc --vein=python
#: d=Reference conformant header: every param described, default= read straight from the env.
#: when=Use when a page must be captured as an image rather than read as text.
#: network=none
#: param=url:required:d=Page URL to capture
#: param=headers:d=JSON object of extra HTTP headers
#: param=settle_ms:default=2000:description=Extra wait after page load
#: output=page.png
import os
print(os.environ["URL"], os.environ.get("HEADERS", ""), int(os.environ["SETTLE_MS"]))
