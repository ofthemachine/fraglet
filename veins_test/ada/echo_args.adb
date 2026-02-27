#!/usr/bin/env -S fragletc --vein=ada
with Ada.Text_IO; use Ada.Text_IO;
with Ada.Command_Line;

procedure Fraglet is
begin
  Put ("Args:");
  for I in 1 .. Ada.Command_Line.Argument_Count loop
    Put (" " & Ada.Command_Line.Argument (I));
  end loop;
  New_Line;
end Fraglet;
