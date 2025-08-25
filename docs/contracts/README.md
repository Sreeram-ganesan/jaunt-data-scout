# Contract Artifacts

This directory contains JSON Schema definitions and examples for the Jaunt Data Scout pipeline contracts.

## Overview

The Jaunt Data Scout system processes place data through multiple stages, from initial discovery through final canonicalization. These schemas define the contracts between different components of the system to ensure data consistency and enable reliable processing.

All schemas use **JSON Schema Draft-07** for broad compatibility and tooling support.

## Schemas

### Frontier Messages

Frontier messages are used to queue work items for different types of data collection:

- **`frontier.web.json`** - Web frontier message schema for web scraping and extraction tasks
  - Required: `type="web"`, `city`, `source_url`, `source_name`, `source_type`, `crawl_depth`, `correlation_id`
  - Optional: `trust_score`, `coordinates_confidence`, `priority`, `timeout_seconds`, `metadata`

- **`frontier.maps.json`** - Maps frontier message schema for geographic search tasks
  - Required: `type="maps"`, `city`, `lat`, `lng`, `radius`, `correlation_id`
  - Optional: `category`, `trust_score`, `coordinates_confidence`, `search_type`, `place_types`, `metadata`

### Entity Schemas

- **`canonical.candidate.json`** - Normalized place entity schema for the final canonical representation
  - Comprehensive schema including location, categorization, lineage tracking, external references
  - Supports merge history for reversible deduplication
  - Includes source-specific additional content (Google, web, signals, scoring)

- **`extraction.web.json`** - LLM extraction output schema for web-based entity extraction
  - Strict JSON schema for LLM outputs
  - Includes extraction metadata, model information, and confidence scores
  - Array of extracted entities with detailed attribution

## Schema Validation

### Local Validation

To validate examples locally, use `ajv-cli`:

```bash
# Install ajv-cli globally
npm install -g ajv-cli

# Validate a specific example against its schema
ajv validate -s schemas/frontier.web.json -d examples/frontier/web.example.json

# Validate all examples (from project root)
npm run validate-schemas
```

### CI Validation

Schema validation runs automatically on push and pull requests via GitHub Actions. The CI workflow validates all example files against their corresponding schemas using `ajv-cli`.

## Usage Guidelines

### For Producers (Data Collectors)

1. **Frontier Messages**: When enqueueing work items, ensure your message conforms to the appropriate frontier schema
2. **Web Extraction**: LLM extractors must output valid JSON according to `extraction.web.json`
3. **Correlation IDs**: Always include correlation IDs for tracing requests across the pipeline
4. **Confidence Scores**: Provide confidence scores where available to aid in downstream processing

### For Consumers (Data Processors)

1. **Validation**: Always validate incoming messages against the appropriate schema
2. **Error Handling**: Gracefully handle schema validation failures with appropriate logging
3. **Optional Fields**: Check for presence of optional fields before accessing
4. **Lineage**: Preserve and extend lineage information when transforming data

## Schema Evolution

When updating schemas:

1. Follow semantic versioning principles
2. Maintain backward compatibility when possible
3. Update corresponding example files
4. Test schema changes against existing data
5. Document breaking changes in release notes

## Field Definitions

### Common Fields

- **`correlation_id`**: UUID v4 for tracking requests across pipeline stages
- **`trust_score`**: 0.0-1.0 confidence in source reliability
- **`coordinates_confidence`**: 0.0-1.0 confidence in location accuracy
- **`confidences`**: Object containing confidence scores for different aspects

### Source Types

- **`google`**: Google Places API data
- **`osm`**: OpenStreetMap data  
- **`otm`**: OpenTripMap data
- **`wikidata`**: Wikidata/Wikipedia data
- **`web`**: Web scraping/LLM extraction
- **`open_data`**: Government/municipal open data
- **`tavily`**: Tavily API results

## Contributing

When contributing to these schemas:

1. Ensure examples are realistic and comprehensive
2. Include all required fields in examples
3. Test schema changes against the CI validation
4. Update documentation for new fields or breaking changes
5. Consider the impact on existing producers and consumers

## Support

For questions about these schemas or their usage, please refer to the project documentation or create an issue in the repository.