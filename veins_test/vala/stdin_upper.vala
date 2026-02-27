#!/usr/bin/env -S fragletc --vein=vala
void main () {
    string? line = stdin.read_line ();
    if (line != null) {
        stdout.printf ("%s\n", line.up ());
    }
}
