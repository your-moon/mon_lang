fn main() {
    lalrpop::Configuration::new()
        .use_cargo_dir_conventions()
        .process_file("./src/front_end/grammar.lalrpop")
        .expect("failed to process grammar.lalrpop");
}
