#!/usr/bin/env -S fragletc --vein=chapel
use IO;

proc main() {
    var line: string;
    if readln(line) {
        writeln(line);
    }
}
