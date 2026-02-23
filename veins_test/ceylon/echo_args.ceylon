#!/usr/bin/env -S fragletc --vein=ceylon
shared void run() {
    print("Args: ``" ".join(process.arguments)``");
}
