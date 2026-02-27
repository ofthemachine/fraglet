#!/usr/bin/env -S fragletc --vein=typescript
const fs = require("fs");
const input: string = fs.readFileSync("/dev/stdin", "utf8");
console.log(input.trim().toUpperCase());
