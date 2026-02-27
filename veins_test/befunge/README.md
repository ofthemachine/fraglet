# Befunge examples

| File | Description |
|------|-------------|
| `test.bf` | Hello World (classic reversed string). |
| `stdin_echo.bf` | Echo stdin to stdout (`~,@`). |
| `factorial.bf` | Computes 5! and prints `120`. |
| `squares.bf` | Prints squares 1²…5²: `1 4 9 16 25`. |

More examples (may need longer timeout or different interpreters):

- **Cat’s Eye** [Befunge-93/eg](https://github.com/catseye/Befunge-93/tree/master/eg): `fact.bf`, `calc.bf`, `euclid.bf`, `beer*.bf`, `cascade.bf`, etc.
- **FizzBuzz / Fibonacci**: various public snippets; some rely on put/get and can be slow or interpreter-dependent.

Run with:

```bash
fragletc --vein=befunge <file>.bf
# With stdin:
echo "input" | fragletc --vein=befunge stdin_echo.bf
```
