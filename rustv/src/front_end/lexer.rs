use logos::Logos;

#[derive(Logos, Debug, PartialEq)]
#[logos(skip r"[ \t\n\f]+")] // Ignore this regex pattern between tokens
pub enum Token {
    #[token("үнэн")]
    True,

    #[token("худал")]
    False,

    #[token("+")]
    Add,

    #[token("-")]
    Subtract,

    #[token("*")]
    Multiply,

    #[token("/")]
    Divide,

    #[token("%")]
    Modulo,

    #[token("==")]
    Equal,

    #[token("!=")]
    NotEqual,

    #[token("<")]
    LessThan,
    #[token("<=")]
    LessThanOrEqual,
    #[token(">")]
    GreaterThan,
    #[token(">=")]
    GreaterThanOrEqual,
    #[token("&&")]
    And,

    #[token("||")]
    Or,

    #[token("!")]
    Not,
    #[token("=")]
    Assign,

    #[token("(")]
    LParen,

    #[token(")")]
    RParen,

    #[token("{")]
    LBrace,

    #[token("}")]
    RBrace,

    #[token("[")]
    LBracket,

    #[token("]")]
    RBracket,

    #[token(".")]
    Comma,
    #[token(",")]
    Colon,

    #[token(";")]
    Semicolon,

    #[token("хэрв")]
    If,

    #[token("эсвэл")]

    Else,
    #[token("хоосон")]
    Null,

    #[regex(r#""([^"\\]|\\["\\bnfrt]|u[a-fA-F0-9]{4})*""#, mnstring)]
    String(String),

     #[regex("[a-zA-Zа-яА-ЯёЁүҮеЕөӨ0-9_][a-zA-Zа-яА-ЯёЁүҮеЕөӨ0-9_]*", mnidentifier)]
    Identifier(String),
    Number(i32),

    Err(String),
}

fn mnstring(lexer: &mut logos::Lexer<Token>) -> String {
    let slice = lexer.slice();
    slice[1..slice.len() - 1].to_string()
}

fn mnidentifier(lexer: &mut logos::Lexer<Token>) -> String {
    lexer.slice().to_string()
}
