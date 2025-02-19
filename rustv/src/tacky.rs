#[derive(Debug, Clone)]
pub enum UnaryOperator {
    Complement,
    Negate,
}

#[derive(Debug, Clone)]
pub enum TackyVal {
    Constant(i32),
    Var(String),
}

#[derive(Debug)]
pub enum Instruction {
    Return(TackyVal),
    Unary {
        op: UnaryOperator,
        src: TackyVal,
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
