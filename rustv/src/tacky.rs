#[derive(Debug, Clone)]
pub enum UnaryOperator {
    Complement,
    Negate,
    Not,
}

#[derive(Debug, Clone)]
pub enum InfixOperator {
    Add,
    Sub,
    Mul,
    Div,
}

#[derive(Debug, Clone)]
pub enum TackyVal {
    Constant(i32),
    Var(String),
}

#[derive(Debug, Clone)]
pub enum Instruction {
    Return(TackyVal),
    Unary {
        op: UnaryOperator,
        src: TackyVal,
        dst: TackyVal,
        },
    Binary {
        op: InfixOperator,
        left: TackyVal,
        right: TackyVal,
        dst: TackyVal,
    },
}

#[derive(Debug)]
pub struct FunctionDefinition {
    pub name: String,
    pub body: Vec<Instruction>,
}

#[derive(Debug)]
pub struct TackyProgram {
    pub function: FunctionDefinition,
}
