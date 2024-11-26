const std = @import("std");

pub fn parse(source: []const u8) !void {
    const stdout = std.io.getStdOut().writer();
    try stdout.print("{s}\n", .{source});
}
