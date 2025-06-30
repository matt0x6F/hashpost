#!/bin/bash

# Setup IBE master keys for HashPost development
# This script helps developers set up their IBE keys

set -e

KEYS_DIR="./keys"
MASTER_KEY_PATH="$KEYS_DIR/master.key"

echo "🔐 HashPost IBE Key Setup"
echo "=========================="
echo ""

# Check if keys directory exists
if [ ! -d "$KEYS_DIR" ]; then
    echo "📁 Creating keys directory..."
    mkdir -p "$KEYS_DIR"
fi

# Check if master key already exists
if [ -f "$MASTER_KEY_PATH" ]; then
    echo "✅ Master key already exists at $MASTER_KEY_PATH"
    echo "📊 Key info:"
    ls -la "$MASTER_KEY_PATH"
    echo "🔍 Key fingerprint: $(sha256sum "$MASTER_KEY_PATH" | cut -d' ' -f1)"
    echo ""
    echo "❓ Do you want to:"
    echo "   1. Keep existing key (recommended)"
    echo "   2. Generate new key (will break existing data)"
    echo "   3. Use test key (development only)"
    echo ""
    read -p "Enter choice (1-3): " choice
    
    case $choice in
        1)
            echo "✅ Keeping existing key"
            ;;
        2)
            echo "🔄 Generating new key..."
            # Generate 32 random bytes and encode as hex (no newline)
            openssl rand -hex 32 | tr -d '\n' > "$MASTER_KEY_PATH"
            chmod 600 "$MASTER_KEY_PATH"
            echo "✅ New key generated"
            echo "⚠️  WARNING: This will break all existing identity mappings!"
            ;;
        3)
            echo "🔄 Creating test key..."
            # Create a deterministic test key (32 bytes as hex, no newline)
            printf 'test_master_secret_32_bytes_long_key_hex' | xxd -p -c 64 | tr -d '\n' > "$MASTER_KEY_PATH"
            chmod 600 "$MASTER_KEY_PATH"
            echo "✅ Test key created"
            echo "⚠️  WARNING: This will break all existing identity mappings!"
            ;;
        *)
            echo "❌ Invalid choice, keeping existing key"
            ;;
    esac
else
    echo "❌ No master key found"
    echo ""
    echo "❓ Choose key type:"
    echo "   1. Generate new production key (recommended)"
    echo "   2. Use test key (development only)"
    echo ""
    read -p "Enter choice (1-2): " choice
    
    case $choice in
        1)
            echo "🔄 Generating new production key..."
            # Generate 32 random bytes and encode as hex (no newline)
            openssl rand -hex 32 | tr -d '\n' > "$MASTER_KEY_PATH"
            chmod 600 "$MASTER_KEY_PATH"
            echo "✅ Production key generated"
            echo "🔍 Key fingerprint: $(sha256sum "$MASTER_KEY_PATH" | cut -d' ' -f1)"
            ;;
        2)
            echo "🔄 Creating test key..."
            # Create a deterministic test key (32 bytes as hex, no newline)
            printf 'test_master_secret_32_bytes_long_key_hex' | xxd -p -c 64 | tr -d '\n' > "$MASTER_KEY_PATH"
            chmod 600 "$MASTER_KEY_PATH"
            echo "✅ Test key created"
            echo "⚠️  WARNING: This will break all existing identity mappings!"
            ;;
        *)
            echo "❌ Invalid choice, generating production key"
            openssl rand -hex 32 | tr -d '\n' > "$MASTER_KEY_PATH"
            chmod 600 "$MASTER_KEY_PATH"
            echo "✅ Production key generated"
            ;;
    esac
fi

echo ""
echo "🎉 IBE key setup complete!"
echo ""
echo "📋 Next steps:"
echo "   1. Start the application: make dev"
echo "   2. The container will mount the key from ./keys/master.key"
echo "   3. All identity mappings will use this consistent key"
echo ""
echo "💡 Tips:"
echo "   - Keep your master key secure and backed up"
echo "   - Use different keys for different environments"
echo "   - Never commit keys to version control"
echo "   - For production, use a proper key management system" 