#!/usr/bin/env -S fragletc --vein=cpp
#include <iostream>
#include <cctype>
int main() {
    char c;
    while (std::cin.get(c)) std::cout << static_cast<char>(std::toupper(static_cast<unsigned char>(c)));
    return 0;
}
