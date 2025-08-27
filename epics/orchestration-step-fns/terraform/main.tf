# Aggregate module; resources defined across files for readability.
# See sfn.tf, sqs.tf, s3.tf, iam.tf, cloudwatch.tf.

module "sfn" {
  source = "./sfn"

  state_machine_name = local.step_function_name
  sfn_role_arn       = aws_iam_role.sfn_role.arn

  lambda_discover_web_sources_arn = var.lambda_discover_web_sources_arn != "" ? var.lambda_discover_web_sources_arn : aws_lambda_function.discover_web_sources.arn
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

  # Feature flag variables
  mock_lambda_arn        = var.mock_lambda_arn
  state_implementations  = var.state_implementations

  # Observability variables
  sfn_log_level              = var.sfn_log_level
  sfn_include_execution_data = var.sfn_include_execution_data
  enable_sfn_tracing         = var.enable_sfn_tracing
  sfn_log_group_arn          = var.sfn_log_level != "OFF" ? "${aws_cloudwatch_log_group.sfn[0].arn}:*" : ""
}

# Optional: CloudWatch dashboards and alarms (enabled only when flags are set)
# Note: Observability module has pre-existing issues, disabled for now
# module "observability" {
#   count  = var.enable_observability && (var.enable_cloudwatch_dashboards || var.enable_cloudwatch_alarms) ? 1 : 0
#   source = "../../../infra/observability"
#
#   dashboard_name      = "${local.step_function_name}-dashboard"
#   state_machine_arn   = module.sfn.state_machine_arn
#   frontier_queue_name = aws_sqs_queue.frontier.name
#   dlq_queue_name      = aws_sqs_queue.frontier_dlq.name
#   aws_region          = var.aws_region
#
#   # Optional alarm email notification
#   alarm_email = ""
#
#   # Alarm thresholds (can be customized via variables if needed)
#   dlq_depth_threshold         = 10
#   execution_failure_threshold = 5
#   error_rate_threshold        = 10
#   budget_cap_threshold        = 90
# }