#!/usr/bin/env -S fragletc --vein=odin
import "core:fmt"
import "core:os"
import "core:strings"

main :: proc() {
    buf: [256]u8
    n, err := os.read(os.stdin, buf[:])
    if err == nil && n > 0 {
        s := string(buf[:n])
        fmt.print(strings.to_upper(s))
    }
}
