# Jaunt Data Scout – MVP Epics and Tasks (v2, Google in scope + Web/LLM extensions)

This document enumerates the MVP Epics and their Tasks derived from the v2 System Design and the expanded web discovery + LLM extraction flow. Use it to drive issue creation and tracking.

How to use
1) Labels: create `MVP`, `backend`, `data`, `API`, `infra`, `LLM`, `Postgres`, `observability`, `ETL`, `QA`, `google`, `compliance`, `fastapi`, `crawler`, `web`.
2) Milestone: create `MVP (Edinburgh)`.
3) Create one issue per Epic below. Copy each Epic section's "Goals", "Acceptance criteria", and "Tasks" into the issue body. Apply relevant labels and the milestone.
4) Keep "Tasks" as a task list in the Epic issue. Convert each task list item to its own issue from the GitHub UI ("…" menu → Convert to issue). This preserves parent/child links and progress.
5) Use a GitHub Project to track work. Suggested views: by Epic, by label, by status; group by Milestone.
6) Reference issues in PRs with `closes #<n>` so they close on merge.

---

## Epic: Foundations & Enablers — Contracts, Budgets, Secrets, Observability (v2)
Labels: MVP, infra, backend, observability, compliance

Goals
- Establish the shared foundations so individual states can be swapped from mock → real without churn: schemas/contracts, budgets/config, secrets/IAM/egress, observability/alerts, and DLQ/runbooks.

Acceptance criteria
- JSON Schemas published for: frontier messages (maps/web union), canonical candidate, and web extraction outputs; CI contract tests validate producers/consumers.
- Config defaults (YAML) present for budgets/concurrency/early-stop with run-level overrides; feature flags can flip per-state mock/real in Step Functions.
- Secrets in AWS Secrets Manager and IAM least-privilege roles for each Lambda; VPC egress allowlists configured for WebFetch/LLM.
- Step Functions logging + tracing enabled; Lambdas emit EMF metrics and propagate correlation_id; CloudWatch dashboards/alarms in place.
- DLQ re-drive runbook and script exist; DLQ payload schema compatible with reprocessing path.

Tasks
- [x] Contracts (JSON Schema + docs + CI validation)
  - [x] Frontier message schemas (maps/web union) incl. optional trust_score, coordinates_confidence, correlation_id; examples in repo.
  - [x] Canonical candidate schema (normalized entity) with lineage and confidences.
  - [x] Web extraction output schema; strict validation and contract tests.
- [x] Config & budgets
  - [x] YAML defaults for budgets, concurrency, early-stop; environment overrides.
  - [x] Budget token taxonomy finalized (google.text, google.nearby, google.details, overpass, otm, wiki, tavily.api, web.fetch, llm.tokens, nominatim).
- [ ] Secrets/IAM/egress
  - [ ] Secrets Manager wiring (Google, Tavily, OTM, DB, LLM); rotation policy documented.
  - [ ] IAM least-privilege roles and S3 policies; SSE-S3/KMS and access logs enabled.
  - [ ] VPC egress allowlists and DNS for WebFetch/LLM; robots/ToS headers enforced at fetcher.
- [ ] Observability baseline
  - [ ] Enable Step Functions loggingConfiguration (ALL) and tracingConfiguration (X-Ray).
  - [ ] EMF metrics per Lambda: calls, errors, retries, duration_ms, http_bytes_in, tokens_in, tokens_out, token_cost_estimate, new_unique_rate.
  - [ ] Correlation_id propagation across SQS → Lambdas; include in logs and spans.
  - [ ] Dashboards + alarms: ExecutionFailed, state timeouts, DLQ depth, budget cap nearing (LLM/Tavily/HTTP), error spikes.
- [ ] DLQ & runbooks
  - [ ] Re-drive Lambda/CLI script to re-enqueue DLQ messages; safeguards to avoid duplicates.
  - [ ] Incident runbook for city jobs (triage, re-drive, rollback toggles).
- [ ] Feature flags
  - [ ] Per-state mock↔real flags via Terraform/template variables and Lambda env.
- [ ] Golden tests
  - [ ] Golden input presets and a web "smoke pack" (hand-curated URLs + expected entities) runnable in CI.

---

## Epic: Orchestration — Step Functions workflow for city discovery (v2)
Labels: MVP, infra, backend

Goals
- Implement AWS Step Functions orchestration for the MVP pipeline, including both maps and web paths:
  DiscoverWebSources → DiscoverTargets → SeedPrimaries → ExpandNeighbors → TileSweep → WebFetch → ExtractWithLLM → GeocodeValidate → DedupeCanonicalize → Persist → Rank → Finalize
