use anyhow::Result;

use super::ast::{Expr, Opcode};

pub enum IR {
    OP_ADD,
    OP_SUB,
    OP_MUL,
    OP_DIV,
}

pub fn simulate(ast: Expr) -> Result<()> {
    match  ast {
        Expr::Number(_) => todo!(),
        Expr::Op(expr, opcode, expr1) => todo!(),
    }
}
pub fn binary(lhs: Box<Expr>, op: Opcode, rhs: Expr) -> Result<()> {
    let rhs_val = simulate(rhs);
    match op {
        Opcode::Add => todo!(),
        Opcode::Sub => todo!(),
        Opcode::Mul => todo!(),
        Opcode::Div => todo!(),
    }
}
