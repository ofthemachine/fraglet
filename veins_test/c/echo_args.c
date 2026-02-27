#!/usr/bin/env -S fragletc --vein=c
#include <stdio.h>
int main(int argc, char *argv[]) {
    if (argc > 1) printf("First: %s\n", argv[1]);
    if (argc > 2) printf("Second: %s\n", argv[2]);
    return 0;
}
