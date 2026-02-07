#!/usr/bin/env -S fragletc --vein=c
#include <stdio.h>
int main(void) {
    int numbers[] = {1, 2, 3, 4, 5};
    int sum = 0;
    int i;
    for (i = 0; i < 5; i++) {
        sum += numbers[i];
    }
    printf("Array sum: %d\n", sum);
    return 0;
}
