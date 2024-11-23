const std = @import("std");
const parser = @import("parser/parser.zig");
const lexer = @import("lexer/lexer.zig");

pub fn main() !void {
    std.debug.print("Main", .{});
    parser.parse();
    _ = try lexer.lex("hi");
}
