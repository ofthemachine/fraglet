#!/usr/bin/env -S fragletc --vein=ballerina
public function main() {
    string? line = io:readln("");
    if line is string {
        io:println(line);
    }
}
