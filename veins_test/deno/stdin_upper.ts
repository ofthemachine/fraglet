#!/usr/bin/env -S fragletc --vein=deno
const buf = new Uint8Array(1024);
while (true) {
  const n = await Deno.stdin.read(buf);
  if (n === null) break;
  const s = new TextDecoder().decode(buf.subarray(0, n));
  console.log(s.trim().toUpperCase());
}
