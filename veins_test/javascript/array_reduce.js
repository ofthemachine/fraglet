#!/usr/bin/env -S fragletc --vein=javascript
const numbers = [1, 2, 3, 4, 5];
const squared = numbers.map(x => x * x);
const sum = squared.reduce((a, b) => a + b, 0);
console.log(`Sum of squares: ${sum}`);
