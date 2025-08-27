# Local Workflow Runner

This is a local implementation of the complete Jaunt Data Scout workflow that runs all the Step Functions states locally without requiring AWS infrastructure.

## Overview

The local runner implements the complete workflow pipeline:

1. **DiscoverWebSources** - Discover city-relevant web sources using mock Tavily functionality
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

## Usage

### Build the runner:

```bash
make local-runner
```

### Run with Edinburgh example:

```bash
make run-edinburgh
```

### Run with custom input:

```bash
./bin/local-runner path/to/input.json
```

### Environment Configuration

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
# Edit .env with your API keys and settings
```

Key environment variables:
- `CONFIG_PATH` - Path to YAML configuration (default: `config/defaults.yaml`)
- `OUTPUT_PATH` - Output JSONL file path (default: `output/edinburgh-results.jsonl`)
- `BUDGET_*` - Budget overrides for connectors
- `ENABLE_*` - Feature flags to enable/disable real connectors

## Input Format

Input files should follow this JSON schema:

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
    "web": { "bytes_per_min": 1000000, "max_bytes_total": 200000000 },
    "llm": { "tokens_per_min": 100, "max_tokens_total": 10000 },
    "open_data": { "calls_per_min": 100, "max_calls_total": 10000 }
  },
  "config": {
    "enable_web_discovery": true,
    "enable_llm_extraction": true,
    "mock_mode": true
  }
}
```

## Output Format

The runner outputs results in JSONL format:

1. **Summary line** - Contains job metadata and aggregate statistics
2. **Location lines** - One line per discovered location with full details

Example output:
```jsonl
{"type":"summary","job_id":"edinburgh-local-workflow-001","city":"Edinburgh","summary":{"total_locations":180,"primary_locations":8,"secondary_locations":172}}
{"type":"location","job_id":"edinburgh-local-workflow-001","city":"Edinburgh","location":{"id":"primary_1","name":"Edinburgh Castle","type":"primary",...}}
```

## Architecture

The local runner uses the existing infrastructure:

- **Budget Management** - Token bucket rate limiting per connector
- **Observability** - Correlation IDs, metrics emission, structured logging  
- **Configuration** - YAML defaults with environment overrides
- **State Machine** - Sequential execution of workflow states
- **Data Types** - Canonical location and message schemas

## Mock vs Real Connectors

By default, the runner uses mock implementations:

- **Google Places** - Generates realistic Edinburgh attractions with ratings
- **Tavily** - Returns curated web sources for the city
- **LLM Extraction** - Simulates structured extraction from web content
- **Geocoding** - Enhances coordinate confidence scores
- **Storage** - In-memory processing with JSONL output

To enable real connectors, set environment flags and provide API keys:

```bash
ENABLE_GOOGLE_CONNECTORS=true
ENABLE_WEB_DISCOVERY=true  
ENABLE_LLM_EXTRACTION=true
GOOGLE_MAPS_API_KEY=your_key_here
TAVILY_API_KEY=your_key_here
```

## Performance

Typical Edinburgh run produces:
- ~8 primary locations 
- ~170 secondary locations
- ~5-15 web sources processed
- Execution time: 1-5 seconds (mock mode)
- Output file: ~200KB JSONL

## Testing

```bash
make test
```

Run coverage analysis:
```bash
make cover  
```

## Related Documentation

- [Configuration Guide](../../docs/configuration.md)
- [Step Functions Workflow](../README.md)  
- [Integration Testing](../INTEGRATION_TESTING.md)