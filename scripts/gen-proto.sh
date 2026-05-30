#!/bin/bash

# gen-proto.sh - Generates code from .proto files for all languages
# Generates Java, Go, Rust, Python, and TypeScript code

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_DIR="$REPO_ROOT/proto"
GEN_DIR="$PROTO_DIR/generated"

echo "Generating code from .proto files..."
echo "Proto directory: $PROTO_DIR"
echo "Output directory: $GEN_DIR"
echo ""

# Create generated directory
mkdir -p "$GEN_DIR"

# Use Buf to generate code for all languages
echo "Using Buf to generate code..."
buf generate "$PROTO_DIR"

echo ""
echo "✓ Code generation complete!"
echo ""
echo "Generated artifacts:"
echo "  - Go: proto/generated/go"
echo "  - Java: services/*/src/main/java/com/rtscs/proto/gen (to be copied)"
echo "  - Python: services/analytics-service/gen"
echo "  - TypeScript: services/api-gateway/src/gen"
echo "  - Rust: services/logistics-service/src/gen"
echo ""
echo "Next steps:"
echo "1. Copy generated Java files to respective services"
echo "2. Update service implementations with generated code"
echo "3. Run tests: make test"
echo ""
