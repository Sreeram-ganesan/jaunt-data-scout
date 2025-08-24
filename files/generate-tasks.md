# Jaunt Data Scout – MVP Epics and Tasks

This document enumerates the MVP Epics and their Tasks derived from data-scout.md (PRD). Use it to drive issue creation and tracking.

How to use
1) Labels: create `MVP`, `backend`, `data`, `API`, `infra`, `LLM`, `Postgres`, `observability`, `ETL`, `QA`.
2) Milestone: create `MVP (Edinburgh)`.
3) Create one issue per Epic below. Copy each Epic section’s “Goals”, “Acceptance criteria”, and “Tasks” into the issue body. Apply relevant labels and the milestone.
4) Keep “Tasks” as a task list in the Epic issue. Convert each task list item to its own issue from the GitHub UI (“…” menu → Convert to issue). This preserves parent/child links and progress.
5) Use a GitHub Project to track work. Suggested views: by Epic, by label, by status; group by Milestone.
6) Reference issues in PRs with `closes #<n>` so they close on merge.

---

## Epic: Orchestration — Step Functions workflow for city discovery
Labels: MVP, infra, backend

Goals
- Implement AWS Step Functions orchestration for the MVP pipeline: DiscoverTargets → SeedPrimaries → ExpandNeighbors → TileSweep → DedupeCanonicalize → Persist → Finalize.
- Integrate SQS frontier + DLQ, S3 raw cache, and metrics/traces.

Acceptance criteria
- A city job completes end-to-end for Edinburgh with mocked connectors.
- State transitions logged; failures routed to DLQ; raw responses cached to S3 by (source, request_hash).
- Configurable budgets respected (max_api_calls, max_wall_clock_hours, min_new_unique_rate).
- CloudWatch metrics and OTEL traces emitted for each state.

Tasks
- [ ] Define Step Functions state machine with all MVP states and retry/backoff.
- [ ] Implement budget guard + early-stop choice states.
- [ ] Wire SQS queues (frontier, DLQ) and message schema (lat/lng, radius, category).
- [ ] S3 layout for raw cache and manifests (per city run).
- [ ] Emit metrics/alerts per state; add tracing context propagation.

## Epic: Connectors — Overpass, OpenTripMap, Wikidata/Wikipedia
Labels: MVP, data, backend

Goals
- Build robust connectors with token buckets, polite rate limits, and S3 raw response caching.

Acceptance criteria
- Overpass: parameterized query builder by city polygon + categories; respects etiquette.
- OpenTripMap: category/box/radius fetch; popularity proxies collected.
- Wikidata/Wikipedia: SPARQL + page extracts; authority signals.
- All connectors return normalized candidates with source lineage.

Tasks
- [ ] Overpass client with city polygon envelope and adaptive radius.
- [ ] OpenTripMap client with pagination, keys, and throttling.
- [ ] Wikidata SPARQL queries + Wikipedia summaries.
- [ ] Raw response cache keying; idempotency and retry/backoff.
- [ ] Unit tests with recorded fixtures.

## Epic: Tavily + LLM City Data Targets Manifest (v1)
Labels: MVP, LLM, data

Goals
- Generate a machine-readable manifest of open-data portals, datasets, and endpoints per city using Tavily + Gemini T1–T4 prompts.

Acceptance criteria
- manifest.json for Edinburgh written to S3 with queries, candidates, datasets, endpoints, citations.
- Compliant crawling (robots/ToS); license captured when present.
- Query pack versioned (query_pack_version=v1).

Tasks
- [ ] Author T1–T4 prompt set; integrate with Tavily search/extract/crawl.
- [ ] Candidate scoring and shortlisting logic.
- [ ] Manifest JSON schema validation.
- [ ] Store Tavily queries and citations in lineage.

## Epic: Discovery — Seed primaries, expand neighbors, tile sweep
Labels: MVP, data, backend

Goals
- Achieve target coverage with primary/secondary split using chain-reaction expansion and H3 tile sweep.

Acceptance criteria
- 150–200 primaries selected; secondaries ≥95% by count.
- Average ≥15 secondaries within 800 m of each primary.
- H3 res 9/10 sweep fills under-dense cells.

Tasks
- [ ] Primary seeding from OSM/OTM/Wiki + open-data.
- [ ] Neighbor expansion around primaries (adaptive 300–800 m); create NEAR edges.
- [ ] Under-dense H3 cell sweep; configurable res.
- [ ] Budget-aware frontier enqueuing.

## Epic: Extraction & Normalization — Structured JSON via T5 and parsers
Labels: MVP, LLM, ETL

Goals
- Prefer machine-readable ingestion; fall back to T5 extractor (Gemini) for HTML.

Acceptance criteria
- Country-aware address normalization with valid city/state/country mapping.
- Strict JSON outputs; schema validation; lineage captured.
- Nominatim fallback for missing coordinates only.

