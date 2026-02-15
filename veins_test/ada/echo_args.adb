#!/usr/bin/env -S fragletc --vein=ada
Put ("Args: ");
for I in 1 .. Ada.Command_Line.Argument_Count loop
  if I > 1 then Put (" "); end if;
  Put (Ada.Command_Line.Argument (I));
end loop;
New_Line;
