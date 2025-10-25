#!/bin/bash

echo "Testing Docker service resolution for hashpost-pds..."

# Test Docker service resolution
echo -n "Testing hashpost-pds service resolution: "
if docker exec hashpost-pds-1 nslookup hashpost-pds > /dev/null 2>&1; then
    echo "✓ PASS"
else
    echo "✗ FAIL"
fi

# Test handle format
echo -n "Testing handle format: "
if docker exec hashpost-pds-1 echo "alice.hashpost-pds" | grep -q "alice.hashpost-pds"; then
    echo "✓ PASS"
else
    echo "✗ FAIL"
fi

echo ""
echo "If any tests failed, ensure Docker DNS is working:"
echo "docker-compose ps"
