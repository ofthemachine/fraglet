#!/usr/bin/env -S fragletc --vein=vala
void main(string[] args) {
    string[] a = args[1:args.length];
    print("Args: " + string.joinv(" ", a) + "\n");
}
