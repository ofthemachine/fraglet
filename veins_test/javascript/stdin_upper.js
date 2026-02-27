#!/usr/bin/env -S fragletc --vein=javascript
const fs = require("fs");
const input = fs.readFileSync("/dev/stdin", "utf8");
console.log(input.trim().toUpperCase());
