#!/usr/bin/env -S fragletc --vein=zig
pub fn main() !void {
    std.debug.print("Hello from fragment!\n", .{});
}
