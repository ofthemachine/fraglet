#!/usr/bin/env -S fragletc --vein=hare
use fmt;
use os;
use bufio;
use io;

export fn main() void = {
    let scan = bufio::newscanner(os::stdin, 4096);
    defer bufio::finish(&scan);
    match (bufio::scan_line(&scan)) {
    case let s: const str =>
        fmt::println(s)!;
    case io::EOF => void;
    case let e: io::error => void;
    };
};
