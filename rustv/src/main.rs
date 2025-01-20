use std::{
    error::Error,
    fs::{self, File, OpenOptions},
};

mod gen;
use front_end::print_parse;

mod front_end;

const out_file_url: &str = "out.s";

fn main() -> anyhow::Result<()> {
    print_parse();
    let mut open_file = OpenOptions::new()
        .write(true)
        .create(true)
        .truncate(true)
        .open(out_file_url)?;
    let result = gen::gen(&mut open_file);
    return result;
}
