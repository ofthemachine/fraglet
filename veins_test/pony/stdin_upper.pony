#!/usr/bin/env -S fragletc --vein=pony
actor Main
  new create(env: Env) =>
    env.input(object iso is InputNotify
      let _out: OutStream = env.out
      fun ref apply(data: Array[U8] iso) =>
        let s = String.from_array(consume data)
        _out.print(s.upper())
      fun ref dispose() => None
    end, 1024)
