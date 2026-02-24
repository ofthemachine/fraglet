#!/usr/bin/env -S fragletc --vein=smalltalk
[ stdin atEnd ]
  whileFalse: [ (stdin nextLine) asUppercase displayNl ]
