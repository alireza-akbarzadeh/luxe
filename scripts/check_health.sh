#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "--- Starting System Health Check ---"

# 1. Check MongoDB
echo -n "Checking MongoDB (mongo_bp)... "
if docker exec mongo_bp mongosh --username alirezas --password Aa.13741995 --eval "db.adminCommand('ping')" > /dev/null 2>&1; then
    echo -e "${GREEN}UP${NC}"
else
    echo -e "${RED}DOWN (Check credentials or startup status)${NC}"
fi

# 2. Check Elasticsearch
echo -n "Checking Elasticsearch... "
ES_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9200)
if [ "$ES_STATUS" == "200" ]; then
    echo -e "${GREEN}UP (HTTP 200)${NC}"
else
    echo -e "${RED}DOWN (Status: $ES_STATUS)${NC}"
fi

# 3. Check Kibana
echo -n "Checking Kibana UI... "
KIBANA_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:5601/api/status)
if [ "$KIBANA_STATUS" == "200" ]; then
    echo -e "${GREEN}UP${NC}"
else
    echo -e "${RED}NOT READY (Usually takes 1-2 mins to start)${NC}"
fi

# 4. Check Go API
echo -n "Checking Go API (/api/v1/health/ready)... "
API_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/health/ready)
if [ "$API_STATUS" == "200" ]; then
    echo -e "${GREEN}UP${NC}"
else
    echo -e "${RED}DOWN (Check Go container logs)${NC}"
fi

echo "--- Health Check Complete ---"