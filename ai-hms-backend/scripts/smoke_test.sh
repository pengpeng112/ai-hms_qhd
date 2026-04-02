#!/bin/bash
# API Smoke Test for Phase 1 (Login + Patient List + Patient Detail)
# Usage: ./smoke_test.sh [BASE_URL]
# Preconditions: seed_phase1.sql has been applied (test_admin user + sample patients exist)
set -euo pipefail

BASE_URL=${1:-http://localhost:8080}
PASS=0
FAIL=0
TOKEN=""
FIRST_PATIENT_ID=""

ok()   { echo "  ✅ PASS: $1"; ((PASS++)); }
fail() { echo "  ❌ FAIL: $1"; ((FAIL++)); }

# ---------- Test 1: Health check ----------
echo "=== 1. Health Check ==="
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
[ "$HTTP_CODE" = "200" ] && ok "Health check ($HTTP_CODE)" || fail "Health check ($HTTP_CODE)"

# ---------- Test 2: Login ----------
echo "=== 2. Login (test_admin / Test@123456) ==="
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"test_admin","password":"Test@123456"}')
TOKEN=$(echo "$LOGIN_RESP" | grep -o '"token":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$TOKEN" ]; then
  ok "Login succeeded, token obtained"
else
  fail "Login failed. Response: $LOGIN_RESP"
fi

# ---------- Test 3: Token validation (GET /me) ----------
echo "=== 3. Token Validation (GET /me) ==="
ME_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/me")
[ "$ME_CODE" = "200" ] && ok "GET /me ($ME_CODE)" || fail "GET /me ($ME_CODE)"

# ---------- Test 4: Patient list ----------
echo "=== 4. Patient List ==="
PATIENT_LIST_RESP=$(curl -s \
  -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/patients?page=1&pageSize=20")
PATIENT_LIST_CODE=$(echo "$PATIENT_LIST_RESP" | head -c0; curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/patients?page=1&pageSize=20")
if [ "$PATIENT_LIST_CODE" = "200" ]; then
  ok "Patient list ($PATIENT_LIST_CODE)"
  # Extract first patient ID for detail test
  FIRST_PATIENT_ID=$(echo "$PATIENT_LIST_RESP" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  if [ -z "$FIRST_PATIENT_ID" ]; then
    # Try numeric ID format
    FIRST_PATIENT_ID=$(echo "$PATIENT_LIST_RESP" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
  fi
  echo "  ℹ️  First patient ID: ${FIRST_PATIENT_ID:-<none found>}"
else
  fail "Patient list ($PATIENT_LIST_CODE)"
fi

# ---------- Test 5: Patient detail ----------
echo "=== 5. Patient Detail ==="
if [ -n "$FIRST_PATIENT_ID" ]; then
  DETAIL_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer $TOKEN" \
    "$BASE_URL/api/v1/patients/$FIRST_PATIENT_ID")
  [ "$DETAIL_CODE" = "200" ] && ok "Patient detail ($DETAIL_CODE)" || fail "Patient detail ($DETAIL_CODE)"
else
  fail "Patient detail (skipped — no patient ID from list)"
fi

# ---------- Test 6: Patient core (used by PatientDetail.tsx) ----------
echo "=== 6. Patient Core ==="
if [ -n "$FIRST_PATIENT_ID" ]; then
  CORE_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer $TOKEN" \
    "$BASE_URL/api/v1/patients/$FIRST_PATIENT_ID/core")
  [ "$CORE_CODE" = "200" ] && ok "Patient core ($CORE_CODE)" || fail "Patient core ($CORE_CODE)"
else
  fail "Patient core (skipped — no patient ID from list)"
fi

# ---------- Test 7: Unauthorized access (no token) ----------
echo "=== 7. Unauthorized Access ==="
UNAUTH_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/patients?page=1&pageSize=20")
if [ "$UNAUTH_CODE" = "401" ] || [ "$UNAUTH_CODE" = "403" ]; then
  ok "Unauthorized access blocked ($UNAUTH_CODE)"
else
  fail "Unauthorized access NOT blocked ($UNAUTH_CODE) — expected 401/403"
fi

# ---------- Test 8: Wrong credentials ----------
echo "=== 8. Wrong Credentials ==="
BAD_LOGIN_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"test_admin","password":"wrong_password"}')
[ "$BAD_LOGIN_CODE" = "401" ] && ok "Wrong password rejected ($BAD_LOGIN_CODE)" || fail "Wrong password NOT rejected ($BAD_LOGIN_CODE)"

# ---------- Summary ----------
echo ""
echo "========================================="
echo " Results: $PASS passed, $FAIL failed"
echo "========================================="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
