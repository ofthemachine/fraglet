#!/usr/bin/env -S fragletc --vein=d-lang
import std.stdio;
import std.string : toUpper, strip;
import std.conv : to;

void main() {
    foreach (line; stdin.byLine) {
        writeln(line.to!string.strip.toUpper);
    }
}
