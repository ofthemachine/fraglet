#!/usr/bin/env -S fragletc --vein=rust
use std::io::{self, BufRead};

fn main() {
    let stdin = io::stdin();
    for line in stdin.lock().lines() {
        if let Ok(l) = line {
            println!("{}", l.to_uppercase());
        }
    }
}
