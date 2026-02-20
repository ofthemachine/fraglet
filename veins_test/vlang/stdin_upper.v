#!/usr/bin/env -S fragletc --vein=vlang
import os

fn main() {
	line := os.input('')
	println(line.to_upper())
}
