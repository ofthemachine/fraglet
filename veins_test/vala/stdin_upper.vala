#!/usr/bin/env -S fragletc --vein=vala
void main () {
    string? line;
    while ((line = stdin.read_line()) != null) {
        print(line.up() + "\n");
    }
}
