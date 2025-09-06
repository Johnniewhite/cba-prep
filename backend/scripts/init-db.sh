#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Initializing CBA Lite Database...${NC}"

# Wait for PostgreSQL to be ready
echo -e "${YELLOW}Waiting for PostgreSQL to be ready...${NC}"
until PGPASSWORD=cbaPREP2025 psql -h localhost -p 5433 -U cba -d cbalite -c '\q' 2>/dev/null; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 2
done

echo -e "${GREEN}PostgreSQL is ready!${NC}"

# Run migrations
echo -e "${GREEN}Running database migrations...${NC}"
if command -v migrate >/dev/null 2>&1; then
    PGPASSWORD=cbaPREP2025 migrate -path migrations -database "postgresql://cba:cbaPREP2025@localhost:5433/cbalite?sslmode=disable" up
    echo -e "${GREEN}Migrations completed successfully!${NC}"
else
    echo -e "${RED}migrate tool not installed. Please install it:${NC}"
    echo "go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

# Test connection
echo -e "${GREEN}Testing database connection...${NC}"
PGPASSWORD=cbaPREP2025 psql -h localhost -p 5433 -U cba -d cbalite -c "SELECT 'Database connection successful!' as message;"

echo -e "${GREEN}Database initialization complete!${NC}"