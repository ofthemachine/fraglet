#!/usr/bin/env -S fragletc --vein=php
echo strtoupper(trim(file_get_contents("php://stdin"))) . "\n";
