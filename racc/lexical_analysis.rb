require "strscan"
require "pp"

status = File.read(ARGV[0])

ss = StringScanner.new(status)
tokens = []

until ss.eos?
  ss.scan(/"(.+)"/)       ? tokens << [:STRING,      ss[1]] :
  ss.scan(/\d+\.\d+/)     ? tokens << [:FLOAT,       ss.matched] :
  ss.scan(/\d+/)          ? tokens << [:INTEGER,     ss.matched] :
  ss.scan(/\[/)           ? tokens << [:LIST_BEGIN,  ss.matched] :
  ss.scan(/\]/)           ? tokens << [:LISG_END,    ss.matched] :
  ss.scan(/{/)            ? tokens << [:TUPLE_BEGIN, ss.matched] :
  ss.scan(/}/)            ? tokens << [:TUPLE_END,   ss.matched] :
  ss.scan(/'([-\w_@]+)'/) ? tokens << [:ATOM,        ss[1]] :
  ss.scan(/[a-z][\w_@]+/) ? tokens << [:ATOM,        ss.matched] :
  ss.scan(/,/)            ? tokens << [:COMMA,       ss.matched] :
  ss.scan(/\s/)           ? nil :
  # ss.scan(/\s/)           ? tokens << [ss.matched,   ss.matched] :
  (pp tokens; p ss; raise "scanner error")
end

pp tokens
