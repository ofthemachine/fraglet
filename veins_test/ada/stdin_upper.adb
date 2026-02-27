#!/usr/bin/env -S fragletc --vein=ada
with Ada.Text_IO; use Ada.Text_IO;
with Ada.Characters.Handling;

procedure Fraglet is
  Line : String (1 .. 1024);
  Last : Natural;
begin
  while not End_Of_File loop
    Get_Line (Line, Last);
    Put_Line (Ada.Characters.Handling.To_Upper (Line (1 .. Last)));
  end loop;
end Fraglet;
