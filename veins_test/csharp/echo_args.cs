#!/usr/bin/env -S fragletc --vein=csharp
using System;

class Program
{
    static void Main(string[] args)
    {
        Console.WriteLine("Args: " + string.Join(" ", args));
    }
}
