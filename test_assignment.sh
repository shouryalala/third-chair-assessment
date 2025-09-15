#!/bin/bash

set -e

echo "🧪 Instagram User Processor - Assignment Test Suite"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"
API_BASE="$BASE_URL/api/v1"

# Test results tracking
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to run tests
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_status="$3"

    echo -n "Testing: $test_name ... "

    if [ "$expected_status" = "200" ]; then
        if response=$(eval "$command" 2>/dev/null) && echo "$response" | jq . >/dev/null 2>&1; then
            echo -e "${GREEN}PASS${NC}"
            ((TESTS_PASSED++))
            return 0
        else
            echo -e "${RED}FAIL${NC}"
            echo "  Response: $response"
            ((TESTS_FAILED++))
            return 1
        fi
    else
        # For non-200 status codes, just check if command runs
        if eval "$command" >/dev/null 2>&1; then
            echo -e "${GREEN}PASS${NC}"
            ((TESTS_PASSED++))
            return 0
        else
            echo -e "${RED}FAIL${NC}"
            ((TESTS_FAILED++))
            return 1
        fi
    fi
}

# Wait for server to be ready
echo "🔍 Checking if server is running..."
for i in {1..30}; do
    if curl -s "$BASE_URL/health" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Server is ready!${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}❌ Server not ready after 30 seconds${NC}"
        echo "Please start the server with: go run cmd/server/main.go"
        exit 1
    fi
    sleep 1
done

echo ""
echo "📋 Running Basic API Tests"
echo "========================="

# Test 1: Health check
run_test "Health Check" "curl -s $BASE_URL/health" "200"

# Test 2: Single user endpoint (should work)
run_test "Single User API" "curl -s $API_BASE/instagram/user/musiclover2024" "200"

# Test 3: Single user with non-existent user
run_test "Non-existent User" "curl -s $API_BASE/instagram/user/nonexistentuser123456" "200"

echo ""
echo "🚀 Testing Batch Processing Implementation"
echo "========================================"

# Test 4: Batch processing endpoint - small batch
BATCH_RESPONSE=$(curl -s -X POST "$API_BASE/instagram/users/batch" \
    -H "Content-Type: application/json" \
    -d '{
        "usernames": ["musiclover2024", "techguru_sarah"],
        "max_concurrency": 2
    }' 2>/dev/null)

if echo "$BATCH_RESPONSE" | jq . >/dev/null 2>&1; then
    echo -e "Small Batch Request: ${GREEN}PASS${NC}"
    ((TESTS_PASSED++))

    # Extract job_id for follow-up tests
    JOB_ID=$(echo "$BATCH_RESPONSE" | jq -r '.job_id // empty')
    echo "  Job ID: $JOB_ID"

    # Check if response has expected fields
    if echo "$BATCH_RESPONSE" | jq -e '.status, .total_users, .processed_users' >/dev/null 2>&1; then
        echo -e "  Response Structure: ${GREEN}VALID${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "  Response Structure: ${RED}INVALID${NC}"
        echo "  Expected fields: job_id, status, total_users, processed_users, successful_users, failed_users"
        echo "  Got: $BATCH_RESPONSE"
        ((TESTS_FAILED++))
    fi
else
    echo -e "Small Batch Request: ${RED}FAIL${NC}"
    echo "  Response: $BATCH_RESPONSE"
    ((TESTS_FAILED++))
fi

# Test 5: Larger batch
echo -n "Testing: Large Batch Processing ... "
LARGE_BATCH_RESPONSE=$(curl -s -X POST "$API_BASE/instagram/users/batch" \
    -H "Content-Type: application/json" \
    -d '{
        "usernames": ["musiclover2024", "techguru_sarah", "foodie_adventures", "fitness_jenny", "artist_david"],
        "max_concurrency": 3
    }' 2>/dev/null)

if echo "$LARGE_BATCH_RESPONSE" | jq . >/dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    ((TESTS_PASSED++))

    LARGE_JOB_ID=$(echo "$LARGE_BATCH_RESPONSE" | jq -r '.job_id // empty')
    echo "  Large Job ID: $LARGE_JOB_ID"
else
    echo -e "${RED}FAIL${NC}"
    echo "  Response: $LARGE_BATCH_RESPONSE"
    ((TESTS_FAILED++))
fi

