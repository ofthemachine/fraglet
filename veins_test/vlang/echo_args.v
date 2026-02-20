#!/usr/bin/env -S fragletc --vein=vlang
import os

fn main() {
	println('Args: ' + os.args[1..].join(' '))
}
