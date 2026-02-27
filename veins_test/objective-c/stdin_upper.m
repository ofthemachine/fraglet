#!/usr/bin/env -S fragletc --vein=objective-c
#import <Foundation/Foundation.h>
#import <stdio.h>
int main(int argc, char *argv[]) {
    char *line = NULL;
    size_t n = 0;
    while (getline(&line, &n, stdin) != -1) {
        NSString *s = [NSString stringWithUTF8String:line];
        printf("%s\n", [[s uppercaseString] UTF8String]);
    }
    return 0;
}
