use std::
    fs::{self, OpenOptions}
;
use std::process::Command;

mod gen;
use anyhow::Ok;
use front_end::print_parse;

mod front_end;

const OUT_FILE_URL: &str = "out/out.asm";
const OBJ_FILE: &str = "out/out.o";
const EXEC_NAME: &str = "program";


fn exec() -> anyhow::Result<()> {
    let mut open_file = OpenOptions::new()
        .write(true)
        .create(true)
        .truncate(true)
        .open(OUT_FILE_URL)?;
    let result = gen::gen(&mut open_file);

    let _output = Command::new("nasm").arg("-felf64").arg(OUT_FILE_URL).output()?;

    let _output = Command::new("ld").arg("-o").arg(EXEC_NAME).arg(OBJ_FILE).output()?;

    let command: String = "./".to_string() + EXEC_NAME;
    let output = Command::new(command).output()?;
    println!("{:?}",output.status.code().unwrap());

    let _ = fs::remove_file(OBJ_FILE);
    result
}

fn main() -> anyhow::Result<()> {
    print_parse();
    exec()?;

    Ok(())
}
