#!/usr/bin/env -S fragletc --vein=java --mode=wordalytica
WordSet<?> words = Wordalytica.loadWords();
int n = words.endingWith("ing").count();
System.out.println("Count: " + n);