# Test 6: Invalid request (missing fields)
echo -n "Testing: Invalid Request Handling ... "
INVALID_RESPONSE=$(curl -s -X POST "$API_BASE/instagram/users/batch" \
    -H "Content-Type: application/json" \
    -d '{"usernames": []}' 2>/dev/null)

if echo "$INVALID_RESPONSE" | grep -q "error\|invalid" 2>/dev/null; then
    echo -e "${GREEN}PASS${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}FAIL${NC}"
    echo "  Should return error for empty usernames"
    echo "  Response: $INVALID_RESPONSE"
    ((TESTS_FAILED++))
fi

echo ""
echo "🗄️ Database Integration Tests"
echo "============================"

# Check if we have database access
if command -v docker >/dev/null 2>&1 && docker-compose ps postgres | grep -q "Up"; then
    echo "Testing database queries..."

    # Test database connection and sample data
    USER_COUNT=$(docker-compose exec -T postgres psql -U user -d instagram_processor -t -c "SELECT COUNT(*) FROM instagram_users;" 2>/dev/null | tr -d ' ')

    if [[ "$USER_COUNT" =~ ^[0-9]+$ ]] && [ "$USER_COUNT" -gt 0 ]; then
        echo -e "Database Connection: ${GREEN}PASS${NC} ($USER_COUNT users)"
        ((TESTS_PASSED++))
    else
        echo -e "Database Connection: ${RED}FAIL${NC}"
        ((TESTS_FAILED++))
    fi

    # Test complex query
    if docker-compose exec -T postgres psql -U user -d instagram_processor -c "SELECT * FROM get_top_users_by_engagement(3);" >/dev/null 2>&1; then
        echo -e "Complex Query: ${GREEN}PASS${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "Complex Query: ${RED}FAIL${NC}"
        ((TESTS_FAILED++))
    fi

    # Test job tracking table
    if docker-compose exec -T postgres psql -U user -d instagram_processor -c "SELECT * FROM processing_jobs LIMIT 1;" >/dev/null 2>&1; then
        echo -e "Job Tracking Table: ${GREEN}PASS${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "Job Tracking Table: ${RED}FAIL${NC}"
        ((TESTS_FAILED++))
    fi

else
    echo -e "${YELLOW}⚠️ Database tests skipped (Docker not available or database not running)${NC}"
fi

echo ""
echo "⏱️ Performance & Concurrency Tests"
echo "================================="

# Test concurrent requests
echo "Testing concurrent batch requests..."
CONCURRENT_PIDS=()

for i in {1..3}; do
    (
        curl -s -X POST "$API_BASE/instagram/users/batch" \
            -H "Content-Type: application/json" \
            -d "{
                \"usernames\": [\"test_user_$i\", \"musiclover2024\"],
                \"max_concurrency\": 2
            }" > "/tmp/concurrent_test_$i.json" 2>/dev/null
    ) &
    CONCURRENT_PIDS+=($!)
done

# Wait for all concurrent requests
CONCURRENT_SUCCESS=0
for pid in "${CONCURRENT_PIDS[@]}"; do
    if wait $pid; then
        ((CONCURRENT_SUCCESS++))
    fi
done

if [ $CONCURRENT_SUCCESS -eq 3 ]; then
    echo -e "Concurrent Requests: ${GREEN}PASS${NC} ($CONCURRENT_SUCCESS/3 succeeded)"
    ((TESTS_PASSED++))
else
    echo -e "Concurrent Requests: ${RED}FAIL${NC} ($CONCURRENT_SUCCESS/3 succeeded)"
    ((TESTS_FAILED++))
fi

# Cleanup temp files
rm -f /tmp/concurrent_test_*.json

echo ""
echo "📊 Test Results Summary"
echo "======================"
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo -e "Total Tests: $((TESTS_PASSED + TESTS_FAILED))"

if [ $TESTS_FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 All tests passed! Your implementation looks good.${NC}"
    echo ""
    echo "Next steps to verify your implementation:"
    echo "1. Check the database for processed users"
    echo "2. Monitor logs for proper rate limiting"
    echo "3. Test with larger batches (20+ users)"
    echo "4. Verify error handling with invalid usernames"
    exit 0
else
    echo ""
    echo -e "${RED}❌ Some tests failed. Please check your implementation.${NC}"
    echo ""
    echo "Common issues:"
    echo "- BatchProcessUsersHandler not implemented yet"
    echo "- Response format doesn't match expected structure"
    echo "- Database not properly updated during batch processing"
    echo "- Error handling needs improvement"
    exit 1
fi