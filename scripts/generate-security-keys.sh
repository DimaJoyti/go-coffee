#!/bin/bash

# Security Keys Generation Script for Go Coffee Platform
# This script generates secure cryptographic keys for the security system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔐 Go Coffee Security Keys Generator${NC}"
echo "=================================================="

# Create keys directory if it doesn't exist
KEYS_DIR="./keys"
mkdir -p "$KEYS_DIR"

# Function to generate random string
generate_random_string() {
    local length=$1
    openssl rand -base64 $((length * 3 / 4)) | tr -d "=+/" | cut -c1-${length}
}

# Function to generate random hex string
generate_random_hex() {
    local length=$1
    openssl rand -hex $((length / 2))
}

echo -e "\n${YELLOW}📋 Generating Cryptographic Keys...${NC}"

# 1. Generate AES-256 Key (32 bytes, base64 encoded)
echo "🔑 Generating AES-256 encryption key..."
AES_KEY=$(openssl rand -base64 32)
echo "AES_KEY=$AES_KEY" > "$KEYS_DIR/aes.key"
echo -e "${GREEN}✅ AES-256 key generated${NC}"

# 2. Generate RSA Key Pair (2048-bit)
echo "🔑 Generating RSA-2048 key pair..."
openssl genrsa -out "$KEYS_DIR/rsa_private.pem" 2048 2>/dev/null
openssl rsa -in "$KEYS_DIR/rsa_private.pem" -pubout -out "$KEYS_DIR/rsa_public.pem" 2>/dev/null

# Convert to single line format for environment variables
RSA_PRIVATE_KEY=$(awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}' "$KEYS_DIR/rsa_private.pem")
RSA_PUBLIC_KEY=$(awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}' "$KEYS_DIR/rsa_public.pem")

echo "RSA_PRIVATE_KEY=\"$RSA_PRIVATE_KEY\"" > "$KEYS_DIR/rsa_private.env"
echo "RSA_PUBLIC_KEY=\"$RSA_PUBLIC_KEY\"" > "$KEYS_DIR/rsa_public.env"
echo -e "${GREEN}✅ RSA-2048 key pair generated${NC}"

# 3. Generate JWT Secret (64 characters)
echo "🔑 Generating JWT secret..."
JWT_SECRET=$(generate_random_string 64)
echo "JWT_SECRET=$JWT_SECRET" > "$KEYS_DIR/jwt.key"
echo -e "${GREEN}✅ JWT secret generated${NC}"

# 4. Generate Webhook Secret (32 characters)
echo "🔑 Generating webhook secret..."
WEBHOOK_SECRET=$(generate_random_string 32)
echo "WEBHOOK_SECRET=$WEBHOOK_SECRET" > "$KEYS_DIR/webhook.key"
echo -e "${GREEN}✅ Webhook secret generated${NC}"

# 5. Generate Database Encryption Key (32 bytes, hex)
echo "🔑 Generating database encryption key..."
DB_ENCRYPTION_KEY=$(generate_random_hex 64)
echo "DB_ENCRYPTION_KEY=$DB_ENCRYPTION_KEY" > "$KEYS_DIR/db_encryption.key"
echo -e "${GREEN}✅ Database encryption key generated${NC}"

# 6. Generate Session Secret (32 characters)
echo "🔑 Generating session secret..."
SESSION_SECRET=$(generate_random_string 32)
echo "SESSION_SECRET=$SESSION_SECRET" > "$KEYS_DIR/session.key"
echo -e "${GREEN}✅ Session secret generated${NC}"

# 7. Generate API Keys for different services
echo "🔑 Generating API keys..."
AUTH_API_KEY=$(generate_random_string 32)
PAYMENT_API_KEY=$(generate_random_string 32)
ORDER_API_KEY=$(generate_random_string 32)
USER_API_KEY=$(generate_random_string 32)

cat > "$KEYS_DIR/api_keys.env" << EOF
AUTH_API_KEY=$AUTH_API_KEY
PAYMENT_API_KEY=$PAYMENT_API_KEY
ORDER_API_KEY=$ORDER_API_KEY
USER_API_KEY=$USER_API_KEY
EOF
echo -e "${GREEN}✅ API keys generated${NC}"

# 8. Generate Redis Password
echo "🔑 Generating Redis password..."
REDIS_PASSWORD=$(generate_random_string 24)
echo "REDIS_PASSWORD=$REDIS_PASSWORD" > "$KEYS_DIR/redis.key"
echo -e "${GREEN}✅ Redis password generated${NC}"

