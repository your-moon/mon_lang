use std::ops::Range;
pub type Spanned<T> = (T, Span);
pub type Span = Range<usize>;

#[derive(Debug)]
pub enum Expr {
    Number(i32),
    Op(Box<Expr>, Opcode, Box<Expr>),
}

#[derive(Debug)]
pub enum Opcode {
    Add,
    Sub,
    Mul,
    Div,
}
