# Aggregate module; resources defined across files for readability.
# See sfn.tf, sqs.tf, s3.tf, iam.tf, cloudwatch.tf.

module "sfn" {
  source = "./sfn"

  state_machine_name = local.step_function_name
  sfn_role_arn       = aws_iam_role.sfn_role.arn

  lambda_discover_web_sources_arn = var.lambda_discover_web_sources_arn
  lambda_discover_targets_arn     = var.lambda_discover_targets_arn
  lambda_seed_primaries_arn       = var.lambda_seed_primaries_arn
  lambda_expand_neighbors_arn     = var.lambda_expand_neighbors_arn
  lambda_tile_sweep_arn           = var.lambda_tile_sweep_arn
  lambda_web_fetch_arn            = var.lambda_web_fetch_arn
  lambda_extract_with_llm_arn     = var.lambda_extract_with_llm_arn
  lambda_geocode_validate_arn     = var.lambda_geocode_validate_arn
  lambda_dedupe_canonicalize_arn  = var.lambda_dedupe_canonicalize_arn
  lambda_persist_arn              = var.lambda_persist_arn
  lambda_rank_arn                 = var.lambda_rank_arn

  frontier_dlq_url = aws_sqs_queue.frontier_dlq.id
}