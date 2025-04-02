#!/bin/bash

# Make the script executable from any directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Configure git to use our custom hooks directory
git config core.hooksPath "$PROJECT_ROOT/.githooks"

# Make the hook scripts executable
chmod +x "$PROJECT_ROOT/.githooks/pre-commit"
chmod +x "$PROJECT_ROOT/.githooks/pre-push"
chmod +x "$PROJECT_ROOT/.githooks/commit-msg"

echo "Git hooks configured successfully!"
echo "Tests will now run automatically before each commit and push."
echo "Commit message conventions will be enforced."
echo "Direct commits to master branch will be prevented."
