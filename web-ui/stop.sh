#!/bin/bash

# Go Coffee Epic UI - Stop Script

echo "🛑 Stopping Go Coffee Epic UI..."

# Stop and remove containers
docker-compose -f docker-compose.ui.yml down

echo "✅ Services stopped successfully!"
echo ""
echo "💡 To start again: ./start.sh"
echo "🧹 To clean up volumes: docker-compose -f docker-compose.ui.yml down -v"
