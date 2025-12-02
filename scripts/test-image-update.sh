#!/bin/bash

# Test script for image update tracking functionality

set -e

BASE_URL="${BASE_URL:-http://localhost:8090}"

echo "==================================="
echo "Testing Image Update Tracking"
echo "==================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Health check
echo -e "${BLUE}Test 1: Health Check${NC}"
curl -s "${BASE_URL}/health" | jq .
echo -e "${GREEN}✓ Health check passed${NC}"
echo ""

# Test 2: List endpoints
echo -e "${BLUE}Test 2: List Endpoints${NC}"
ENDPOINTS=$(curl -s "${BASE_URL}/api/v1/endpoints" | jq -r '.[].name' | head -1)
if [ -z "$ENDPOINTS" ]; then
    echo -e "${YELLOW}⚠ No endpoints found. Please create an endpoint first.${NC}"
    echo "Example:"
    echo 'curl -X POST http://localhost:8090/api/v1/endpoints \\'
    echo '  -H "Content-Type: application/json" \\'
    echo '  -d '\''{'
    echo '    "endpoint": "test-endpoint",'
    echo '    "specName": "cpu-1x",'
    echo '    "image": "nginx:latest",'
    echo '    "imagePrefix": "nginx:",'
    echo '    "replicas": 1'
    echo '  }'\'''
    exit 1
else
    echo "Found endpoint: $ENDPOINTS"
    echo -e "${GREEN}✓ List endpoints passed${NC}"
fi
echo ""

# Test 3: Check image for specific endpoint
echo -e "${BLUE}Test 3: Check Image Update for Endpoint${NC}"
echo "Checking endpoint: $ENDPOINTS"
RESULT=$(curl -s -X POST "${BASE_URL}/api/v1/endpoints/${ENDPOINTS}/check-image")
echo "$RESULT" | jq .
UPDATE_AVAILABLE=$(echo "$RESULT" | jq -r '.updateAvailable')
if [ "$UPDATE_AVAILABLE" = "true" ]; then
    echo -e "${YELLOW}⚠ Update available for endpoint: $ENDPOINTS${NC}"
else
    echo -e "${GREEN}✓ No update available${NC}"
fi
echo ""

# Test 4: Check all images
echo -e "${BLUE}Test 4: Check All Images${NC}"
RESULT=$(curl -s -X POST "${BASE_URL}/api/v1/endpoints/check-images")
echo "$RESULT" | jq .
UPDATES_FOUND=$(echo "$RESULT" | jq -r '.updatesFound')
echo -e "${GREEN}✓ Checked all images, found $UPDATES_FOUND updates${NC}"
echo ""

# Test 5: Get endpoint details (should show image tracking fields)
echo -e "${BLUE}Test 5: Get Endpoint Details${NC}"
curl -s "${BASE_URL}/api/v1/endpoints/${ENDPOINTS}" | jq '{
    name: .name,
    image: .image,
    imagePrefix: .imagePrefix,
    imageDigest: .imageDigest,
    imageUpdateAvailable: .imageUpdateAvailable,
    imageLastChecked: .imageLastChecked
}'
echo -e "${GREEN}✓ Got endpoint details${NC}"
echo ""

# Test 6: Simulate DockerHub webhook
echo -e "${BLUE}Test 6: Simulate DockerHub Webhook${NC}"
echo "This test simulates a DockerHub webhook notification"
echo "In production, configure this URL in DockerHub:"
echo "${BASE_URL}/api/v1/webhooks/dockerhub"
echo ""
echo "Example webhook payload from DockerHub:"
cat <<'EOF' | jq .
{
  "push_data": {
    "tag": "latest",
    "pushed_at": 1699123456,
    "pusher": "testuser"
  },
  "repository": {
    "name": "nginx",
    "namespace": "library",
    "repo_name": "nginx"
  }
}
EOF
echo ""

echo -e "${GREEN}==================================="
echo "All Tests Completed!"
echo "===================================${NC}"
echo ""
echo "Next steps:"
echo "1. Set FEISHU_WEBHOOK_URL environment variable to enable notifications"
echo "2. Configure DockerHub webhook: ${BASE_URL}/api/v1/webhooks/dockerhub"
echo "3. Set imagePrefix when creating endpoints for automatic tracking"
echo ""
