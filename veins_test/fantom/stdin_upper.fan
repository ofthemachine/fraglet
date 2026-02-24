#!/usr/bin/env -S fragletc --vein=fantom
class Fraglet {
  Void main() {
    in := Env.cur().in
    in.eachLine |line| { echo(line.upper) }
  }
}
