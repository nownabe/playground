#[derive(Debug)]
struct ImportantExcerpt<'a> {
    part: &'a str,
}

fn main() {
    let str1 = String::from("abcd");
    let str2 = "xyz";

    let result = longest(str1.as_str(), str2);

    println!("The longest string is {}", result);

    let i: ImportantExcerpt;
    let novel = String::from("Call me Ishmael. Some years ago...");

    {
        let first_sentence = novel.split('.').next().expect("Could not find a '.'");
        i = ImportantExcerpt {
            part: first_sentence,
        };
    }

    println!("important = {:?}", i);
}

fn longest<'a>(x: &'a str, y: &'a str) -> &'a str {
    if x.len() > y.len() {
        x
    } else {
        y
    }
}
