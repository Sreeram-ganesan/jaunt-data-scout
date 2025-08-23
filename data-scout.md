---

# Product Requirements Document (PRD)

**Project:** Data Scout — City-wide Place Discovery Engine
**Owner:** \[Your Name / Team]
**Date:** 2025-08-23
**Status:** Draft
-----------------

## 1. Overview

Build a headless service that, given a **city**, discovers and maintains a canonical list of **places** (primary must-see POIs and secondary nearby/hidden gems) with **name, latitude, longitude, address, categories, and provenance**. The system explores a city with a **hybrid strategy**: seed high-signal primaries, **chain-reaction** expand via nearby search, and **tile sweep** under-covered areas using **H3 geo-cells**.

**MVP data sources (no Google Places):**

* **OpenStreetMap / Overpass**, **OpenTripMap**, **Wikidata/Wikipedia**
* **Tavily** for compliant **search / extract / crawl** of city open-data portals, heritage registries, public-art inventories, etc.
* **LLM (Gemini)** for structured extraction (T5 prompts), triage, and tie-breaks
* **Nominatim** fallback geocoding (only when coordinates missing)

**Infra:** AWS Step Functions (orchestration), SQS (frontier), ECS/Lambda (workers), S3 (raw cache + manifests), **existing Postgres** (canonical store, see §6.2 mapping).

**Primary vs Secondary:** Each city targets **150–200 primaries** (ratings/ranks relevant) and **\~95% secondaries** (adjacent coverage; ratings optional).

---

## 2. Goals & Objectives

* ✅ **Coverage (Edinburgh MVP):** ≥ **4,000** unique places (primary + secondary) within city boundary
* ✅ **Quality:** ≤ **3%** duplicates; ≥ **98%** coordinates within 50 m; ≥ **95%** valid addresses
* ✅ **Hidden gems:** ≥ **20%** from niche/authority sources (HMDB, Atlas Obscura, heritage/open-data)
* ✅ **Cost/Time:** Full city crawl **≤ 6 hours**; variable cost target **≤ \$30** per city run (ex-infra)
* ✅ **Extendable:** City-agnostic discovery via **Tavily + LLM query packs** (no hardcoded domains)
* ✅ **Reusability:** Clear APIs and lineage for downstream planning/search/RAG

---

## 3. Scope

### 3.1 In Scope (MVP)

* **City resolution:** Admin polygon + centroid (OSM/Nominatim)
* **Discovery pipeline:**

  * **Tavily+LLM City Data Targets** manifest (open-data portals & datasets)
  * **Seeding primaries** from OSM/Overpass, OpenTripMap, Wikidata/Wikipedia, open-data endpoints
  * **Chain-reaction expansion** around primaries (adaptive 300–800 m); create **NEAR** edges
  * **Tile sweep** under-dense H3 cells (res 9/10)
* **Extraction & normalization:**

  * Prefer **machine-readable** (CKAN/Socrata/ArcGIS/GeoJSON)
  * For HTML pages, use **T5 extractor** (Gemini) → strict JSON
  * Country-aware address normalization
* **Validation & dedupe:** polygon containment, ≤60 m proximity, token-set name sim ≥0.90, address checks, Wikidata aliases; LLM tie-break only when needed
* **Ranking:** balanced composite for **primaries**; **adjacency scoring** for secondaries (no ratings required)
* **Storage:** write to existing Postgres tables (`t_locations_prd`, `t_location_coordinates`, `t_locations_prd_details`, `t_location_neighborhoods`); raw caches + manifest to S3
* **APIs:** jobs, list by city (primary/secondary), place detail (+ lineage), nearby secondaries
* **Observability:** metrics (yield, dupes, coverage, cost estimate), traces, alerts

### 3.2 Out of Scope (Future Phases)

* Google Places/Details/Photos integration
* Full photo/media pipeline
* Neo4j graph store, OpenSearch index (Phase 2+)
* Real-time UGC, ads/sponsored ranking, full operator UI

---

## 4. Success Criteria

* [ ] **Edinburgh** job completes and persists **≥ 4,000** places passing validation
* [ ] **≤ 3%** dupes; **≥ 98%** coord accuracy (50 m); **≥ 95%** valid addresses (spot checks n=100)
* [ ] **City Data Targets Manifest v1** saved to S3 (queries, shortlisted candidates, datasets, citations)
* [ ] **Primary set** identified (150–200) with **≥ 90%** having non-null popularity/authority fields
* [ ] **Secondary share ≥ 95%** of total; **≥ 15** secondaries within **800 m** of each primary on average
* [ ] API p95 ≤ **300 ms** for list/detail on 10k-place cities
* [ ] Postgres populated using **existing schema** and mappings in §6.2

---

## 5. Architecture

### 5.1 Components

* **Orchestrator:** AWS Step Functions workflow
  `Discover City Data Targets → Seed → Expand → Tile Sweep → Dedupe/Canonicalize → Persist → Finalize`
