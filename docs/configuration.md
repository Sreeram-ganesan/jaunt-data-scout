# Configuration: Budgets, Concurrency, Early-Stop (Defaults and Overrides)

This document describes how defaults are defined and how to override them per environment or run.

## Files
- `config/defaults.yaml` — canonical defaults for:
  - Token-bucket budgets per connector (google.text, google.nearby, google.details, overpass, otm, wiki, tavily.api, web.fetch, llm.tokens, nominatim)
  - Split ratio (primaries vs secondaries)
  - Early-stop thresholds (min_new_unique_rate, window, wall-clock, max_api_calls)
  - Advisory concurrency hints

These keys align with:
- Go budget guard types in `epics/orchestration-step-fns/go/internal/budget/budget_guard.go`
- Step Functions budget/early-stop gates in `epics/orchestration-step-fns/terraform/sfn/definition.asl.json`

## Override Precedence
1. Environment variables (highest)
2. Provided run-level config file (if any)
3. `config/defaults.yaml` (fallback)

## Environment Variable Mapping

For per-connector token buckets:
- Name format: `BUDGET_<TOKEN>_<FIELD>`
- Token mapping: replace dots with underscores and uppercase
  - `google.text` → `GOOGLE_TEXT`
  - `tavily.api` → `TAVILY_API`
  - `llm.tokens` → `LLM_TOKENS`

Fields:
- `CAPACITY` (int)
- `REFILL` (int)
- `PERIOD` (Go duration string, e.g., `1m`, `30s`, `2h`)

Examples:
- `BUDGET_GOOGLE_TEXT_CAPACITY=1200`
- `BUDGET_TAVILY_API_REFILL=120`
- `BUDGET_LLM_TOKENS_PERIOD=0s`

Global settings:
- `BUDGET_SPLIT_RATIO=0.7`
- `EARLY_STOP_MIN_NEW_UNIQUE_RATE=0.05`
- `EARLY_STOP_WINDOW=200`
- `BUDGET_MAX_API_CALLS=5000`
- `BUDGET_MAX_WALL_CLOCK_HOURS=6`

Concurrency (advisory):
- `CONCURRENCY_WEB_FETCH=8`
- `CONCURRENCY_EXTRACT_LLM=4`
- `CONCURRENCY_GEOCODE_VALIDATE=4`
- `CONCURRENCY_MAPS_EXPAND_NEIGHBORS=6`

## Notes
- Durations use Go-style syntax (`time.ParseDuration`).
- Keep taxonomy in sync across code, config, and Step Functions gates.