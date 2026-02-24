#!/usr/bin/env -S fragletc --vein=smalltalk
'Args: ' display. ((Smalltalk arguments) inject: '' into: [:acc :arg | (acc = '') ifTrue: [arg] ifFalse: [acc, ' ', arg]]) displayNl
