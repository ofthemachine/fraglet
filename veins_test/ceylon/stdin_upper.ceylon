#!/usr/bin/env -S fragletc --vein=ceylon
shared void run() {
    while (exists line = process.readLine()) {
        print(line.uppercased);
    }
}
