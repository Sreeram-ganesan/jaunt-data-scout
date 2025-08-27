locals {
  # Apply feature flags to determine which Lambda ARN to use for each state
  effective_lambda_arns = {
    discover_web_sources   = var.state_implementations.discover_web_sources == "mock" ? var.mock_lambda_arn : var.lambda_discover_web_sources_arn
    discover_targets       = var.state_implementations.discover_targets == "mock" ? var.mock_lambda_arn : var.lambda_discover_targets_arn
    seed_primaries        = var.state_implementations.seed_primaries == "mock" ? var.mock_lambda_arn : var.lambda_seed_primaries_arn
    expand_neighbors      = var.state_implementations.expand_neighbors == "mock" ? var.mock_lambda_arn : var.lambda_expand_neighbors_arn
    tile_sweep           = var.state_implementations.tile_sweep == "mock" ? var.mock_lambda_arn : var.lambda_tile_sweep_arn
    web_fetch            = var.state_implementations.web_fetch == "mock" ? var.mock_lambda_arn : var.lambda_web_fetch_arn
    extract_with_llm     = var.state_implementations.extract_with_llm == "mock" ? var.mock_lambda_arn : var.lambda_extract_with_llm_arn
    geocode_validate     = var.state_implementations.geocode_validate == "mock" ? var.mock_lambda_arn : var.lambda_geocode_validate_arn
    dedupe_canonicalize  = var.state_implementations.dedupe_canonicalize == "mock" ? var.mock_lambda_arn : var.lambda_dedupe_canonicalize_arn
    persist              = var.state_implementations.persist == "mock" ? var.mock_lambda_arn : var.lambda_persist_arn
    rank                 = var.state_implementations.rank == "mock" ? var.mock_lambda_arn : var.lambda_rank_arn
  }

  asl = templatefile("${path.module}/definition.asl.json", {
    lambda_discover_web_sources_arn = local.effective_lambda_arns.discover_web_sources,
    lambda_discover_targets_arn     = local.effective_lambda_arns.discover_targets,
    lambda_seed_primaries_arn       = local.effective_lambda_arns.seed_primaries,
    lambda_expand_neighbors_arn     = local.effective_lambda_arns.expand_neighbors,
    lambda_tile_sweep_arn           = local.effective_lambda_arns.tile_sweep,
    lambda_web_fetch_arn            = local.effective_lambda_arns.web_fetch,
    lambda_extract_with_llm_arn     = local.effective_lambda_arns.extract_with_llm,
    lambda_geocode_validate_arn     = local.effective_lambda_arns.geocode_validate,
    lambda_dedupe_canonicalize_arn  = local.effective_lambda_arns.dedupe_canonicalize,
    lambda_persist_arn              = local.effective_lambda_arns.persist,
    lambda_rank_arn                 = local.effective_lambda_arns.rank,
    frontier_dlq_url                = var.frontier_dlq_url
  })
}

resource "aws_sfn_state_machine" "orchestrator_v2" {
  name       = var.state_machine_name
  role_arn   = var.sfn_role_arn
  definition = local.asl

  tracing_configuration {
    enabled = var.enable_sfn_tracing
  }

  dynamic "logging_configuration" {
    for_each = var.sfn_log_level == "OFF" ? [] : [1]
    content {
      include_execution_data = var.sfn_include_execution_data
      level                  = var.sfn_log_level
      log_destination        = var.sfn_log_group_arn
    }
  }
}

