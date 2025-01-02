const std = @import("std");
const token = @import("../token.zig");

pub fn lex(expr: []const u8) ![]const token.Token {
    const stdout = std.io.getStdOut().writer();
    _ = try stdout.write(expr);
    std.debug.print(@tagName(token.Token.Let), .{});
    return &[_]token.Token{ token.Token.Let, token.Token.Import };
}
