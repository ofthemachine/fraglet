#!/usr/bin/env -S fragletc --vein=visual-basic
Module Fraglet
    Sub Main(args As String())
        Console.WriteLine("Args: " & String.Join(" ", args))
    End Sub
End Module
