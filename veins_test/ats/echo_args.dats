#!/usr/bin/env -S fragletc --vein=ats
implement main0{n}(argc, argv): void =
  let
    val args = listize_argc_argv(argc, argv)
  in
    list0_foreach(args, lam(arg) => println!(arg))
  end
