use crate::{
    front_end::ast::{Expr, Program, Stmt, UnaryOp},
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
            Expr::Number(num) => tacky::TackyVal::Constant(num.parse().unwrap()),
            Expr::Unary(op, expr) => {
                let src = self.emit_expr(*expr);
                let dst = TackyVal::Var(format!("tmp.{}", self.counter));
                self.counter += 1;
                let op = match op {
                    UnaryOp::Neg => tacky::UnaryOperator::Negate,
                    UnaryOp::Complement => tacky::UnaryOperator::Complement,
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
