#!/usr/bin/env -S fragletc --vein=chapel
proc main(args: [] string) {
    writeln("Args: ", " ".join(args[1..]));
}
