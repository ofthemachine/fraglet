#!/usr/bin/env -S fragletc --vein=pony
actor Main
  new create(env: Env) =>
    env.out.print("Hello from fragment!")
