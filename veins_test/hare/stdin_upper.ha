#!/usr/bin/env -S fragletc --vein=hare
use fmt;
use os;
use io;
use bufio;
use ascii;

export fn main() void = {
    const scan = bufio::newscanner(os::stdin, 4096);
    defer bufio::finish(&scan);
    for (true) {
        match (bufio::scan_line(&scan)) {
        case let line: const str =>
            const upper = ascii::strupper(line);
            fmt::println(upper as str)!;
            free(upper as str);
        case io::EOF =>
            break;
        case let err: io::error =>
            break;
        };
    };
};
