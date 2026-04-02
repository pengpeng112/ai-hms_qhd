#!/bin/bash
set -e
echo "=== Frontend Verification ==="
echo "--- npm run lint ---"
npm run lint
echo "PASS: npm run lint"
echo "--- npm run build ---"
npm run build
echo "PASS: npm run build"
echo "=== All frontend checks passed ==="
