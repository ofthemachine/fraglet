#!/usr/bin/env -S fragletc --vein=pony
actor Main
  new create(env: Env) =>
    let args = env.args.slice(1)
    env.out.print("Args: " + " ".join(args.values()))
