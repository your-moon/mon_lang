use logos::{Logos, SpannedIter};
use thiserror::Error;

#[derive(Debug, PartialEq, Error)]
pub enum Error {
    #[error("Lexical error")]
    LexicalError,
}

pub type Spanned<Tok, Loc, Error> = Result<(Loc, Tok, Loc), Error>;

pub struct Lexer<'input> {
    token_stream: SpannedIter<'input, Token>,
}

impl<'input> Lexer<'input> {
    pub fn new(input: &'input str) -> Self {
        Self {
            token_stream: Token::lexer(input).spanned(),
        }
    }
}

impl<'input> Iterator for Lexer<'input> {
    type Item = Spanned<Token, usize, Error>;

    fn next(&mut self) -> Option<Self::Item> {
        self.token_stream.next().map(|(token, span)| match token {
            Ok(token) => Ok((span.start, token, span.end)),
            Err(_) => Err(Error::LexicalError),
        })
    }
}

#[derive(Logos, Clone, Debug, PartialEq)]
#[logos(skip r"[ \t\n\f]+", skip r"#.*\n?" )]
pub enum Token {
    #[token("фн")]
    Fn,
    #[token("->")]
    RightArrow,

    #[token("тоо")]
    NumberType,

    #[token("буц")]
    Return,

    #[token("үнэн")]
    True,

    #[token("худал")]
    False,

    #[token("+")]
    Add,

    #[token("-")]
    Minus,

    #[token("!")]
    Not,

    #[token("~")]
    Tilde,

    #[token("(")]
    LParen,

    #[token(")")]
    RParen,

    #[token("{")]
    LBrace,

    #[token("}")]
    RBrace,

    #[token(";")]
    Semicolon,

    #[token("хоосон")]
    Null,

    #[regex(r#""([^"\\]|\\["\\bnfrt]|u[a-fA-F0-9]{4})*""#, mnstring)]
    String(String),

    #[regex("[a-zA-Zа-яА-ЯёЁүҮеЕөӨ][a-zA-Zа-яА-ЯёЁүҮеЕөӨ]*", mnidentifier)]
    Identifier(String),

    #[regex(r"[0-9]+", mnnumber)]
    NumberLiteral(String),
}

fn mnnumber(lexer: &mut logos::Lexer<Token>) -> String {
    lexer.slice().to_string()
}

fn mnstring(lexer: &mut logos::Lexer<Token>) -> String {
    let slice = lexer.slice();
    slice[1..slice.len() - 1].to_string()
}

fn mnidentifier(lexer: &mut logos::Lexer<Token>) -> String {
    lexer.slice().to_string()
}
