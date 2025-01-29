use std::result::Result::Ok;
use std::
    fs::{OpenOptions}
;
use std::
    fs;

mod gen;
use front_end::lexer::Token;
use front_end::print_parse;
use logos::Logos;

mod front_end;

const IN_FILE_URL: &str = "examples/alpha.mn";
const OUT_FILE_URL: &str = "out/out.asm";
const OBJ_FILE: &str = "out/out.o";
const EXEC_NAME: &str = "program";


// fn exec() -> anyhow::Result<()> {
//     let mut open_file = OpenOptions::new()
//         .write(true)
//         // .create(true)
//         .truncate(true)
//         .open(OUT_FILE_URL)?;
//     let result = gen::gen(&mut open_file);
//     // let contents = fs::read_to_string(IN_FILE_URL)
//     //     .expect("Should have been able to read the file");
//     // let mut scanner_ = scanner::Scanner::new(contents.clone());
//     // for _n in 1..(contents.to_string().len() as i32) {
//     //     let token = scanner_.scan();
//     //     println!("{:?}", token);
//     // }
//
//     let _output = Command::new("nasm").arg("-felf64").arg(OUT_FILE_URL).output()?;
//
//     let _output = Command::new("ld").arg("-o").arg(EXEC_NAME).arg(OBJ_FILE).output()?;
//
//     let command: String = "./".to_string() + EXEC_NAME;
//     let output = Command::new(command).output()?;
//     println!("{:?}",output.status.code().unwrap());
//
//     let _ = fs::remove_file(OBJ_FILE);
//     result
// }


fn main() -> anyhow::Result<()> {
    let contents = fs::read_to_string(IN_FILE_URL)
        .expect("Should have been able to read the file");
    for result in Token::lexer(&contents) {
        match result {
            Ok(token) => println!("{:#?}", token),
            Err(e) => panic!("some error occurred {:?}", e),
        }
    }
    // exec()?;

    print_parse(contents.clone());
    Ok(())
}
