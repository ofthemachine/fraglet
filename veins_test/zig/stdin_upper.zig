#!/usr/bin/env -S fragletc --vein=zig
const std = @import("std");

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    const allocator = gpa.allocator();
    const input = try std.io.getStdIn().readToEndAlloc(allocator, 1024 * 1024);
    defer allocator.free(input);
    for (input) |c| {
        const out = [_]u8{std.ascii.toUpper(c)};
        _ = try std.io.getStdOut().writer().write(&out);
    }
}
