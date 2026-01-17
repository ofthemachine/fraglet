#!/usr/bin/env -S fragletc --vein=java

public static void main(String[] args) {
    int i = 0;
    for (String arg : args) {
        System.out.println(i++ + ": " + arg);
    }
}