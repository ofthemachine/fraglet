#!/usr/bin/env -S fragletc --vein=zig
pub fn main() !void {
    var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
    defer arena.deinit();
    var args = try std.process.argsWithAllocator(arena.allocator());
    defer args.deinit();
    _ = args.next();
    var list = std.ArrayList(u8).init(arena.allocator());
    while (args.next()) |arg| {
        if (list.items.len > 0) try list.appendSlice(" ");
        try list.appendSlice(arg);
    }
    std.debug.print("Args: {s}\n", .{list.items});
}
