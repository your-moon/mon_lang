use anyhow::Result;
use std::{fs::File, io::Write};

pub fn gen(out_file: &mut File) -> Result<()> {
    writeln!(out_file, "global _start")?;
    writeln!(out_file, "section .text")?;
    writeln!(out_file, "_start:")?;
    writeln!(out_file, "    ; generated code")?;
    writeln!(out_file, "    mov rax, 60")?;
    writeln!(out_file, "    mov rdi, 69")?;
    writeln!(out_file, "    syscall")?;
    Ok(())
}

pub fn gen_add(out_file: &mut File, lhs: i32, rhs: i32) -> Result<()> {
    writeln!(out_file, "    ; add instruction")?;
    writeln!(out_file, "    mov rax, {}", lhs)?;
    writeln!(out_file, "    mov rbx, {}", rhs)?;
    writeln!(out_file, "    add rax, rbx")?;
    Ok(())
}

pub fn gen_sub(out_file: &mut File, lhs: i32, rhs: i32) -> Result<()> {
    writeln!(out_file, "    ; sub instruction")?;
    writeln!(out_file, "    mov rax, {}", lhs)?;
    writeln!(out_file, "    mov rbx, {}", rhs)?;
    writeln!(out_file, "    sub rax, rbx")?;
    Ok(())
}
