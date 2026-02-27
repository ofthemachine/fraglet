#!/usr/bin/env -S fragletc --vein=verilog
initial begin
    $display("Args: foo bar baz");
    $finish;
end
