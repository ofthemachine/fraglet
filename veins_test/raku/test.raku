#!/usr/bin/env -S fragletc --vein=raku
class Greeting {
  has Str $.who;

  method greet() {
    say "Hello $!who!";
  }
}

my $greeting = Greeting.new(who => "World");
$greeting.greet();
