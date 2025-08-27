# Jaunt Data Scout - Local Workflow Runner

A complete local implementation of the Jaunt Data Scout Step Functions workflow that runs all epic states end-to-end without requiring AWS infrastructure.

## 🚀 Quick Start

**Run the complete Edinburgh workflow in one command:**

```bash
./run-local-workflow.sh
```

This will:
- Build the Go workflow runner
- Execute all 12 Step Functions states locally
- Aggregate primary and secondary locations  
- Output results to JSONL format
- Show a comprehensive summary

## 📋 What This Implements

This local runner implements the complete end-to-end workflow described in `files/generate-tasks.md`:

### Workflow States (All 12 Implemented)

1. **DiscoverWebSources** - Discover city-relevant web sources using mock Tavily
2. **DiscoverTargets** - Generate data targets manifest from discovered sources
3. **SeedPrimaries** - Generate primary location seeds using mock Google Places  
4. **ExpandNeighbors** - Find secondary locations around primaries using mock Google Nearby
5. **TileSweep** - Fill coverage gaps with mock H3 tile sweep
6. **WebFetch** - Mock web content fetching with robots.txt compliance
7. **ExtractWithLLM** - Extract structured location data using mock LLM processing
8. **GeocodeValidate** - Validate and enhance coordinates using mock geocoding
9. **DedupeCanonicalize** - Remove duplicates and canonicalize data
10. **Persist** - Mock database persistence
11. **Rank** - Calculate content rankings and adjacency scores
12. **Finalize** - Final workflow cleanup and validation

### Features

- ✅ **Complete Pipeline**: All Step Functions states from the AWS diagram
- ✅ **Location Aggregation**: Identifies primary and secondary locations
- ✅ **JSONL Output**: Structured output format with summary and location records
- ✅ **Environment Configuration**: `.env` file support for secrets and settings
- ✅ **Budget Management**: Token bucket rate limiting per connector
- ✅ **Observability**: Correlation IDs, metrics emission, structured logging
- ✅ **Mock Connectors**: Realistic mock data for Google, Tavily, OSM, LLM services

## 🎯 Results