- Integrate SQS frontier + DLQ, S3 raw cache, budgets, early-stop, and metrics/traces.

Acceptance criteria
- A city job completes end-to-end for Edinburgh using mixed connectors (Google + open data + web/LLM).
- State transitions logged; failures to DLQ; raw responses cached to S3 by (source, request_hash/content_hash).
- Budget controls enforced (per-connector caps & 70/30 primaries/secondaries); LLM token budget enforced; early-stop when new_unique_rate < 5% over last 200 calls; wall-clock ≤ 6 hours respected.
- CloudWatch metrics and OTEL traces emitted per state; kill switches supported via job params.
- Frontier message schema supports union types for maps and web with optional trust_score and coordinates_confidence.

Tasks
- [x] Define Step Functions state machine with all v2 states (including DiscoverWebSources, WebFetch, ExtractWithLLM, GeocodeValidate), retries/backoff, budget tokens, and early-stop gates.
- [x] Budget manager: token buckets per connector (google.text, google.nearby, google.details, overpass, otm, wiki, tavily.api, web.fetch, llm.tokens, nominatim) with 70/30 allocation controls.
- [x] Implement frontier SQS schemas (seed/expand/tile_sweep/open_data_pull + web) and DLQ; include budget_token, city context, and correlation_id.
- [x] S3 layout and lifecycle for raw/html, raw/json, extracted, manifests.
- [x] Orchestrator config: YAML defaults for budgets, concurrency, early-stop; run-level overrides; kill switches and circuit breakers; resume/reentry semantics.
- [x] Emit metrics/alerts per state; tracing context propagation across SQS and Step Functions; stitch traces across WebFetch and LLM extraction.
- [ ] Feature flags to flip individual states mock↔real (per-state env/vars); document toggling procedure.
- [ ] DLQ re-drive runbook and helper CLI; example reprocessing flow.
- [ ] Execution input presets (golden inputs) per city for smoke/e2e tests.

---

## Epic: Connectors — Google Places (Text, Nearby, Details)
Labels: MVP, data, backend, google, compliance

Goals
- Build Google Places connectors with token buckets, polite rate limits, S3 raw cache, and compliance (field masks, attribution, retention).

Acceptance criteria
- Text Search: category + keyword packs seeded by city; respects budgets and backoff.
- Nearby Search: adaptive radius (300–800 m) for expansion; enqueues NEAR edges.
- Details: primaries-only with must-have fields; optional nice-to-haves controlled by budget flags; no bulk photo storage (metadata only).
- Compliance: field masks enforced; "Powered by Google" attribution data captured; raw payload retention 30 days; ratings cache TTL 7 days.
- All connectors return normalized candidates with source lineage and idempotent request hashing.

Tasks
- [ ] Implement Text Search client + query builder for category/keyword packs; pagination and backoff with jitter.
- [ ] Implement Nearby Search with adaptive radius and anchor context; enqueue neighbors with NEAR edge hints.
- [ ] Implement Details with explicit field masks: must-haves (place_id, name, formatted_address, geometry/location, types, rating, user_ratings_total, business_status, photos.photo_reference, html_attributions).
- [ ] Photos: store metadata only (photo_reference + attributions); no bulk photo storage; capture attribution strings.
- [ ] Compliance guardrails: enforce field masks, response-filtering, retention tagging (30d), exclusion of Google-only fields from partner APIs by default; audit logs.
- [ ] Raw response cache keyed by request_hash; idempotency; retry/backoff; error taxonomy.
- [ ] Unit/contract tests using recorded fixtures; redaction of PII and API keys.

## Epic: Connectors — Open Data & Open Graph (OTM/OSM/Wiki/Tavily)
Labels: MVP, data, backend

Goals
- Build robust connectors for OSM/Overpass, OpenTripMap, Wikidata/Wikipedia, and Tavily-driven open-data sources with caching and rate limits.

Acceptance criteria
- Overpass: parameterized by city polygon; polite QPS; backoff.
- OpenTripMap: category/box/radius; popularity proxies; throttling.
- Wikidata/Wikipedia: SPARQL + page extracts; authority signals captured.
- Tavily: search/extract/crawl for dataset endpoints; robots-respecting.
- Normalized outputs with lineage; raw cached to S3.

Tasks
- [ ] Overpass client with city polygon and adaptive radius; pagination/chunking; cache and retry.
- [ ] OpenTripMap client with keys, pagination, and throttle; capture rate/popularity proxy.
- [ ] Wikidata SPARQL queries and Wikipedia summaries; authority scoring fields.
- [ ] Tavily integration (search/extract/crawl) with robots/ToS compliance.
- [ ] Raw cache keying; idempotency; retry/backoff; fixtures and unit tests.

