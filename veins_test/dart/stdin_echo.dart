#!/usr/bin/env -S fragletc --vein=dart
import 'dart:io';
void main() {
  while (true) {
    var line = stdin.readLineSync();
    if (line == null) break;
    print(line.toUpperCase());
  }
}
