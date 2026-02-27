#!/usr/bin/env -S fragletc --vein=cpp
#include <iostream>
int main(int argc, char *argv[]) {
    if (argc > 1) std::cout << "First: " << argv[1] << std::endl;
    if (argc > 2) std::cout << "Second: " << argv[2] << std::endl;
    return 0;
}
