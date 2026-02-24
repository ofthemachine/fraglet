#!/usr/bin/env -S fragletc --vein=fantom
class Fraglet {
  Void main() {
    args := Env.cur().args
    echo("Args: " + args.join(" "))
  }
}
