#!/usr/bin/env -S fragletc --vein=ballerina
public function main(string... args) {
    foreach string arg in args {
        io:println(arg);
    }
}
