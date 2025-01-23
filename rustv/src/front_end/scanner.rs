#[derive(Debug)]
pub struct Scanner {
    source: Vec<char>,
    token_type: i32,
    cursor: usize,
    start: usize,
    line: usize,
}

impl Scanner {
    pub fn new(source: String) -> Scanner {
        return Scanner{
            token_type: 0,
            line: 0,
            cursor: 0,
            start:0,
            source: source.chars().collect(),
        }
    }

    fn next(&mut self) -> char {
        let c = self.source.get(self.cursor as usize).unwrap();
        self.cursor += 1;
        return c.to_owned();
    }

    fn is_alpha(&mut self, c: char) -> bool {
        match c {
            'a'..='z' | 'A'..='Z' => true,
            'а'..='я' | 'А'..='Я' => true,
            _ => false,
        }
    }

    pub fn scan(&mut self) -> Result<(), String> {
        self.start = self.cursor;
        let c = self.next();
        if self.is_alpha(c) {
            println!("alpha {:?}", c);
        } else {
            println!("check {:?}", c);
        }

        return Err("unimplemented".to_string());
    }
}
