#!/bin/bash

echo "🚀 Testing HashPost UI and Backend Connection"
echo "=============================================="

# Check if Docker Compose is running
echo "📋 Checking Docker Compose services..."
if ! docker-compose ps | grep -q "hashpost-app\|hashpost-ui"; then
    echo "❌ Docker Compose services are not running."
    echo "   Start them with: docker-compose --profile dev up --build"
    exit 1
fi

echo "✅ Docker Compose services are running"

# Test backend health
echo "🔍 Testing backend health..."
BACKEND_HEALTH=$(curl -s http://localhost:8888/health 2>/dev/null)
if [ $? -eq 0 ]; then
    echo "✅ Backend is healthy: $BACKEND_HEALTH"
else
    echo "❌ Backend health check failed"
    echo "   Backend might still be starting up..."
fi

# Test backend hello endpoint
echo "🔍 Testing backend hello endpoint..."
BACKEND_HELLO=$(curl -s http://localhost:8888/hello 2>/dev/null)
if [ $? -eq 0 ]; then
    echo "✅ Backend hello endpoint: $BACKEND_HELLO"
else
    echo "❌ Backend hello endpoint failed"
fi

# Test UI accessibility
echo "🔍 Testing UI accessibility..."
UI_RESPONSE=$(curl -s -I http://localhost:3000 2>/dev/null | head -1)
if [ $? -eq 0 ]; then
    echo "✅ UI is accessible: $UI_RESPONSE"
else
    echo "❌ UI is not accessible"
fi

echo ""
echo "🌐 URLs:"
echo "   Backend API: http://localhost:8888"
echo "   UI Frontend: http://localhost:3000"
echo "   API Docs: http://localhost:8888/docs"
echo ""
echo "🔧 If you're still getting NetworkError:"
echo "   1. Make sure both services are fully started"
echo "   2. Check browser console for detailed error messages"
echo "   3. Try refreshing the UI page"
echo "   4. Check Docker logs: docker-compose logs ui app" 