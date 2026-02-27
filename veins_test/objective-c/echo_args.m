#!/usr/bin/env -S fragletc --vein=objective-c
#import <stdio.h>
int main(int argc, char *argv[]) {
    printf("Args:");
    int i;
    for (i = 1; i < argc; i++)
        printf(" %s", argv[i]);
    printf("\n");
    return 0;
}
