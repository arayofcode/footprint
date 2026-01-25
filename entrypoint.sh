#!/bin/sh
set -e

# Extract arguments for the footprint binary
# We pass all arguments received by the script to the binary
/footprint "$@"

# If OUTPUT_BRANCH is set, we handle the git push
if [ -n "$OUTPUT_BRANCH" ]; then
    echo "🚀 Storing artifacts to branch: $OUTPUT_BRANCH"
    
    # Configure git
    git config --global user.name "github-actions[bot]"
    git config --global user.email "github-actions[bot]@users.noreply.github.com"
    git config --global --add safe.directory /github/workspace
    
    # Remote URL with token for authentication
    # GITHUB_REPOSITORY is automatically provided by the Actions runner (e.g. "user/repo")
    REMOTE_REPO="https://x-access-token:${GITHUB_TOKEN}@github.com/${GITHUB_REPOSITORY}.git"
    
    # Determine the directory where artifacts were generated
    # Defaults to 'dist' if not set in args (which action.yaml handles)
    TARGET_DIR="${OUTPUT_DIR:-dist}"
    
    if [ ! -d "$TARGET_DIR" ]; then
        echo "Error: Output directory '$TARGET_DIR' not found."
        exit 1
    fi

    # Create a temporary directory for the push to avoid messing with the workspace
    PUSH_DIR=$(mktemp -d)
    cp -r "$TARGET_DIR/." "$PUSH_DIR/"
    
    cd "$PUSH_DIR"
    git init
    git checkout -b "$OUTPUT_BRANCH"
    git add .
    
    # Only commit and push if there are changes
    if git commit -m "Auto-generated footprint data"; then
        git push "$REMOTE_REPO" "$OUTPUT_BRANCH" --force
        echo "✓ Footprint artifacts successfully pushed to $OUTPUT_BRANCH"
    else
        echo "No changes to commit."
    fi
fi
