#!/usr/bin/env -S fragletc --vein=scala
object Main {
  def main(args: Array[String]): Unit = {
    scala.io.Source.stdin.getLines().foreach(line => println(line.toUpperCase))
  }
}
