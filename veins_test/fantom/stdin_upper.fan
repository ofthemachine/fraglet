#!/usr/bin/env -S fragletc --vein=fantom
class HelloWorld {
  Void main() {
    in := Env.cur().in
    in.eachLine |line| { echo(line.upper) }
  }
}
