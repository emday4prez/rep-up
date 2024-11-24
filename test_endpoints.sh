#!/bin/bash
# test_endpoints.sh

# Set the base URL
BASE_URL="http://localhost:8080/api"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "Testing RepUp API endpoints..."

# Function to make requests and display results
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4

    echo -e "\n${GREEN}Testing: $description${NC}"
    echo "Method: $method"
    echo "Endpoint: $endpoint"
    
    if [ -n "$data" ]; then
        echo "Data: $data"
        response=$(curl -s -X "$method" "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -X "$method" "$BASE_URL$endpoint")
    fi

    echo -e "Response:\n$response\n"
}

# Test CREATE
echo -e "\n${GREEN}Testing CREATE operations${NC}"
make_request "POST" "/body-parts" \
    '{"name": "Chest"}' \
    "Create new body part"

# Test READ
echo -e "\n${GREEN}Testing READ operations${NC}"
make_request "GET" "/body-parts" "" "Get all body parts"
make_request "GET" "/body-parts/1" "" "Get specific body part"

# Test UPDATE
echo -e "\n${GREEN}Testing UPDATE operations${NC}"
make_request "PUT" "/body-parts/1" \
    '{"name": "Upper Chest"}' \
    "Update body part"

# Test DELETE
echo -e "\n${GREEN}Testing DELETE operations${NC}"
make_request "DELETE" "/body-parts/1" "" "Delete body part"

echo -e "\n${GREEN}All tests completed!${NC}"