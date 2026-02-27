#!/usr/bin/env -S fragletc --vein=zig
pub fn main() !void {
    const stdout = std.io.getStdOut().writer();
    try stdout.print("Hello from fragment!\n", .{});
}
