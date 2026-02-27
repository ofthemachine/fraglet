#!/usr/bin/env -S fragletc --vein=wren
import "os" for Process
System.print("Args: " + Process.arguments.join(" "))
