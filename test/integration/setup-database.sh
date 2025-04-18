#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Log function
log() {
  local level=$1
  local message=$2
  local color=$NC

  case $level in
    "INFO") color=$GREEN ;;
    "ERROR") color=$RED ;;
  esac

  echo -e "${color}[$level] $message${NC}"
}

# Check if PostgreSQL is installed
if ! command -v psql &> /dev/null; then
    log "ERROR" "PostgreSQL is not installed. Please install PostgreSQL first."
    exit 1
fi

# Database configuration
DB_NAME="libgo"
DB_USER="libgo"
DB_PASSWORD="libgo"

# Create database user if it doesn't exist
log "INFO" "Creating database user..."
sudo -u postgres psql -c "DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$DB_USER') THEN
    CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
  END IF;
END
\$\$;"

# Create database if it doesn't exist
log "INFO" "Creating database..."
sudo -u postgres psql -c "SELECT 'CREATE DATABASE $DB_NAME OWNER $DB_USER' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME')\gexec"

# Grant privileges
log "INFO" "Granting privileges..."
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"

log "INFO" "Database setup completed successfully!"
