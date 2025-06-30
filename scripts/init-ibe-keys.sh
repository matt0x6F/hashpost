#!/bin/bash

# Initialize IBE master keys for HashPost
# This script ensures consistent IBE keys across container restarts

set -e

KEYS_DIR="/app/keys"
MASTER_KEY_PATH="$KEYS_DIR/master.key"
MASTER_KEY_BACKUP_PATH="$KEYS_DIR/master.key.backup"

echo "🔐 Checking IBE master keys..."

# Create keys directory if it doesn't exist
mkdir -p "$KEYS_DIR"

# Check if master key already exists
if [ -f "$MASTER_KEY_PATH" ]; then
    echo "✅ Master key found at $MASTER_KEY_PATH"
    echo "📊 Master key info:"
    ls -la "$MASTER_KEY_PATH"
    echo "🔍 Key fingerprint: $(sha256sum "$MASTER_KEY_PATH" | cut -d' ' -f1)"
    
    # Verify key is readable by the application
    if [ -r "$MASTER_KEY_PATH" ]; then
        echo "✅ Key is readable by application"
    else
        echo "⚠️  Warning: Key may not be readable by application"
        echo "   Consider running: chmod 600 $MASTER_KEY_PATH"
    fi
else
    echo "❌ No master key found at $MASTER_KEY_PATH"
    echo ""
    echo "📋 To create a master key, run one of these commands from the host:"
    echo ""
    echo "   # Option 1: Generate a new key"
    echo "   mkdir -p ./keys"
    echo "   openssl rand -out ./keys/master.key 32"
    echo "   chmod 600 ./keys/master.key"
    echo ""
    echo "   # Option 2: Copy from existing key"
    echo "   cp /path/to/existing/master.key ./keys/master.key"
    echo "   chmod 600 ./keys/master.key"
    echo ""
    echo "   # Option 3: Use test key (for development only)"
    echo "   mkdir -p ./keys"
    echo "   echo -n 'test_master_secret_32_bytes_long_key' > ./keys/master.key"
    echo "   chmod 600 ./keys/master.key"
    echo ""
    echo "⚠️  WARNING: Using a test key will break all existing identity mappings!"
    echo "   Only use test keys in development environments."
    echo ""
    
    # Check if we're in development mode and should create a test key
    if [ "$ENV" = "development" ] && [ "$IBE_USE_TEST_KEY" = "true" ]; then
        echo "🔄 Creating test key for development environment..."
        echo -n 'test_master_secret_32_bytes_long_key' > "$MASTER_KEY_PATH"
        chmod 600 "$MASTER_KEY_PATH"
        echo "✅ Test key created"
        echo "⚠️  WARNING: This will break all existing identity mappings!"
    else
        echo "❌ Cannot proceed without a master key"
        echo "   Please create a master key as shown above, then restart the container"
        exit 1
    fi
fi

# Check if we need to rotate keys (if rotation is enabled)
if [ "$IBE_KEY_ROTATION_ENABLED" = "true" ]; then
    echo "🔄 Key rotation is enabled"
    echo "⏰ Rotation interval: $IBE_KEY_ROTATION_INTERVAL"
    echo "⏳ Grace period: $IBE_KEY_ROTATION_GRACE_PERIOD"
    
    # TODO: Implement key rotation logic here
    # This would involve:
    # 1. Checking if rotation is due
    # 2. Creating a new master key
    # 3. Re-encrypting existing identity mappings
    # 4. Updating the key version
    echo "⚠️  Key rotation logic not yet implemented"
else
    echo "⏸️  Key rotation is disabled"
fi

echo "🎉 IBE key check complete!" 