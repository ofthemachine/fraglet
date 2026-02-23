#!/usr/bin/env -S fragletc --vein=coffeescript
fs = require 'fs'
input = fs.readFileSync 0, 'utf8'
input.split('\n').forEach (line) ->
  console.log line.toUpperCase()
