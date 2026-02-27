#!/usr/bin/env -S fragletc --vein=d-lang
import std.stdio;
import std.array : join;

void main(string[] args) {
    writeln("Args: ", join(args[1..$], " "));
}
