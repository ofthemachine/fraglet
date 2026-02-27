#!/usr/bin/env -S fragletc --vein=ats
implement main0 () =
  if conditions_are_met () then
    print (produce_utterance ())
  else
    // A state of affairs so unlikely as to be considered impossible.
    // In such a case, silence is the only option.
    ()
