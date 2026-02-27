#!/usr/bin/env -S fragletc --vein=d-lang
import std.stdio;
import std.string;
import std.conv;

void main() {
    foreach (line; stdin.byLine) {
        writeln(line.to!string.strip.toUpper);
    }
}