## Epic: Web Discovery & Extraction — Tavily → WebFetch → LLM
Labels: MVP, LLM, data, web, crawler

Goals
- Discover city-relevant sources via Tavily, fetch pages/APIs, and extract structured POIs using an LLM extractor with bounded schema.

Acceptance criteria
- Tavily returns URL candidates with lineage (queries, citations); robots-respecting fetches; raw HTML/JSON cached to S3.
- LLM extraction emits strict JSON entities with fields: name, lat, lng (optional), category?, source, source_url, confidences; schema-validated.
- For missing/low-confidence coords, GeocodeValidate fills/validates via Maps APIs; coordinates_confidence recorded.
- Extraction jobs respect token/time/cost budgets; retries/backoff with idempotency.
- Optional trust_score assigned per source_url or domain and propagated to downstream ranking.

Tasks
- [ ] Tavily Connector: city-scoped queries; scoring and dedup of URL candidates; enqueue web messages to frontier with correlation_id.
- [ ] Web Fetcher: robots.txt compliance, polite concurrency/QPS, egress allowlist; cache raw HTML/JSON to S3 under raw/html and raw/json.
- [ ] LLM Extractor: provider-agnostic (Bedrock/OpenAI/Gemini) with bounded JSON schema, timeouts, token and cost guards; write extracted/<city>/<run_id>/<content_hash>.json.
- [ ] Error handling and idempotency: deterministic content_hash; dedupe repeated URLs; retry taxonomy; correlation_id propagation.
- [ ] Observability: metrics for http_bytes_in, pages_fetched, extractor_token_count, extractor_cost_estimate; traces across fetch→extract.
- [ ] NOTE: Extraction Schema & Canonical Candidate Schema are defined in Foundations and reused here.

## Epic: Tavily City Data Targets Manifest (v2)
Labels: MVP, LLM, data

Goals
- Generate a manifest of open-data portals, datasets, and endpoints per city using Tavily + prompt set T1–T4.

Acceptance criteria
- manifest.json for Edinburgh written to S3 with queries, candidates, datasets, endpoints, citations; versioned query_pack_version=v1.
- License/compliance captured when present; robots respected.

Tasks
- [ ] Author T1–T4 prompt set; integrate with Tavily search/extract/crawl.
- [ ] Candidate scoring and shortlist logic (authority, structure, freshness, license).
- [ ] Manifest JSON schema + validation; persist with lineage and citations.
- [ ] Store Tavily queries and citations in lineage.

## Epic: Discovery — Seed primaries, expand neighbors, tile sweep
Labels: MVP, data, backend

Goals
- Achieve target coverage with primary/secondary split using Google-anchored seeding, chain-reaction expansion, and H3 tile sweep.

Acceptance criteria
- 150–200 primaries selected; secondaries ≥95% by count.
- Average ≥15 secondaries within 800 m of each primary.
- H3 res 9/10 sweep fills under-dense cells.

Tasks
- [ ] Primary seeding from Google Text + OTM + Wikidata/Wikipedia + open-data endpoints + (optional) high-trust web-extracted entities; pre-rank primaries; schedule Google Details calls for primaries.
- [ ] Neighbor expansion around primaries via Google Nearby + OSM/OTM/open-data; create NEAR edges with distances.
- [ ] Under-dense H3 cell sweep (res 9/10) with coverage thresholds; configurable.
- [ ] Budget-aware frontier enqueuing (70/30 primaries/secondaries); early-stop on low new_unique_rate.

## Epic: Extraction & Normalization — Readers + LLM fallback
Labels: MVP, LLM, ETL

Goals
- Prefer machine-readable ingestion; fall back to LLM extractor for HTML when APIs absent; ensure country-aware normalization.

Acceptance criteria
- Country-aware address normalization; valid city/state/country mapping.
- Strict JSON outputs; schema validation; lineage captured.
- Nominatim reverse/geocode only as fallback with conservative usage.

Tasks
- [ ] CKAN/Socrata/ArcGIS/GeoJSON readers; schema mappers into canonical candidate.
- [ ] LLM extractor with bounded JSON schema (provider-agnostic), timeouts, and token/cost guards.
- [ ] Address normalizer; Nominatim client (≤1 QPS, ≤1k calls) as fallback only.
- [ ] Field mappers to canonical candidate schema; capture google/open-data/web-specific fields in additional_content sub-objects.

## Epic: Validation & Dedupe — Canonicalization pipeline
Labels: MVP, ETL, backend

Goals
- Enforce quality bars: ≤3% dupes; ≥98% coordinates within 50 m; ≥95% valid addresses.

