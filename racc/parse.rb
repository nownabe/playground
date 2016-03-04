require "pp"
require "./erlang_parser"

pp ErlangParser.new.parse(File.read(ARGV[0]))
