#!/usr/bin/env -S fragletc --vein=php
echo "Args: " . implode(" ", array_slice($argv, 1)) . "\n";
