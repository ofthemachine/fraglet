#!/usr/bin/env -S fragletc --image=100hellos/ada:local
with Ada.Text_IO; use Ada.Text_IO;
with Ada.Command_Line; use Ada.Command_Line;

procedure Hello is
begin
  Put ("Args: ");
  for I in 1 .. Argument_Count loop
     if I > 1 then
        Put (" ");
     end if;
     Put (Argument (I));
  end loop;
  New_Line;
end Hello;
