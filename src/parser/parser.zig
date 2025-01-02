const std = @import("std");
const token = @import("../token.zig");

pub fn parse(source: []const token.Token) !void {
    const stdout = std.io.getStdOut().writer();
    for (source) |ltoken| {
        try stdout.print("{any}\n", .{ltoken});
    }
}
