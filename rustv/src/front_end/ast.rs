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
pub struct ExprUnary {
    pub op: UnaryOp,
    pub expr: Spanned<Expr>,
}

#[derive(Debug)]
pub struct ExprInfix {
    pub op: OpInfix,
    pub lhs: Spanned<Expr>,
    pub rhs: Spanned<Expr>,
}


#[derive(Debug)]
pub enum Expr {
    Binary(Box<ExprInfix>),
    Unary(Box<ExprUnary>),
    Literal(ExprLiteral),
    Number(String),
}

#[derive(Debug)]
pub enum ExprLiteral {
    String(String),
    Number(String),
}

#[derive(Debug)]
pub enum UnaryOp {
    Neg,
    Complement,
    Not,
}

#[derive(Debug)]
pub enum OpInfix {
    Add,
    Sub,
    Mul,
    Div,

    Eq,
    Ne,

    Lt,
    Le,

    Gt,
    Ge,

    And,
    Or,
}

#[derive(Debug)]
pub struct ReturnStmt {
    pub expr: Option<Expr>,
}