# 9. Generate Database Passwords
echo "🔑 Generating database passwords..."
DB_PASSWORD=$(generate_random_string 24)
DB_ROOT_PASSWORD=$(generate_random_string 32)

cat > "$KEYS_DIR/database.env" << EOF
DATABASE_PASSWORD=$DB_PASSWORD
DATABASE_ROOT_PASSWORD=$DB_ROOT_PASSWORD
EOF
echo -e "${GREEN}✅ Database passwords generated${NC}"

# 10. Generate Monitoring Passwords
echo "🔑 Generating monitoring passwords..."
GRAFANA_ADMIN_PASSWORD=$(generate_random_string 16)
PROMETHEUS_PASSWORD=$(generate_random_string 16)

cat > "$KEYS_DIR/monitoring.env" << EOF
GRAFANA_ADMIN_PASSWORD=$GRAFANA_ADMIN_PASSWORD
PROMETHEUS_PASSWORD=$PROMETHEUS_PASSWORD
EOF
echo -e "${GREEN}✅ Monitoring passwords generated${NC}"

# Create comprehensive .env file
echo -e "\n${YELLOW}📄 Creating comprehensive .env file...${NC}"

cat > "$KEYS_DIR/.env.generated" << EOF
# Generated Security Keys for Go Coffee Platform
# Generated on: $(date)
# WARNING: Keep these keys secure and never commit to version control!

# =============================================================================
# ENCRYPTION KEYS
# =============================================================================

# AES-256 Encryption Key (Base64)
AES_KEY=$AES_KEY

# RSA Key Pair (PEM format, single line)
RSA_PRIVATE_KEY="$RSA_PRIVATE_KEY"
RSA_PUBLIC_KEY="$RSA_PUBLIC_KEY"

# =============================================================================
# AUTHENTICATION & AUTHORIZATION
# =============================================================================

# JWT Secret (64 characters)
JWT_SECRET=$JWT_SECRET

# Session Secret (32 characters)
SESSION_SECRET=$SESSION_SECRET

# =============================================================================
# API KEYS
# =============================================================================

# Service API Keys
AUTH_API_KEY=$AUTH_API_KEY
PAYMENT_API_KEY=$PAYMENT_API_KEY
ORDER_API_KEY=$ORDER_API_KEY
USER_API_KEY=$USER_API_KEY

# Webhook Secret
WEBHOOK_SECRET=$WEBHOOK_SECRET

# =============================================================================
# DATABASE CREDENTIALS
# =============================================================================

# Database Passwords
DATABASE_PASSWORD=$DB_PASSWORD
DATABASE_ROOT_PASSWORD=$DB_ROOT_PASSWORD

# Database Encryption Key
DB_ENCRYPTION_KEY=$DB_ENCRYPTION_KEY

# =============================================================================
# CACHE & SESSION STORE
# =============================================================================

# Redis Password
REDIS_PASSWORD=$REDIS_PASSWORD

# =============================================================================
# MONITORING & OBSERVABILITY
# =============================================================================

# Monitoring Passwords
GRAFANA_ADMIN_PASSWORD=$GRAFANA_ADMIN_PASSWORD
PROMETHEUS_PASSWORD=$PROMETHEUS_PASSWORD

# =============================================================================
# SECURITY NOTES
# =============================================================================

# 1. These keys are randomly generated and cryptographically secure
# 2. Store these keys in a secure location (e.g., password manager, vault)
# 3. Rotate these keys regularly (every 90 days minimum)
# 4. Never commit these keys to version control
# 5. Use different keys for different environments (dev, staging, prod)
# 6. Monitor key usage and access
# 7. Have a key rotation strategy in place
# 8. Backup keys securely with proper encryption

EOF

echo -e "${GREEN}✅ Comprehensive .env file created${NC}"

# Create key rotation script
echo -e "\n${YELLOW}🔄 Creating key rotation script...${NC}"

cat > "$KEYS_DIR/rotate-keys.sh" << 'EOF'
#!/bin/bash

# Key Rotation Script
# This script helps rotate security keys safely

echo "🔄 Key Rotation Script"
echo "======================"

echo "⚠️  WARNING: This will generate new keys and invalidate existing ones!"
echo "Make sure you have a backup and deployment plan before proceeding."
echo ""

read -p "Are you sure you want to rotate all keys? (yes/no): " confirm

