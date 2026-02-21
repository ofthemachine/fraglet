#!/usr/bin/env -S fragletc --vein=rust
use std::env;

fn main() {
    let args: Vec<String> = env::args().skip(1).collect();
    let mut s = String::new();
    for (i, a) in args.iter().enumerate() {
        if i > 0 {
            s.push(' ');
        }
        s.push_str(a);
    }
    println!("Args: {}", s);
}
