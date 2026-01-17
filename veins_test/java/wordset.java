#!/usr/bin/env -S fragletc --vein=java:wordalytica

WordSet<?> words = HelloWorld.loadWords();
int count = words.endingWith("ing").count();
System.out.println("Words ending with 'ing': " + count);
