const std = @import("std");
const parser = @import("parser/parser.zig");
const lexer = @import("lexer/lexer.zig");
const token = @import("token.zig");

pub fn main() !void {
    const source = "hello this is source";
    const lexedTokens = try lexer.lex(source);
    _ = try parser.parse(lexedTokens);
}
