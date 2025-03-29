use std::fs;

mod tacky;
mod tacky_gen;
use anyhow::{Ok, Result};
use clap::Parser;
use front_end::lang;
use front_end::lexer::Lexer;

mod front_end;
use clap::arg;

const IN_FILE_URL: &str = "examples/return.mn";

#[derive(Parser)]
struct Commands {
    #[arg(short, long)]
    lex: bool,
    #[arg(short, long)]
    parse: bool,
    #[arg(short, long)]
    tacky: bool,
    #[arg(short, long)]
    gen: bool,
    #[arg(short, long)]
    compile: bool,
}

impl Commands {
    pub fn run(&self) -> Result<()> {
        if self.lex {
            let contents =
                fs::read_to_string(IN_FILE_URL).expect("Should have been able to read the file");
            let lexer = Lexer::new(&contents);
            for token in lexer {
                println!("{:?}", token);
            }
            return Ok(());
        }
        if self.parse {
            let contents =
                fs::read_to_string(IN_FILE_URL).expect("Should have been able to read the file");
            let lexer = Lexer::new(&contents);
            for token in lexer {
                println!("{:?}", token);
            }

            let lexer = Lexer::new(&contents);
            let parsed = lang::ProgramParser::new().parse(lexer).unwrap();
            println!("{:?}", parsed);
            return Ok(());
        }
        if self.tacky {
            let contents =
                fs::read_to_string(IN_FILE_URL).expect("Should have been able to read the file");
            let lexer = Lexer::new(&contents);
            let parsed = lang::ProgramParser::new().parse(lexer).unwrap();
            let mut tack_gen = tacky_gen::TackyGen::new();
            let tacky_program = tack_gen.emit_tacky(parsed);
            for instr in tacky_program.function.body {
                println!("{:?}", instr);
            }
            return Ok(());
        }
        if self.gen {
            let contents =
                fs::read_to_string(IN_FILE_URL).expect("Should have been able to read the file");
            let lexer = Lexer::new(&contents);
            let parsed = lang::ProgramParser::new().parse(lexer).unwrap();
            let mut tack_gen = tacky_gen::TackyGen::new();
            let tacky_program = tack_gen.emit_tacky(parsed);
            for instr in tack_gen.instructions {
                println!("{:?}", instr);
            }
            return Ok(());
        }

        if self.compile {
            unimplemented!("Compile not implemented");
        }
        Ok(())
    }
}

struct Cli {}

impl Cli {
    pub fn run() -> Result<()> {
        let commands = Commands::parse();
        commands.run()
    }
}

fn main() -> anyhow::Result<()> {
    Cli::run()
}
