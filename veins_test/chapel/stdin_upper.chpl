#!/usr/bin/env -S fragletc --vein=chapel
use IO;
proc main() {
    var line: string;
    while stdin.readLine(line) {
        write(line.toUpper());
    }
}
