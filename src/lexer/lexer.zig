const std = @import("std");

const Token = enum {
    Let,
    Meow,
    Variable,
    Fn,
};

pub fn lex(expr: []const u8) !Token {
    const stdout = std.io.getStdOut().writer();
    _ = try stdout.write(expr);
    std.debug.print(@tagName(Token.Let), .{});
    return Token.Let;
}
