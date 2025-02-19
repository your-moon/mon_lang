pub mod ast;
pub mod lexer;
use crate::front_end::lexer::Token;
use lalrpop_util::{lalrpop_mod, ParseError};
lalrpop_mod!(pub lang, "/src/front_end/grammar.rs");
