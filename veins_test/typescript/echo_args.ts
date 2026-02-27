#!/usr/bin/env -S fragletc --vein=typescript
declare var process: any;
console.log("Args: " + process.argv.slice(2).join(" "));