* **Frontier Queue:** SQS (lat/lng + radius + category jobs; DLQ for failures)
* **Workers:** ECS Fargate (or Lambda) running connectors/extractors/validators
* **Connectors:** Overpass, OpenTripMap, Wikidata/Wikipedia, CKAN/Socrata/ArcGIS/GeoJSON readers
* **Search & Crawl:** Tavily (search/extract/crawl) with robots/ToS compliance
* **LLM:** Gemini for T1–T4 prompts (query pack, triage, extraction) and dedupe tie-breaks
* **Storage:** S3 (raw responses + manifest), Postgres (canonical), optional Glue/Athena parquet
* **Observability:** CloudWatch metrics/alarms, OTEL traces

### 5.2 Data Flow (high level)

1. **City Targets** (T1–T4 + Tavily) → **manifest.json** (S3)
2. **Seed** primaries from OSM/OTM/Wiki/Open-data → normalized candidates
3. **Expand** neighbors per seed; **Tile sweep** sparse H3 cells
4. **Validate & Dedupe** → canonicalize + score
5. **Persist** to Postgres + lineage; write raw to S3
6. **APIs** serve city lists, details, nearby secondaries

---

## 6. Deliverables

### 6.1 Public API (JSON)

**List by city**

```
GET /v1/cities/{city}/places?filter=primary|secondary|all&category=&limit=&cursor=
200 [
  {"contentid":"UUIDv7","title":"Calton Hill","latitude":55.955,"longitude":-3.182,
   "category":"Landmark","subcategory":"Viewpoint","primary_data":true,
   "average_rating":0.82,"total_reviews":null}
]
```

**Place detail**

```
GET /v1/places/{id}
200 {
  "contentid":"UUIDv7","title":"Calton Hill","latitude":55.955,"longitude":-3.182,
  "address":"Edinburgh EH7 5AA, UK","city":"Edinburgh","country":"UK",
  "category":"Landmark","subcategory":"Viewpoint","primary_data":true,
  "score":0.87,"score_breakdown":{"popularity":0.45,"authority":0.18,"geo":0.10,"novelty":0.06,"graph":0.08},
  "external_refs":{"osm":["node/..."],"wikidata":"Q...","opentripmap_id":"..."},
  "lineage":[{"source":"OSM","source_id":"node/123","retrieved_at":"..."}]
}
```

**Nearby secondaries for a primary**

```
GET /v1/places/{id}/nearby?type=secondary&radius=800&limit=50
200 [{"contentid":"UUIDv7","title":"Nelson Monument","distance_m":240,"category":"Monument"}]
```

**Jobs**

```
POST /v1/jobs/city-discovery { "city":"Edinburgh","budget":{"max_calls":5000} }
GET  /v1/jobs/{id} -> {status, counts, coverage, errors}
```

### 6.2 Postgres Mapping (use existing schema)

* **t\_locations\_prd** (header)

  * `contentid (UUID)` ← place\_id
  * `title` ← name
  * `city/state/country/borough/neighborhood/zipcode` ← normalized address
  * `primary_data` ← true for primaries, false for secondaries
  * `content_rank` ← primary composite score (optional; also in details JSON)
  * `average_rating`, `total_reviews` ← OpenTripMap/Wiki proxies (if present)

* **t\_location\_coordinates**

  * `contentid` link; set `latitude`, `longitude`, bbox fields if known

* **t\_locations\_prd\_details**

  * `contentid` link; `externalid` (preferred source id, e.g., OSM or OTM)
  * `address`, `address_detailed`; `category`, `subcategory`
  * `h3_index_8/9/12` precomputed
  * `additional_content` JSONB stores:

    * `external_refs` (osm ids, opentripmap\_id, wikidata, hmdb\_id, atlas\_obscura\_url, city\_portal\_url)
    * `signals` (has\_wikipedia, niche\_source, novelty, photo\_count)
    * `score_breakdown` (per factor)
    * `lineage` (source ids/urls, retrieved\_at, parser version)
    * `crawler` (tavily\_queries, citations)
    * `secondary_signals` (anchors: \[{primary\_id, distance\_m}], adjacency\_score)

* **t\_location\_neighborhoods** for neighborhood/borough links

**Indexes (recommended)**

```sql
CREATE INDEX IF NOT EXISTS idx_locations_city_title ON t_locations_prd (city, title);
CREATE INDEX IF NOT EXISTS idx_loc_h3_9  ON t_locations_prd_details (h3_index_9);
CREATE INDEX IF NOT EXISTS idx_loc_h3_12 ON t_locations_prd_details (h3_index_12);
CREATE INDEX IF NOT EXISTS idx_loc_coords_latlng ON t_location_coordinates (latitude, longitude);
```

### 6.3 City Data Targets Manifest (JSON schema)

