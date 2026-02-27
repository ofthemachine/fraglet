#!/usr/bin/env -S fragletc --vein=rust
use std::env;

fn main() {
    let args: Vec<String> = env::args().skip(1).collect();
    println!("Args: {}", args.join(" "));
}
