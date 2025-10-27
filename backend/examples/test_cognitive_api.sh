#!/bin/bash

# Erebus Cognitive Engine API Test Script
# This script demonstrates the cognitive API endpoints

set -e

BASE_URL="http://localhost:8080"
TENANT_ID="test-tenant-$(date +%s%N)"

# Cleanup function
cleanup() {
    echo "Cleaning up tenant resources..."
    curl -s -X DELETE "$BASE_URL/api/cognitive/tenants/$TENANT_ID" >/dev/null 2>&1 || true
}

# Set trap for cleanup on exit
trap cleanup EXIT

echo "=== Erebus Cognitive Engine API Demo ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper function
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    echo -e "${BLUE}$method $endpoint${NC}"
    if [ -n "$data" ]; then
        curl -s -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data" | jq '.'
    else
        curl -s -X $method "$BASE_URL$endpoint" | jq '.'
    fi
    echo ""
}

# 1. Check health
echo -e "${GREEN}1. Health Check${NC}"
api_call GET "/api/cognitive/health"

# 2. Initialize tenant
echo -e "${GREEN}2. Initialize Tenant${NC}"
api_call POST "/api/cognitive/tenants/$TENANT_ID/init"

# 3. Create concepts
echo -e "${GREEN}3. Create Concepts${NC}"

echo "Creating 'Cat' concept..."
CAT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/cognitive/tenants/$TENANT_ID/concepts" \
    -H "Content-Type: application/json" \
    -d '{"name": "Cat"}')
CAT_ID=$(echo $CAT_RESPONSE | jq -r '.atom_id')
echo $CAT_RESPONSE | jq '.'
echo "Cat ID: $CAT_ID"
echo ""

echo "Creating 'Dog' concept..."
DOG_RESPONSE=$(curl -s -X POST "$BASE_URL/api/cognitive/tenants/$TENANT_ID/concepts" \
    -H "Content-Type: application/json" \
    -d '{"name": "Dog"}')
DOG_ID=$(echo $DOG_RESPONSE | jq -r '.atom_id')
echo $DOG_RESPONSE | jq '.'
echo "Dog ID: $DOG_ID"
echo ""

echo "Creating 'Mammal' concept..."
MAMMAL_RESPONSE=$(curl -s -X POST "$BASE_URL/api/cognitive/tenants/$TENANT_ID/concepts" \
    -H "Content-Type: application/json" \
    -d '{"name": "Mammal"}')
MAMMAL_ID=$(echo $MAMMAL_RESPONSE | jq -r '.atom_id')
echo $MAMMAL_RESPONSE | jq '.'
echo "Mammal ID: $MAMMAL_ID"
echo ""

echo "Creating 'Animal' concept..."
ANIMAL_RESPONSE=$(curl -s -X POST "$BASE_URL/api/cognitive/tenants/$TENANT_ID/concepts" \
    -H "Content-Type: application/json" \
    -d '{"name": "Animal"}')
ANIMAL_ID=$(echo $ANIMAL_RESPONSE | jq -r '.atom_id')
echo $ANIMAL_RESPONSE | jq '.'
echo "Animal ID: $ANIMAL_ID"
echo ""

# 4. Query atoms
echo -e "${GREEN}4. Query All Atoms${NC}"
api_call GET "/api/cognitive/tenants/$TENANT_ID/atoms"

# 5. Create inheritance links
echo -e "${GREEN}5. Create Inheritance Links${NC}"

echo "Creating Cat → Mammal link..."
api_call POST "/api/cognitive/tenants/$TENANT_ID/links/inheritance" \
    "{\"source_id\": \"$CAT_ID\", \"target_id\": \"$MAMMAL_ID\"}"

echo "Creating Dog → Mammal link..."
api_call POST "/api/cognitive/tenants/$TENANT_ID/links/inheritance" \
    "{\"source_id\": \"$DOG_ID\", \"target_id\": \"$MAMMAL_ID\"}"

echo "Creating Mammal → Animal link..."
api_call POST "/api/cognitive/tenants/$TENANT_ID/links/inheritance" \
    "{\"source_id\": \"$MAMMAL_ID\", \"target_id\": \"$ANIMAL_ID\"}"

# 6. Get a specific atom
echo -e "${GREEN}6. Get Specific Atom (Cat)${NC}"
api_call GET "/api/cognitive/tenants/$TENANT_ID/atoms/$CAT_ID"

# 7. Update atom truth value
echo -e "${GREEN}7. Update Atom Truth Value${NC}"
api_call PUT "/api/cognitive/tenants/$TENANT_ID/atoms/$CAT_ID" \
    '{"strength": 0.95, "confidence": 0.9}'

# 8. Query by type
echo -e "${GREEN}8. Query Atoms by Type${NC}"
api_call GET "/api/cognitive/tenants/$TENANT_ID/atoms?type=concept"

# 9. Run inference
echo -e "${GREEN}9. Run Inference${NC}"
api_call POST "/api/cognitive/tenants/$TENANT_ID/inference" \
    '{"max_iterations": 5}'

# 10. Query atoms after inference
echo -e "${GREEN}10. Query Atoms After Inference${NC}"
api_call GET "/api/cognitive/tenants/$TENANT_ID/atoms"

# 11. Create pipeline
echo -e "${GREEN}11. Create Cognitive Pipeline${NC}"
PIPELINE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/cognitive/tenants/$TENANT_ID/pipelines" \
    -H "Content-Type: application/json" \
    -d '{"name": "test-pipeline", "use_default": true}')
PIPELINE_ID=$(echo $PIPELINE_RESPONSE | jq -r '.pipeline_id')
echo $PIPELINE_RESPONSE | jq '.'
echo "Pipeline ID: $PIPELINE_ID"
echo ""

# 12. Get pipeline details
echo -e "${GREEN}12. Get Pipeline Details${NC}"
api_call GET "/api/cognitive/tenants/$TENANT_ID/pipelines/$PIPELINE_ID"

# 13. Execute pipeline
echo -e "${GREEN}13. Execute Pipeline${NC}"
api_call POST "/api/cognitive/tenants/$TENANT_ID/pipelines/$PIPELINE_ID/execute"

# 14. Get agents
echo -e "${GREEN}14. Get Tenant Agents${NC}"
api_call GET "/api/cognitive/tenants/$TENANT_ID/agents"

# 15. Get tenant statistics
echo -e "${GREEN}15. Get Tenant Statistics${NC}"
api_call GET "/api/cognitive/tenants/$TENANT_ID/stats"

# 16. Get global statistics
echo -e "${GREEN}16. Get Global Statistics${NC}"
api_call GET "/api/cognitive/stats"

echo -e "${GREEN}=== Demo Complete ===${NC}"
echo ""
echo "Tenant ID used: $TENANT_ID"
echo "You can query this tenant's data at any time using the API endpoints."
