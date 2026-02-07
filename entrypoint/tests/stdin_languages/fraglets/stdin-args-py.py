import sys
args = sys.argv[1:]
print(f"args: {' '.join(args)}")
for line in sys.stdin:
    print(f"stdin: {line.strip()}")
