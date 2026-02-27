#!/usr/bin/env -S fragletc --vein=pascal
var
  i: integer;
begin
  write('Args:');
  for i := 1 to paramcount do
    write(' ', paramstr(i));
  writeln;
end.
