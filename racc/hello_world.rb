require "pp"
require "ripper"

pp Ripper.lex('puts "Hello, world!"')
