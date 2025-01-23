pub mod ast;
pub mod lexer;
pub mod scanner;
pub mod ir;
use crate::front_end::lexer::Token;
use lalrpop_util::{lalrpop_mod, ParseError};
lalrpop_mod!(pub calculator1, "/src/front_end/grammar.rs");

pub fn print_parse() {
    let parsed = calculator1::ExprParser::new()
        .parse("22 + 44 * 66")
        .unwrap();
    println!("{:?}", parsed);
}

#[test]
fn calculator1() {
    assert!(calculator1::ExprParser::new().parse("22").is_ok());
    assert!(calculator1::ExprParser::new().parse("22 + 44").is_ok());
    assert!(calculator1::ExprParser::new().parse("22 + 44 * 66").is_ok());
}
