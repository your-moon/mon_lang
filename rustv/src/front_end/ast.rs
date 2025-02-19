use std::ops::Range;
pub type Spanned<T> = (T, Span);
pub type Span = Range<usize>;

#[derive(Debug)]
pub struct Program {
    pub func: FnDef,
}

#[derive(Debug)]
pub struct FnDef {
    pub name: String,
    pub body: Stmt,
}

#[derive(Debug)]
pub enum Stmt {
    Return(ReturnStmt),
}

#[derive(Debug)]
pub enum Expr {
    Number(String),
    Unary(UnaryOp, Box<Expr>),
}

#[derive(Debug)]
pub enum UnaryOp {
    Neg,
    Complement,
}

#[derive(Debug)]
pub struct ReturnStmt {
    pub expr: Option<Expr>,
}
