#!/usr/bin/env -S fragletc --vein=odin
import "core:fmt"
import "core:os"
import "core:strings"

main :: proc() {
    args := os.args[1:]
    fmt.println("Args:", strings.join(args, " "))
}
