int main() {
    std::string line;
    while (std::getline(std::cin, line)) {
        for (auto& c : line) c = std::toupper(c);
        std::cout << line << std::endl;
    }
    return 0;
}
