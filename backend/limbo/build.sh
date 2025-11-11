#!/bin/sh
# Build script for Limbo cognitive architecture modules
# Compiles .m and .b files to .dis bytecode

set -e

echo "=== Building Limbo Cognitive Architecture ==="
echo ""

# Check if limbo compiler is available
if ! command -v limbo > /dev/null 2>&1; then
    echo "Warning: limbo compiler not found"
    echo "This script requires Inferno OS or Hosted Inferno"
    echo "See: http://www.vitanuova.com/inferno/"
    echo ""
    echo "Skipping Limbo compilation..."
    exit 0
fi

# Set module path
export MKFLAGS="-I/module"

echo "Step 1: Compiling module definitions (.m files)..."
limbo -I/module modules/atomspace.m
echo "  ✓ atomspace.m"

limbo -I/module modules/inference.m
echo "  ✓ inference.m"

limbo -I/module modules/agents.m
echo "  ✓ agents.m"

limbo -I/module modules/pipeline.m
echo "  ✓ pipeline.m"

limbo -I/module dis/disvm.m
echo "  ✓ disvm.m"

echo ""
echo "Step 2: Compiling module implementations (.b files)..."
limbo modules/atomspace.b
echo "  ✓ atomspace.b → atomspace.dis"

limbo modules/inference.b
echo "  ✓ inference.b → inference.dis"

limbo modules/agents.b
echo "  ✓ agents.b → agents.dis"

limbo modules/pipeline.b
echo "  ✓ pipeline.b → pipeline.dis"

limbo dis/disvm.b
echo "  ✓ disvm.b → disvm.dis"

echo ""
echo "Step 3: Compiling examples..."
limbo examples/cognitive_demo.b
echo "  ✓ cognitive_demo.b → cognitive_demo.dis"

echo ""
echo "=== Build Complete ==="
echo ""
echo "To run the demo:"
echo "  /dis/limbo/examples/cognitive_demo.dis"
echo ""
echo "Or using emu (Inferno emulator):"
echo "  emu /dis/limbo/examples/cognitive_demo.dis"
echo ""