Acceptance criteria
- Dedupe rules: polygon containment, ≤60 m proximity, token-set name similarity ≥0.90, address agreement, alias handling; LLM tie-break only for conflicts; reversible merges with lineage.
- coordinates_confidence computed/propagated; low-confidence geocodes flagged for review.

Tasks
- [ ] Similarity functions and thresholds (name/address/geo); configurable.
- [ ] Merge strategy, conflict resolution with LLM tie-break for top conflicts; reversible lineage.
- [ ] QA harness for random sample checks (n=100) and dupe audit with reports.

## Epic: Ranking — Primary composite and secondary adjacency scoring
Labels: MVP, data, backend

Goals
- Implement balanced primary score and secondary adjacency scoring; write to content_rank and details JSON.

Acceptance criteria
- Tunable weights: popularity, authority, geo, novelty, graph; scores persisted with breakdown.
- Secondary adjacency_score computed and stored with anchors.

Tasks
- [ ] Implement primary composite: w_pop*(Google rating + log(reviews+1), OTM rate) + w_auth*(Wikidata/Wikipedia/registries) + w_geo*(geo_centrality) + w_nov*(niche/open-data/web) + w_net*(graph_context).
- [ ] Compute graph centrality based on NEAR edges; parameterize weighting with city defaults.
- [ ] Implement secondary adjacency_score = 0.6*proximity + 0.2*novelty + 0.2*coverage_need; persist anchors/distances.

## Epic: Storage — Postgres writers and S3 raw cache/manifests (v2)
Labels: MVP, Postgres, backend

Goals
- Persist canonical places to existing schema; store raw payloads and manifests to S3 with retention and compliance constraints.

Acceptance criteria
- Tables populated: t_locations_prd, t_location_coordinates, t_locations_prd_details, t_location_neighborhoods.
- additional_content JSONB populated per spec:
  - external_refs (google_place_id, osm ids, otm id, wikidata, hmdb, atlas url, city_portal_url)
  - google (types, opening_hours, website, phone, photos: photo_reference + html_attributions)
  - web (source_url, source_domain, extraction_confidences, trust_score)
  - signals (has_wikipedia, niche_source, novelty, photo_count)
  - score_breakdown, lineage, secondary_signals
- Recommended indexes created; ratings cache TTL logic in DB writer (7 days window) if enabled.

Tasks
- [ ] DB mappers + upsert logic; bbox/h3 precompute (h3_index_8/9/12).
- [ ] External refs consolidation and preferred externalid selection (Google place_id primaries; OSM/OTM otherwise).
- [ ] SQL migrations to add recommended indexes if missing.
- [ ] S3 lifecycle policies: Google raw json 30d, open-data json 90d, web raw html 90d, extracted json 90d; encryption SSE-S3/KMS.
- [ ] Implement ratings cache policy (persist timestamp; skip refresh within 7 days); exclude Google-only fields from partner API responses by default.

## Epic: Public APIs — Jobs, list by city, place detail, nearby secondaries (FastAPI v1)
Labels: MVP, API, backend, fastapi, compliance

Goals
- Expose REST endpoints with p95 ≤ 300 ms on 10k-place cities; ensure Google compliance on response fields.

Acceptance criteria
- Endpoints implemented per design (jobs, list by city, place detail, nearby secondaries); cursor pagination; filters.
- Default responses exclude Google-only fields (ratings/phone/website/photos) for partner contexts unless explicitly allowed; include attribution when Google-sourced fields are present.
- Input validation and error semantics documented.

Tasks
- [ ] Implement FastAPI handlers + schema types (filter=primary|secondary|all, category, anchor_id, radius, cursor).
- [ ] DB access paths and indexes usage; materialized views optional for read performance.
- [ ] Response field filtering middleware: exclude Google-only fields by default; inject attribution strings when needed.
- [ ] Load/perf tests; caching where needed; p95 tracking and tuning.

## Epic: Observability — Metrics, traces, and alerts (v2)
Labels: MVP, observability, infra

Goals
- Provide visibility into yield, dupes, coverage, and cost; alerts for anomalies; distributed tracing.

Acceptance criteria
- CloudWatch dashboards and alarms per state and connector; OTEL traces stitched across workers (maps + web + LLM).
- Cost estimation metrics and budget breach alerts; new_unique_rate tracked.
- LLM-specific metrics (tokens_in, tokens_out, token_cost_estimate) and web fetch metrics (http_bytes_in, pages_fetched) exposed.

