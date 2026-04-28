#!/usr/bin/env bash
set -euo pipefail

# Convenience entrypoint for people who expect a simple skills installer.
# Delegates to the maintained TestGen agent integration installer.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$SCRIPT_DIR/scripts/install-agent-skill.sh" "$@"