if [ "$confirm" = "yes" ]; then
    echo "🔄 Rotating keys..."
    
    # Backup existing keys
    if [ -f ".env.generated" ]; then
        cp .env.generated ".env.backup.$(date +%Y%m%d_%H%M%S)"
        echo "✅ Existing keys backed up"
    fi
    
    # Generate new keys
    cd ..
    ./generate-security-keys.sh
    
    echo "✅ New keys generated"
    echo "📋 Next steps:"
    echo "1. Update your deployment configuration"
    echo "2. Deploy services with new keys"
    echo "3. Verify all services are working"
    echo "4. Remove old key backups securely"
else
    echo "❌ Key rotation cancelled"
fi
EOF

chmod +x "$KEYS_DIR/rotate-keys.sh"
echo -e "${GREEN}✅ Key rotation script created${NC}"

# Create key validation script
echo -e "\n${YELLOW}🔍 Creating key validation script...${NC}"

cat > "$KEYS_DIR/validate-keys.sh" << 'EOF'
#!/bin/bash

# Key Validation Script
# This script validates the generated keys

echo "🔍 Key Validation Script"
echo "========================"

# Source the generated environment file
if [ -f ".env.generated" ]; then
    source .env.generated
else
    echo "❌ .env.generated file not found"
    exit 1
fi

# Validate AES key
if [ ${#AES_KEY} -eq 44 ]; then
    echo "✅ AES key length is correct (44 characters base64)"
else
    echo "❌ AES key length is incorrect (expected 44, got ${#AES_KEY})"
fi

# Validate JWT secret
if [ ${#JWT_SECRET} -eq 64 ]; then
    echo "✅ JWT secret length is correct (64 characters)"
else
    echo "❌ JWT secret length is incorrect (expected 64, got ${#JWT_SECRET})"
fi

# Validate RSA keys
if echo "$RSA_PRIVATE_KEY" | grep -q "BEGIN RSA PRIVATE KEY"; then
    echo "✅ RSA private key format is correct"
else
    echo "❌ RSA private key format is incorrect"
fi

if echo "$RSA_PUBLIC_KEY" | grep -q "BEGIN PUBLIC KEY"; then
    echo "✅ RSA public key format is correct"
else
    echo "❌ RSA public key format is incorrect"
fi

# Test RSA key pair
echo "$RSA_PRIVATE_KEY" | sed 's/\\n/\n/g' > temp_private.pem
echo "$RSA_PUBLIC_KEY" | sed 's/\\n/\n/g' > temp_public.pem

if openssl rsa -in temp_private.pem -pubout 2>/dev/null | diff - temp_public.pem > /dev/null; then
    echo "✅ RSA key pair is valid"
else
    echo "❌ RSA key pair is invalid"
fi

rm -f temp_private.pem temp_public.pem

echo ""
echo "🔐 Key validation complete!"
EOF

chmod +x "$KEYS_DIR/validate-keys.sh"
echo -e "${GREEN}✅ Key validation script created${NC}"

# Set proper permissions
echo -e "\n${YELLOW}🔒 Setting secure file permissions...${NC}"
chmod 600 "$KEYS_DIR"/*.key
chmod 600 "$KEYS_DIR"/*.env
chmod 600 "$KEYS_DIR"/*.pem
chmod 600 "$KEYS_DIR/.env.generated"
echo -e "${GREEN}✅ Secure permissions set${NC}"

# Summary
echo -e "\n${GREEN}🎉 Security Keys Generation Complete!${NC}"
echo "=================================================="
echo ""
echo "📁 Generated files in $KEYS_DIR/:"
echo "   • .env.generated - Complete environment file"
echo "   • *.key - Individual key files"
echo "   • *.env - Service-specific environment files"
echo "   • *.pem - RSA key files"
echo "   • rotate-keys.sh - Key rotation script"
echo "   • validate-keys.sh - Key validation script"
echo ""
echo -e "${YELLOW}⚠️  IMPORTANT SECURITY NOTES:${NC}"
echo "1. 🔒 Keep these keys secure and never commit to version control"
echo "2. 🔄 Rotate keys regularly (every 90 days minimum)"
echo "3. 💾 Backup keys securely with proper encryption"
echo "4. 🔍 Monitor key usage and access"
echo "5. 🚀 Use different keys for different environments"
echo ""
echo -e "${BLUE}📋 Next steps:${NC}"
echo "1. Copy keys to your secure storage (password manager, vault)"
echo "2. Update your .env file with generated values"
echo "3. Deploy services with new keys"
echo "4. Run validation script: cd $KEYS_DIR && ./validate-keys.sh"
echo "5. Test all services to ensure they work with new keys"
echo ""
echo -e "${GREEN}🛡️ Your Go Coffee platform is now secured with enterprise-grade cryptography!${NC}"
