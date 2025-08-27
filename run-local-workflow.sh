#!/bin/bash

# Jaunt Data Scout - Local Workflow Runner
# This script runs the complete end-to-end workflow locally

set -e  # Exit on any error

echo "🚀 Jaunt Data Scout - Local Workflow Runner"
echo "==========================================="

# Change to the Go module directory
cd "$(dirname "$0")/epics/orchestration-step-fns/go"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.22+ to continue."
    exit 1
fi

GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | head -1)
echo "✅ Found Go version: $GO_VERSION"

# Build the local runner
echo "🔨 Building local workflow runner..."
make local-runner

# Check if input file was provided
INPUT_FILE=${1:-"examples/input.local-edinburgh.json"}
if [ ! -f "$INPUT_FILE" ]; then
    echo "❌ Input file not found: $INPUT_FILE"
    echo "Available input files:"
    ls -la examples/*.json 2>/dev/null || echo "No example files found"
    exit 1
fi

echo "📋 Using input file: $INPUT_FILE"

# Show input summary
echo "📊 Input Summary:"
if command -v jq &> /dev/null; then
    echo "   City: $(jq -r '.city' "$INPUT_FILE")"
    echo "   Job ID: $(jq -r '.job_id' "$INPUT_FILE")"
    echo "   Center: $(jq -r '.seed.center.lat' "$INPUT_FILE"), $(jq -r '.seed.center.lng' "$INPUT_FILE")"
    echo "   Radius: $(jq -r '.seed.radius_km' "$INPUT_FILE")km"
else
    echo "   (Install 'jq' for detailed input summary)"
fi

# Run the workflow
echo ""
echo "🎯 Running workflow..."
echo "==================="
./bin/local-runner "$INPUT_FILE"

# Show output summary
OUTPUT_FILE="output/edinburgh-results.jsonl"
if [ -f "$OUTPUT_FILE" ]; then
    echo ""
    echo "📈 Results Summary:"
    echo "=================="
    
    if command -v jq &> /dev/null; then
        SUMMARY=$(head -n 1 "$OUTPUT_FILE")
        echo "   Total locations: $(echo "$SUMMARY" | jq -r '.summary.total_locations')"
        echo "   Primary locations: $(echo "$SUMMARY" | jq -r '.summary.primary_locations')"  
        echo "   Secondary locations: $(echo "$SUMMARY" | jq -r '.summary.secondary_locations')"
        echo "   Processing time: $(echo "$SUMMARY" | jq -r '.summary.processing_time_ms')ms"
        echo "   API calls: $(echo "$SUMMARY" | jq -r '.summary.api_calls')"
        echo "   Sources used: $(echo "$SUMMARY" | jq -r '.summary.sources_used | join(", ")')"
        echo ""
        echo "📍 Sample Primary Locations:"
        grep '"type":"primary"' "$OUTPUT_FILE" | head -3 | jq -r '.location.name' | sed 's/^/   • /'
        echo ""
        echo "🎯 Output file: $OUTPUT_FILE ($(wc -l < "$OUTPUT_FILE") lines)"
    else
        echo "   Output file: $OUTPUT_FILE"
        echo "   Lines: $(wc -l < "$OUTPUT_FILE")"
    fi
    
    echo ""
    echo "✅ Workflow completed successfully!"
else
    echo "❌ Output file not found: $OUTPUT_FILE"
    exit 1
fi

echo ""
echo "💡 Next steps:"
echo "   • View full results: cat $OUTPUT_FILE | jq ."
echo "   • Run with different city: $0 examples/input.london.json"  
echo "   • Configure .env file for real APIs (copy from .env.example)"
echo "   • See documentation: epics/orchestration-step-fns/go/LOCAL_RUNNER.md"