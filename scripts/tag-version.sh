#!/bin/bash
set -e

# Script to tag a new version for check-projects
# Usage: ./scripts/tag-version.sh [patch|minor|major]

# Get current version from latest git tag
CURRENT_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
CURRENT_VERSION="${CURRENT_TAG#v}"

# Calculate next versions
IFS='.' read -r -a VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR="${VERSION_PARTS[0]}"
MINOR="${VERSION_PARTS[1]}"
PATCH="${VERSION_PARTS[2]}"

NEXT_PATCH="$MAJOR.$MINOR.$((PATCH + 1))"
NEXT_MINOR="$MAJOR.$((MINOR + 1)).0"
NEXT_MAJOR="$((MAJOR + 1)).0.0"

# If version type provided as argument, use it
if [ -n "$1" ]; then
  VERSION_TYPE="$1"
else
  # Interactive prompt
  echo "Select version bump type:"
  echo "  1) patch  $NEXT_PATCH"
  echo "  2) minor  $NEXT_MINOR"
  echo "  3) major  $NEXT_MAJOR"
  echo ""
  read -p "Enter choice [1-3] (default: 1): " choice

  case "$choice" in
    2)
      VERSION_TYPE="minor"
      ;;
    3)
      VERSION_TYPE="major"
      ;;
    1|"")
      VERSION_TYPE="patch"
      ;;
    *)
      echo "❌ Invalid choice. Defaulting to patch."
      VERSION_TYPE="patch"
      ;;
  esac

  echo ""
fi

if [[ ! "$VERSION_TYPE" =~ ^(patch|minor|major)$ ]]; then
  echo "❌ Invalid version type. Use: patch, minor, or major"
  exit 1
fi

# Calculate new version
case "$VERSION_TYPE" in
  patch) NEW_VERSION="$NEXT_PATCH" ;;
  minor) NEW_VERSION="$NEXT_MINOR" ;;
  major) NEW_VERSION="$NEXT_MAJOR" ;;
esac

echo "========================================="
echo "🔄 Bumping version: $VERSION_TYPE"
echo "========================================="
echo ""
echo "Current version: $CURRENT_VERSION"
echo "New version: $NEW_VERSION"
echo ""

# Create annotated tag
git tag -a "v$NEW_VERSION" -m "v$NEW_VERSION"

echo "========================================="
echo "✅ Tagged: v$NEW_VERSION"
echo "========================================="
echo ""
echo "Latest commit: $(git log -1 --oneline)"
echo "Tag: v$NEW_VERSION"
echo ""

# Ask for push confirmation (Y by default)
read -p "Push to remote? (Y/n): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Nn]$ ]]; then
  echo "🚀 Pushing to remote..."
  git push --follow-tags
  echo ""
  echo "========================================="
  echo "✅ Pushed to remote!"
  echo "========================================="
  echo ""
  echo "GitHub Actions will now build and release."
  echo "Monitor: https://github.com/uralys/check-projects/actions"
else
  echo ""
  echo "⏭️  Skipped push. Push manually with:"
  echo "  git push --follow-tags"
fi
echo ""