Tasks
- [ ] Define metrics schema: calls, errors, backoffs, new_unique_rate, dedupe ratio, H3 coverage, cost estimate, duration_ms, extractor_token_count, extractor_cost_estimate, http_bytes_in.
- [ ] Tracing propagation (OTEL) across orchestrator, SQS, connectors, LLM, DB writes; link WebFetch→ExtractWithLLM spans.
- [ ] Dashboards and alarms: state timeouts, DLQ depth, budget cap nearing (LLM tokens, Tavily, HTTP), DB write failures, under-coverage, error spikes.

## Epic: Security & Compliance — Google, Web & Data handling
Labels: MVP, infra, compliance

Goals
- Enforce secrets management, IAM least privilege, encryption, Google Maps Platform compliance, and web crawling/LLM safety.

Acceptance criteria
- Secrets in AWS Secrets Manager; IAM scoped for ECS/S3/DB; VPC egress controls.
- Google compliance: field masks applied, retention (30d raw), ratings cache (7d), attributions captured, no bulk photo storage.
- Web compliance: robots.txt respected; per-domain crawl etiquette; ToS observed; HTML retention bounded (90d); PII redaction in logs and fixtures.
- LLM safety: prompt/response logging with redaction, max tokens/rate limits, provider keys scoped; no storage of raw user-provided secrets in corpora.

Tasks
- [ ] Secrets Manager wiring (Google, Tavily, OTM, DB, LLM); rotation policy.
- [ ] IAM roles and S3 bucket policies; SSE-S3/KMS; access logs enabled; per-service egress allowlists.
- [ ] Compliance middleware/guards: enforce field masks, response filtering, retention tagging, attribution presence, robots compliance headers.
- [ ] Audit logs for Details calls (fields requested, purpose, timestamp) and LLM usage (model, tokens_in/out, cost).

## Epic: QA & Acceptance — Edinburgh MVP run (v2)
Labels: MVP, QA

Goals
- Execute full Edinburgh city run and meet success & compliance criteria.

Acceptance criteria
- ≥ 4,000 places passing validation; 150–200 primaries with ≥90% having non-null popularity/authority.
- ≤3% dupes; ≥98% coord accuracy; ≥95% valid addresses (spot checks n=100).
- City Data Targets Manifest v1 saved to S3.
- Compliance: Google raw retention ≤30 days verified; API responses filtered by default; attribution present where required; robots adherence verified for web fetches.

Tasks
- [ ] Tile sweep tuning and budget parameter review (70/30 split).
- [ ] Runbook for city job; incident handling and on-call playbooks.
- [ ] Golden-run validator: schema checks + basic KPI thresholds (dupes, coord accuracy, address validity) runnable from CI.
- [ ] Final acceptance checks, data quality sampling, and compliance audit.

---

## Standalone Tasks (optional quick-links)
Use these as child tasks under the relevant Epics, or as individual issues.

- [ ] H3 utilities and under-dense cell detection (res 9/10/12) — Labels: MVP, backend, data
- [ ] Address normalization and geo validation (country-aware) — Labels: MVP, ETL, data
- [ ] Dedupe similarity functions and thresholds (name/address/geo) — Labels: MVP, ETL, backend
- [ ] Postgres upserts and index creation — Labels: MVP, Postgres, backend
- [ ] API performance targets and load testing (p95 ≤ 300 ms) — Labels: MVP, API, observability
- [ ] Edinburgh-specific open-data sources triage — Labels: MVP, data, LLM
- [ ] Lineage and citations in additional_content — Labels: MVP, ETL, data
- [ ] Contract tests for Google/OTM/OSM/Socrata/CKAN/ArcGIS/Wikidata adapters — Labels: MVP, QA, backend

---

### Orchestration — Additional tasks (Go Lambdas and AWS clients)

- [ ] Replace mock-go Lambda with real Go handlers and AWS SDK clients
  - [ ] Implement SQS FrontierQueue client (Go) for Enqueue/Dequeue/DeadLetter with correlation_id in MessageAttributes.
  - [ ] Implement S3 RawCache client (Go) for Get/Put using the defined key layout and SSE.
  - [ ] Create Go Lambda handler(s) for Step Functions states (start with one shared handler, then split per state).
  - [ ] Wire Terraform lambda_*_arn variables to the deployed Go Lambdas; update ASL template to use these ARNs.
  - [ ] CI: build/test Go handlers (GOOS=linux, GOARCH=arm64), run unit tests, and package artifacts for deployment.
  - [ ] Observability: instrument handlers with EMF metrics (calls, errors, duration_ms, http_bytes_in, tokens_in/out) and X-Ray; include correlation_id in logs.
  - [ ] Gradually replace lambdas/mock-go references with the new per-state Go handlers.