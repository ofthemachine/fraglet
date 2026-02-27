#!/usr/bin/env -S fragletc --vein=pascal
var
  line: string;
begin
  while not eof do
  begin
    readln(line);
    writeln(upcase(line));
  end;
end.
