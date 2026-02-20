#!/usr/bin/env -S fragletc --vein=c
#include <stdio.h>
int main(int argc, char *argv[]) {
	int i;
	printf("Args:");
	for (i = 1; i < argc; i++)
		printf(" %s", argv[i]);
	printf("\n");
	return 0;
}