```json
{
  "city": "Edinburgh",
  "country": "UK",
  "query_pack_version": "v1",
  "generated_at": "YYYY-MM-DDThh:mm:ssZ",
  "queries": [{"query":"...", "intent":"open_data|heritage|public_art|parks|tourism|transport|wikidata", "notes":"..."}],
  "candidates": [{"url":"...","title":"...","category":"open_data","domain_authority":"gov","api_type":"CKAN","has_machine_readable":true,"license":"...","last_updated_hint":"YYYY-MM","score":0.84,"reasons":["..."]}],
  "datasets": [{"source_url":"...","title":"...","category":"public_art","endpoints":[{"type":"CSV|JSON|GeoJSON|Socrata|CKAN|ArcGIS-FeatureService|ArcGIS-MapServer|WFS|WMS","url":"..."}],"fields_hint":["name","latitude","longitude","address","category","id"],"updated":"YYYY-MM","license":"..."}]
}
```

### 6.4 Primary vs Secondary Strategy

* **Targets per city:** **150–200 primaries**; **≥95% secondaries** by count
* **Primaries:** aggressive multi-source discovery; composite score to `content_rank`
* **Secondaries:** adjacency-only scoring (proximity to primaries, novelty, coverage need); ratings optional
* **API:** `/v1/cities/{city}/places?filter=primary` and `/v1/places/{id}/nearby?type=secondary&radius=800`

### 6.5 Ranking (balanced profile)

```
score =  w_pop*popularity(OpenTripMap etc.)
       + w_auth*authority(Wikidata/Wikipedia/official registries)
       + w_geo *geo_centrality(1/(1+distance_to_centroid_km))
       + w_nov *novelty(niche/heritage/open-data)
       + w_net *graph_centrality(NEAR degree)
```

* **Secondaries:** `adjacency_score = 0.6*proximity + 0.2*novelty + 0.2*coverage_need`

### 6.6 Dedupe Rules

* Polygon containment, **≤60 m** great-circle distance
* Token-set name similarity **≥0.90** (normalized)
* Address equality/overlap; Wikidata aliases/redirects
* Gemini tie-break for conflicted merges (bounded JSON prompts)

### 6.7 Orchestration States (MVP)

* `DiscoverTargets` → `SeedPrimaries` → `ExpandNeighbors` → `TileSweep` → `DedupeCanonicalize` → `Persist` → `Finalize`
* SQS frontier; DLQ on failures; S3 raw cache keyed by `(source, request_hash)`

### 6.8 Config & Budgets (defaults)

```yaml
city: Edinburgh
h3_res: 9
seed_top_n: 200
radius_m_primary: 600
radius_m_secondary: 400
budget:
  max_api_calls: 5000
  max_wall_clock_hours: 6
  min_new_unique_rate: 0.05
sources:
  osm: true
  opentripmap: true
  wikidata: true
  tavily: true
  city_open_data: true
llm:
  use_extractor: true
  use_tiebreak: selective
refresh:
  primaries: weekly
  secondaries: monthly
```

### 6.9 QA Checklist

* Random sample (n=100): **≥98** correct coords (≤50 m), **≥95** valid addresses
* Dupe audit: **≤3%** merged/over-merged
* H3 coverage: **≥95%** cells with ≥1 place; **≥15** secondaries within 800 m of each primary on average
* Lineage present with citations; manifest saved to S3

---

## 7. Risks & Mitigations

* **Variable open-data quality** → prioritize machine-readable endpoints; use validators; fall back to Nominatim + Overpass verify
* **Rate limits / robots** → per-connector token buckets, backoff & jitter, polite crawling via Tavily
* **Over-merging dupes** → conservative thresholds; LLM tie-break only for conflicts; keep lineage for reversibility
* **Cost drift** → cap per-cell calls; cache raw responses; early stop on diminishing returns

---

## 8. Implementation Plan (MVP)

* **Week 1:** Step Functions + SQS + S3; Overpass/OTM/Wiki connectors; Postgres writers; H3 utils
* **Week 2:** Tavily query pack + manifest; CKAN/Socrata/ArcGIS readers; T5 extractor + validators + Nominatim fallback
* **Week 3:** Dedupe pipeline; ranking; APIs; indexes; metrics/alerts; QA harness
* **Week 4:** Tile sweep tuning; Edinburgh run; acceptance checks; docs

---

## 9. Open Questions

1. Use `t_locations_prd.content_rank` for primary score or keep only in `additional_content.score_breakdown`?
2. Persist per-source payloads in a new table vs. only `additional_content.lineage`?
3. Include restaurants now (route to existing restaurant tables) or exclude in MVP?
4. Defer Neo4j/OpenSearch to Phase 2?
5. Edinburgh-specific open-data sets to prioritize first run?
6. Connector-specific budget ceilings (Tavily, Overpass etiquette, OTM) for MVP?

---

