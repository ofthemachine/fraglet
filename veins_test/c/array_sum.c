#!/usr/bin/env -S fragletc --vein=c
int numbers[] = {1, 2, 3, 4, 5};
int sum = 0;
for (int i = 0; i < 5; i++) {
    sum += numbers[i];
}
printf("Array sum: %d\n", sum);
