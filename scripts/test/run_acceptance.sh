#!/bin/bash
set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default values
KEEP_API_URL="${KEEP_API_URL:-http://localhost:8080}"
KEEP_API_KEY="${KEEP_API_KEY:-}"
TEST_PATTERN="${TEST_PATTERN:-""}"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --api-key)
      KEEP_API_KEY="$2"
      shift 2
      ;;
    --api-url)
      KEEP_API_URL="$2"
      shift 2
      ;;
    --pattern)
      TEST_PATTERN="$2"
      shift 2
      ;;
    *)
      echo "Unknown parameter: $1"
      exit 1
      ;;
  esac
done

# Check if API key is set
if [ -z "$KEEP_API_KEY" ]; then
  echo -e "${RED}Error: KEEP_API_KEY environment variable or --api-key argument is required${NC}"
  exit 1
fi

echo -e "${GREEN}Starting acceptance tests...${NC}"
echo "API URL: $KEEP_API_URL"
echo "Test pattern: ${TEST_PATTERN:-all tests}"

# Run the tests
TF_ACC=1 \
  KEEP_API_URL="$KEEP_API_URL" \
  KEEP_API_KEY="$KEEP_API_KEY" \
  go test -v -tags=acceptance \
  -timeout 30m \
  ${TEST_PATTERN:+"./test/acceptance/..." -run "$TEST_PATTERN"} \
  ${TEST_PATTERN:-"./test/acceptance/..."}

# Capture the exit code
test_exit_code=$?

if [ $test_exit_code -eq 0 ]; then
  echo -e "${GREEN}All tests passed!${NC}"
else
  echo -e "${RED}Some tests failed${NC}"
fi

exit $test_exit_code
