locals {
  asl = templatefile("${path.module}/definition.asl.json", {
    lambda_discover_web_sources_arn = var.lambda_discover_web_sources_arn,
    lambda_discover_targets_arn     = var.lambda_discover_targets_arn,
    lambda_seed_primaries_arn       = var.lambda_seed_primaries_arn,
    lambda_expand_neighbors_arn     = var.lambda_expand_neighbors_arn,
    lambda_tile_sweep_arn           = var.lambda_tile_sweep_arn,
    lambda_web_fetch_arn            = var.lambda_web_fetch_arn,
    lambda_extract_with_llm_arn     = var.lambda_extract_with_llm_arn,
    lambda_geocode_validate_arn     = var.lambda_geocode_validate_arn,
    lambda_dedupe_canonicalize_arn  = var.lambda_dedupe_canonicalize_arn,
    lambda_persist_arn              = var.lambda_persist_arn,
    lambda_rank_arn                 = var.lambda_rank_arn,
    frontier_dlq_url                = var.frontier_dlq_url
  })
}

resource "aws_sfn_state_machine" "orchestrator_v2" {
  name       = var.state_machine_name
  role_arn   = var.sfn_role_arn
  definition = local.asl
}