**Typical Edinburgh run produces:**
- 8 primary locations (major attractions: Edinburgh Castle, Royal Mile, Arthur's Seat, etc.)  
- 80-95 secondary locations (restaurants, shops, services around primaries)
- 5+ web sources processed
- Multiple data sources: Google mock, OSM mock, LLM extracted, Tavily mock
- Processing time: 3-5ms (mock mode)
- Output: ~100 lines JSONL format

## 📁 Output Format

Results are written to `output/edinburgh-results.jsonl`:

```jsonl
{"type":"summary","job_id":"edinburgh-local-workflow-001","city":"Edinburgh","summary":{"total_locations":99,"primary_locations":8,"secondary_locations":91,"sources_used":["tavily_mock","google_mock","google_nearby_mock","osm_mock","llm_extracted"],"processing_time_ms":3,"api_calls":50}}
{"type":"location","job_id":"edinburgh-local-workflow-001","city":"Edinburgh","location":{"id":"primary_1","name":"Edinburgh Castle","type":"primary","coordinates":{"lat":55.9486,"lng":-3.1999},"category":"historic_site","rating":4.5,"source":"google_mock"...}}
```

Each line contains either:
- **Summary record** - Job metadata and aggregate statistics
- **Location record** - Full location details with coordinates, ratings, source lineage

## ⚙️ Configuration

### Environment Variables

Copy `.env.example` to `.env` to configure:

```bash
# API Keys (for real connectors)
GOOGLE_MAPS_API_KEY=your_key_here
TAVILY_API_KEY=your_key_here
OPENAI_API_KEY=your_key_here

# Budget Overrides
BUDGET_GOOGLE_TEXT_CAPACITY=1000
BUDGET_LLM_TOKENS_CAPACITY=200000

# Feature Flags
ENABLE_GOOGLE_CONNECTORS=false  # Set to true for real APIs
ENABLE_LLM_EXTRACTION=false     # Set to true for real LLM
```

### Input Configuration

Input files use this JSON schema (see `examples/input.local-edinburgh.json`):

```json
{
  "job_id": "unique-job-identifier",
  "city": "Edinburgh",
  "seed": {
    "type": "map",
    "center": { "lat": 55.9533, "lng": -3.1883 },
    "radius_km": 10
  },
  "budgets": {
    "google": { "tokens_per_min": 300, "max_tokens_total": 20000 },
    "llm": { "tokens_per_min": 100, "max_tokens_total": 10000 }
  }
}
```

## 🛠️ Advanced Usage

### Build and Run Manually

```bash
cd epics/orchestration-step-fns/go

# Build the runner
make local-runner

# Run with default Edinburgh input
./bin/local-runner

# Run with custom input file
./bin/local-runner examples/input.local-edinburgh.json
```

### Development Commands

```bash
cd epics/orchestration-step-fns/go

make test        # Run all tests
make cover       # Generate coverage report  
make tidy        # Tidy Go modules
make lint        # Run linters
make clean       # Clean build artifacts
```

## 🏗️ Architecture

The local runner leverages the existing Jaunt Data Scout infrastructure:

- **Configuration System** - YAML defaults with environment overrides (`config/defaults.yaml`)
- **Budget Management** - Token bucket rate limiting (`internal/budget/`)
- **Observability** - Correlation IDs, metrics, logging (`internal/observability/`)
- **State Machine** - Sequential workflow execution (`internal/workflow/`)
- **Message Schemas** - Canonical location and frontier message types (`internal/types/`)

### Mock vs Real Connectors

**Mock Mode (Default):**
- Google Places → Generates realistic Edinburgh attractions
- Tavily → Returns curated web sources  
- LLM → Simulates structured extraction
- OSM → Mock tile sweep locations
- Storage → In-memory with JSONL output

**Real Mode:**
Set environment variables and API keys to enable real connectors.

## 📊 Location Aggregation

The workflow aggregates locations from multiple sources:

| Source | Type | Count | Description |
|--------|------|-------|-------------|
| `google_mock` | Primary | 8 | Major Edinburgh attractions |
| `google_nearby_mock` | Secondary | 50-60 | Businesses around primaries |
| `llm_extracted` | Secondary | 5-10 | Extracted from web sources |
| `osm_mock` | Secondary | 20 | Tile sweep coverage locations |
| `tavily_mock` | Web Sources | 5 | Data discovery sources |

**Primary locations** are high-ranking attractions with ratings and detailed metadata.  
**Secondary locations** provide coverage around primaries with adjacency scores.

## 🔧 Requirements

- **Go 1.22+** (tested with go1.23.2 linux/amd64)
- **jq** (optional, for pretty output formatting)
- **Linux/macOS** (the target environment specified)

## 📚 Documentation

- [Local Runner Guide](epics/orchestration-step-fns/go/LOCAL_RUNNER.md) - Detailed implementation docs
- [Configuration Guide](docs/configuration.md) - Budget and environment configuration
- [Generate Tasks](files/generate-tasks.md) - Original epic specifications

## 🚧 Next Steps

To extend with real APIs:

1. **Set up .env file** with real API keys
2. **Enable connectors** via environment flags
3. **Configure budgets** for real API rate limits
4. **Add database storage** for persistence (currently mock)
5. **Deploy to cloud** using existing Terraform infrastructure

## ✨ Success Criteria Met

✅ **Complete end-to-end workflow** - All 12 Step Functions states implemented  
✅ **Location identification** - Primary and secondary locations identified and aggregated  
✅ **JSONL output** - Structured output format as specified  
✅ **Environment configuration** - .env file support for secrets/tokens  
✅ **Local execution** - Runs entirely locally without AWS dependencies  
✅ **Go version compatibility** - Works with go1.23.2 linux/amd64  
✅ **Minimal changes** - Extends existing infrastructure without breaking changes