#!/bin/bash

# Epic Crypto Terminal - Setup Script
echo "🚀 Setting up Epic Crypto Terminal..."

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+ first."
    exit 1
fi

# Check Node.js version
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo "❌ Node.js version 18+ is required. Current version: $(node -v)"
    exit 1
fi

echo "✅ Node.js version: $(node -v)"

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "❌ npm is not installed."
    exit 1
fi

echo "✅ npm version: $(npm -v)"

# Install dependencies
echo "📦 Installing dependencies..."
npm install

if [ $? -ne 0 ]; then
    echo "❌ Failed to install dependencies"
    exit 1
fi

echo "✅ Dependencies installed successfully"

# Create environment file if it doesn't exist
if [ ! -f ".env.local" ]; then
    echo "🔧 Creating environment file..."
    cp .env.example .env.local
    echo "✅ Environment file created (.env.local)"
    echo "📝 Please edit .env.local with your configuration"
else
    echo "✅ Environment file already exists"
fi

# Run type check
echo "🔍 Running TypeScript type check..."
npm run type-check

if [ $? -ne 0 ]; then
    echo "❌ TypeScript type check failed"
    exit 1
fi

echo "✅ TypeScript type check passed"

# Run linting
echo "🧹 Running ESLint..."
npm run lint

if [ $? -ne 0 ]; then
    echo "⚠️  Linting issues found (non-critical)"
else
    echo "✅ Linting passed"
fi

# Test build
echo "🏗️  Testing production build..."
npm run build

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

# Cleanup build files
rm -rf .next

echo ""
echo "🎉 Epic Crypto Terminal setup complete!"
echo ""
echo "🚀 To start development:"
echo "   npm run dev"
echo ""
echo "🌐 Then open: http://localhost:3001"
echo ""
echo "📚 For troubleshooting, see: TROUBLESHOOTING.md"
echo ""
echo "🐳 For Docker deployment:"
echo "   cd .. && docker-compose -f docker-compose.epic.yml up -d"
echo ""
