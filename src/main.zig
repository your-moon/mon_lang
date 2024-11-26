const std = @import("std");
const parser = @import("parser/parser.zig");
const lexer = @import("lexer/lexer.zig");

pub fn main() !void {
    _ = try parser.parse("hello this is source");
    _ = try lexer.lex("hi");
}
