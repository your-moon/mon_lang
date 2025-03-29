use crate::{
    front_end::ast::{self, Expr, ExprLiteral, Program, Stmt, UnaryOp},
    tacky::{self, Instruction, TackyVal},
};

pub struct TackyGen {
    pub counter: i32,
    pub instructions: Vec<Instruction>,
}

impl TackyGen {
    pub fn new() -> Self {
        TackyGen {
            counter: 0,
            instructions: Vec::new(),
        }
    }

    pub fn emit_tacky(&mut self, program: Program) -> tacky::TackyProgram {
        match program.func.body {
            Stmt::Return(ret_stmt) => {
                if let Some(expr) = ret_stmt.expr {
                    let val = self.emit_expr(expr);
                    self.instructions.push(tacky::Instruction::Return(val));
                }
            }
        }

        tacky::TackyProgram {
            function: tacky::FunctionDefinition {
                name: program.func.name,
                body: self.instructions.clone(),
            },
        }
    }

    fn emit_expr(&mut self, expr: crate::front_end::ast::Expr) -> tacky::TackyVal {
        match expr {
            Expr::Literal(expr_literal) => {
                match expr_literal {
                    ExprLiteral::String(s) => tacky::TackyVal::Constant(s.parse().unwrap()),
                    ExprLiteral::Number(n) => tacky::TackyVal::Constant(n.parse().unwrap()),
                }
            }
            Expr::Binary(expr_infix) => {
                let left = self.emit_expr(expr_infix.lhs.0);
                let right = self.emit_expr(expr_infix.rhs.0);
                let dst = TackyVal::Var(format!("tmp.{}", self.counter));
                self.counter += 1;
                let op = match expr_infix.op {
                    ast::OpInfix::Add => tacky::InfixOperator::Add,
                    _ => todo!(),
                };
                self.instructions.push(tacky::Instruction::Binary {
                    op: op.clone(),
                    left: left.clone(),
                    right: right.clone(),
                    dst: dst.clone(),
                });
                dst
            }
            Expr::Number(num) => tacky::TackyVal::Constant(num.parse().unwrap()),
            Expr::Unary(expr_prefix) => {
                let src = self.emit_expr(expr_prefix.expr.0);
                let dst = TackyVal::Var(format!("tmp.{}", self.counter));
                self.counter += 1;
                let op = match expr_prefix.op {
                    UnaryOp::Neg => tacky::UnaryOperator::Negate,
                    UnaryOp::Complement => tacky::UnaryOperator::Complement,
                    UnaryOp::Not => tacky::UnaryOperator::Not,
                };
                self.instructions.push(tacky::Instruction::Unary {
                    op: op.clone(),
                    src: src.clone(),
                    dst: dst.clone(),
                });
                Instruction::Unary {
                    op,
                    src,
                    dst: dst.clone(),
                };
                dst
            }
        }
    }
}
