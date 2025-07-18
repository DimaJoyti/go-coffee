#!/bin/bash

# Fix imports script for Developer DAO Platform
echo "🔧 Fixing import paths in frontend applications..."

# Function to fix imports in a file
fix_imports() {
    local file="$1"
    local app_dir="$2"
    
    echo "Fixing imports in: $file"
    
    # Determine the relative path to shared based on the file location
    local depth=$(echo "$file" | sed "s|$app_dir/||" | tr -cd '/' | wc -c)
    local relative_path=""
    
    for ((i=0; i<depth; i++)); do
        relative_path="../$relative_path"
    done
    
    # Add one more level to get to the web directory, then to shared
    relative_path="../${relative_path}shared/src"
    
    # Replace the imports
    sed -i "s|@developer-dao/shared/|$relative_path/|g" "$file"
}

# Fix imports in dao-portal
echo "Fixing dao-portal imports..."
find dao-portal -name "*.tsx" -o -name "*.ts" | while read file; do
    if grep -q "@developer-dao/shared" "$file"; then
        fix_imports "$file" "dao-portal"
    fi
done

# Fix imports in governance-ui
echo "Fixing governance-ui imports..."
find governance-ui -name "*.tsx" -o -name "*.ts" | while read file; do
    if grep -q "@developer-dao/shared" "$file"; then
        fix_imports "$file" "governance-ui"
    fi
done

echo "✅ Import paths fixed!"

# Also fix the package.json files to remove the workspace dependency
echo "🔧 Updating package.json files..."

# Remove the shared dependency from dao-portal package.json
if [ -f "dao-portal/package.json" ]; then
    echo "Updating dao-portal package.json..."
    # Remove the @developer-dao/shared dependency line
    sed -i '/"@developer-dao\/shared":/d' dao-portal/package.json
fi

# Remove the shared dependency from governance-ui package.json
if [ -f "governance-ui/package.json" ]; then
    echo "Updating governance-ui package.json..."
    # Remove the @developer-dao/shared dependency line
    sed -i '/"@developer-dao\/shared":/d' governance-ui/package.json
fi

echo "✅ Package.json files updated!"
echo "🎉 All import fixes completed!"
