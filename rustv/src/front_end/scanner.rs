use anyhow::Ok;
use anyhow::Result;

use super::lexer::Token;

#[derive(Debug)]
pub struct Scanner {
    source: Vec<char>,
    token_type: i32,
    cursor: usize,
    start: usize,
    line: usize,
}

impl Scanner {
    pub fn new(source: String) -> Scanner {
        Scanner{
            token_type: 0,
            line: 0,
            cursor: 0,
            start:0,
            source: source.chars().collect(),
        }
    }

    fn next(&mut self) -> char {
        let c = self.source.get(self.cursor).unwrap();
        self.cursor += 1;
        let var_name = c.to_owned();
        var_name
    }

    fn is_alpha(&mut self, c: char) -> bool {
        matches!(c, 'a'..='z' | 'A'..='Z' | 'а'..='я' | 'А'..='Я')
    }

    pub fn scan(&mut self) -> Result<Token> {
        self.start = self.cursor;
        let c = self.next();
        match c {
            c if self.is_alpha(c) => Result::Ok(Token::Identifier(c.to_string())),
            '(' => Ok(Token::LParen),
            ')' => Ok(Token::RParen),
            _ => unimplemented!()
        }
    }
}
