#!/usr/bin/env -S fragletc --vein=c
#include <stdio.h>
#include <string.h>
#include <ctype.h>

int main(void) {
    char line[1024];
    int line_num = 0;
    
    // C: Read lines, count characters, add "C:" prefix
    while (fgets(line, sizeof(line), stdin) != NULL) {
        line_num++;
        int len = strlen(line);
        // Remove newline from count
        if (len > 0 && line[len-1] == '\n') len--;
        printf("C[%d](len=%d): %s", line_num, len, line);
    }
    return 0;
}
