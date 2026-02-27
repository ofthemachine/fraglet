#!/usr/bin/env -S fragletc --vein=odin
import "core:fmt"
import "core:os"
import "core:strings"

main :: proc() {
    buf: [4096]u8
    for {
        n, err := os.read(os.stdin, buf[:])
        if err != nil || n == 0 { break }
        line := string(buf[:n])
        fmt.print(strings.to_upper(line))
    }
}
