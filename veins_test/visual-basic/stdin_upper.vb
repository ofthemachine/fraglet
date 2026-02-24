#!/usr/bin/env -S fragletc --vein=visual-basic
Module Fraglet
    Sub Main(args As String())
        Console.WriteLine(Console.In.ReadToEnd().ToUpper())
    End Sub
End Module
