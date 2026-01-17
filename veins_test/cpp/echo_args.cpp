#!/usr/bin/env -S fragletc --vein=cpp
int main(int argc, char* argv[]) {
    std::cout << "Args: ";
    for (int i = 1; i < argc; i++) {
        if (i > 1) {
            std::cout << " ";
        }
        std::cout << argv[i];
    }
    std::cout << std::endl;
    return 0;
}
