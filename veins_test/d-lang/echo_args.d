#!/usr/bin/env -S fragletc --vein=d-lang
import std.stdio;
import std.array;

void main(string[] args) {
    writeln("Args: ", args[1..$].join(" "));
}
