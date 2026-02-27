#!/usr/bin/env -S fragletc --vein=octave
args = argv();
printf("Args:");
for i = 1:length(args)
    printf(" %s", args{i});
end
printf("\n");