Tasks
- [ ] CKAN/Socrata/ArcGIS/GeoJSON readers.
- [ ] Gemini T5 extractor with bounded JSON schema and timeouts.
- [ ] Address normalizer; Nominatim client with conservative usage.
- [ ] Field mappers to canonical candidate schema.

## Epic: Validation & Dedupe — Canonicalization pipeline
Labels: MVP, ETL, backend

Goals
- Enforce quality bars: ≤3% dupes; ≥98% coordinates within 50 m; ≥95% valid addresses.

Acceptance criteria
- Dedupe rules: polygon containment, ≤60 m proximity, token-set name sim ≥0.90, address checks, Wikidata aliases.
- LLM tie-break only for conflicts; reversible merges with lineage.

Tasks
- [ ] Similarity functions and thresholds; configurable.
- [ ] Merge strategy and conflict resolution with Gemini tie-break.
- [ ] QA harness for random sample checks and dupe audit.

## Epic: Ranking — Primary composite and secondary adjacency scoring
Labels: MVP, data, backend

Goals
- Implement balanced primary score and secondary adjacency scoring; write to content_rank and details JSON.

Acceptance criteria
- Tunable weights: popularity, authority, geo, novelty, graph.
- Secondary adjacency_score computed and stored with anchors.

Tasks
- [ ] Score computation and breakdown JSON.
- [ ] Graph centrality based on NEAR edges.
- [ ] Parameterized weighting; city defaults in config.

## Epic: Storage — Postgres writers and S3 raw cache/manifests
Labels: MVP, Postgres, backend

Goals
- Persist canonical places to existing schema; store raw payloads and manifests to S3.

Acceptance criteria
- Tables populated: t_locations_prd, t_location_coordinates, t_locations_prd_details, t_location_neighborhoods.
- additional_content JSONB fields populated per spec (external_refs, signals, score_breakdown, lineage, crawler, secondary_signals).
- Recommended indexes created.

Tasks
- [ ] DB mappers + upsert logic; bbox/h3 precompute.
- [ ] External refs consolidation and preferred externalid selection.
- [ ] SQL migrations to add indexes if missing.

## Epic: Public APIs — Jobs, list by city, place detail, nearby secondaries
Labels: MVP, API, backend

Goals
- Expose REST endpoints with p95 ≤ 300 ms on 10k-place cities.

Acceptance criteria
- Endpoints implemented per PRD (list by city, place detail, nearby secondaries, jobs).
- Cursor pagination; filters; category param; radius for nearby.
- Input validation and error semantics documented.

Tasks
- [ ] API handlers + schema types.
- [ ] DB access paths and indexes usage.
- [ ] Load/perf tests; caching strategy where needed.

## Epic: Observability — Metrics, traces, and alerts
Labels: MVP, observability, infra

Goals
- Provide visibility into yield, dupes, coverage, cost estimate; alerts for anomalies.

Acceptance criteria
- CloudWatch dashboards and alarms per state.
- OTEL traces stitched across workers.
- Cost estimation metrics and budget breach alerts.

Tasks
- [ ] Metrics schema and emission.
- [ ] Tracing propagation; SQS + Step Functions context.
- [ ] Alarm definitions and runbooks.

## Epic: QA & Acceptance — Edinburgh MVP run
Labels: MVP, QA

Goals
- Execute full Edinburgh city run and meet success criteria.

Acceptance criteria
- ≥ 4,000 places passing validation; primary set 150–200 with ≥90% having non-null popularity/authority.
- ≤3% dupes; ≥98% coord accuracy; ≥95% valid addresses (spot checks n=100).
- City Data Targets Manifest v1 saved to S3.

Tasks
- [ ] Tile sweep tuning and parameter review.
- [ ] Runbook for city job; incident handling.
- [ ] Final acceptance checks and documentation.

---

## Standalone Tasks
Use these as child tasks under the relevant Epics, or as individual issues.

- [ ] Config & Budgets defaults (YAML) and runtime overrides — Labels: MVP, infra, backend
- [ ] H3 utilities and under-dense cell detection (res 9/10/12) — Labels: MVP, backend, data
- [ ] Address normalization and geo validation (country-aware) — Labels: MVP, ETL, data
- [ ] Dedupe similarity functions and thresholds (name/address/geo) — Labels: MVP, ETL, backend
- [ ] Postgres upserts and index creation — Labels: MVP, Postgres, backend
- [ ] API performance targets and load testing (p95 ≤ 300 ms) — Labels: MVP, API, observability
- [ ] Edinburgh-specific open-data sources triage — Labels: MVP, data, LLM
- [ ] Budget caps per connector and etiquette compliance — Labels: MVP, infra, backend
- [ ] Lineage and citations in additional_content — Labels: MVP, ETL, data
- [ ] Runbook and incident handling for city jobs — Labels: MVP, infra, QA