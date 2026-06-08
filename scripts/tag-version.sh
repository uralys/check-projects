#!/bin/bash
set -e

# Tag a new release for check-projects.
# Flow: generate changelog -> test -> build -> commit changelog -> tag.
# Pushing is left to you (printed at the end).
# Usage: ./scripts/tag-version.sh [patch|minor|major]

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

# --- Resolve current and next versions -------------------------------------

CURRENT_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
CURRENT_VERSION="${CURRENT_TAG#v}"

IFS='.' read -r -a VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR="${VERSION_PARTS[0]}"
MINOR="${VERSION_PARTS[1]}"
PATCH="${VERSION_PARTS[2]}"

NEXT_PATCH="$MAJOR.$MINOR.$((PATCH + 1))"
NEXT_MINOR="$MAJOR.$((MINOR + 1)).0"
NEXT_MAJOR="$((MAJOR + 1)).0.0"

if [ -n "$1" ]; then
  VERSION_TYPE="$1"
else
  echo "Select version bump type:"
  echo "  1) patch  $NEXT_PATCH"
  echo "  2) minor  $NEXT_MINOR"
  echo "  3) major  $NEXT_MAJOR"
  echo ""
  read -p "Enter choice [1-3] (default: 1): " choice

  case "$choice" in
    2) VERSION_TYPE="minor" ;;
    3) VERSION_TYPE="major" ;;
    1|"") VERSION_TYPE="patch" ;;
    *) echo "❌ Invalid choice. Defaulting to patch."; VERSION_TYPE="patch" ;;
  esac
  echo ""
fi

if [[ ! "$VERSION_TYPE" =~ ^(patch|minor|major)$ ]]; then
  echo "❌ Invalid version type. Use: patch, minor, or major"
  exit 1
fi

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
echo "New version:     $NEW_VERSION"
echo ""

# --- 1. Generate changelog --------------------------------------------------

CHANGELOG_FILE="changelogs/${NEW_VERSION}.md"
INDEX_FILE="changelogs/index.md"
TODAY=$(date +%F)

if [ -f "$CHANGELOG_FILE" ]; then
  echo "📄 Using existing changelog: $CHANGELOG_FILE"
else
  echo "📝 Generating changelog: $CHANGELOG_FILE"

  FEATS=$(git log --no-merges --pretty=format:'%s' "${CURRENT_TAG}..HEAD" | grep -E '^feat(\(.+\))?!?:' | sed -E 's/^feat(\(.+\))?!?: */- /' || true)
  FIXES=$(git log --no-merges --pretty=format:'%s' "${CURRENT_TAG}..HEAD" | grep -E '^fix(\(.+\))?!?:' | sed -E 's/^fix(\(.+\))?!?: */- /' || true)
  OTHERS=$(git log --no-merges --pretty=format:'%s' "${CURRENT_TAG}..HEAD" | grep -vE '^(feat|fix|docs|test|chore)(\(.+\))?!?:' | sed -E 's/^/- /' || true)

  {
    echo "# v${NEW_VERSION} - ${TODAY}"
    echo ""
    if [ -n "$FEATS" ]; then
      echo "## Features"
      echo ""
      echo "$FEATS"
      echo ""
    fi
    if [ -n "$FIXES" ]; then
      echo "## Bug Fixes"
      echo ""
      echo "$FIXES"
      echo ""
    fi
    if [ -n "$OTHERS" ]; then
      echo "## Changes"
      echo ""
      echo "$OTHERS"
      echo ""
    fi
  } > "$CHANGELOG_FILE"

  echo "   Draft written from commits since ${CURRENT_TAG}. Review before pushing."
fi

# Insert index entry if not already present.
if ! grep -q "(\./${NEW_VERSION}.md)" "$INDEX_FILE"; then
  TITLE=$(grep -m1 '^### ' "$CHANGELOG_FILE" | sed -E 's/^### *//')
  [ -z "$TITLE" ] && TITLE="$VERSION_TYPE release"
  ENTRY="- [v${NEW_VERSION}](./${NEW_VERSION}.md) - ${TODAY} - ${TITLE}"
  awk -v entry="$ENTRY" '
    !done && /^## Versions/ { print; getline; print; print entry; done=1; next }
    { print }
  ' "$INDEX_FILE" > "$INDEX_FILE.tmp" && mv "$INDEX_FILE.tmp" "$INDEX_FILE"
  echo "   Index updated: $INDEX_FILE"
fi
echo ""

# --- 2. Test ----------------------------------------------------------------

echo "🧪 Running tests..."
go test ./...
echo "✅ Tests passed!"
echo ""

# --- 3. Build ---------------------------------------------------------------

echo "🔨 Building..."
make build
echo "✅ Build succeeded!"
echo ""

# --- 4. Commit changelog ----------------------------------------------------

echo "📝 Committing changelog..."
git add "$CHANGELOG_FILE" "$INDEX_FILE"
git commit -m "chore: release v${NEW_VERSION}"

# --- 5. Tag -----------------------------------------------------------------

git tag -a "v${NEW_VERSION}" -m "v${NEW_VERSION}"

echo ""
echo "========================================="
echo "✅ Committed and tagged: v${NEW_VERSION}"
echo "========================================="
echo ""
echo "Commit: $(git log -1 --oneline)"
echo "Tag:    v${NEW_VERSION}"
echo ""
echo "When ready, push:"
echo "  git push --follow-tags"
echo ""
echo "Pushing the tag triggers the GitHub release."
echo ""
