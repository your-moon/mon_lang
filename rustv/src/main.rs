use std::fs;
use std::result::Result::Ok;

mod gen;
mod tacky;
mod tacky_gen;
use front_end::lang;
use front_end::lexer::Lexer;

mod front_end;

const IN_FILE_URL: &str = "examples/return.mn";

fn main() -> anyhow::Result<()> {
    let contents = fs::read_to_string(IN_FILE_URL).expect("Should have been able to read the file");
    let lexer = Lexer::new(&contents);
    let parsed = lang::ProgramParser::new().parse(lexer).unwrap();
    println!("{:?}", parsed);
    let mut tack_gen = tacky_gen::TackyGen::new();
    let tacky_program = tack_gen.emit_tacky(parsed);
    for instr in tack_gen.instructions {
        println!("{:?}", instr);
    }
    Ok(())
}
