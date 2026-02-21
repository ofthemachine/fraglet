#!/usr/bin/env -S fragletc --vein=java
import java.util.*;

public class Fraglet {
    public static void main(String[] args) {
        java.util.Scanner s = new java.util.Scanner(System.in);
        while (s.hasNextLine()) {
            System.out.println(s.nextLine().toUpperCase());
        }
    }
}
