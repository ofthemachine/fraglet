#!/usr/bin/env -S fragletc --image=100hellos/ada:local
with Ada.Text_IO; use Ada.Text_IO;

procedure Hello is
begin
  Put_Line ("Hello from fragment!");
end Hello;
