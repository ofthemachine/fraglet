#!/usr/bin/env -S fragletc --vein=hare
use fmt;
use os;
use strings;

export fn main() void = {
    let args = os::args[1..];
    fmt::print("Args: ")!;
    for (let i = 0z; i < len(args); i += 1) {
        if (i > 0) fmt::print(" ")!;
        fmt::print(args[i])!;
    };
    fmt::println("")!;
};
