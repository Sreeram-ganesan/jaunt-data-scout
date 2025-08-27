# CloudWatch Dashboard
resource "aws_cloudwatch_dashboard" "jaunt_observability" {
  dashboard_name = var.dashboard_name

  dashboard_body = templatefile("${path.module}/dashboard.json", {
    state_machine_arn    = var.state_machine_arn
    frontier_queue_name  = var.frontier_queue_name
    dlq_queue_name       = var.dlq_queue_name
    aws_region          = var.aws_region
  })
}

# SNS Topic for alarm notifications (optional)
resource "aws_sns_topic" "alarms" {
  count = var.alarm_email != "" ? 1 : 0
  name  = "jaunt-data-scout-alarms"
}

resource "aws_sns_topic_subscription" "email_alerts" {
  count     = var.alarm_email != "" ? 1 : 0
  topic_arn = aws_sns_topic.alarms[0].arn
  protocol  = "email"
  endpoint  = var.alarm_email
}

# CloudWatch Alarms
resource "aws_cloudwatch_metric_alarm" "dlq_depth" {
  alarm_name          = "jaunt-dlq-depth-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = "300"
  statistic           = "Average"
  threshold           = var.dlq_depth_threshold
  alarm_description   = "This metric monitors DLQ depth"
  alarm_actions       = var.alarm_email != "" ? [aws_sns_topic.alarms[0].arn] : []

  dimensions = {
    QueueName = var.dlq_queue_name
  }

  tags = {
    Project = "jaunt-data-scout"
  }
}

resource "aws_cloudwatch_metric_alarm" "execution_failures" {
  alarm_name          = "jaunt-step-functions-failures"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "ExecutionsFailed"
  namespace           = "AWS/StepFunctions"
  period              = "300"
  statistic           = "Sum"
  threshold           = var.execution_failure_threshold
  alarm_description   = "This metric monitors Step Functions execution failures"
  alarm_actions       = var.alarm_email != "" ? [aws_sns_topic.alarms[0].arn] : []

  dimensions = {
    StateMachineArn = var.state_machine_arn
  }

  tags = {
    Project = "jaunt-data-scout"
  }
}

resource "aws_cloudwatch_metric_alarm" "error_rate_spike" {
  alarm_name          = "jaunt-error-rate-spike"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "ErrorRate"
  namespace           = "JauntDataScout"
  period              = "300"
  statistic           = "Average"
  threshold           = var.error_rate_threshold
  alarm_description   = "This metric monitors error rate spikes"
  alarm_actions       = var.alarm_email != "" ? [aws_sns_topic.alarms[0].arn] : []

  # This alarm uses a metric math expression to calculate error rate
  metric_query {
    id          = "e1"
    expression  = "m2/m1*100"
    label       = "Error Rate"
    return_data = "true"
  }

  metric_query {
    id = "m1"

    metric {
      metric_name = "Calls"
      namespace   = "JauntDataScout"
      period      = "300"
      stat        = "Sum"

      dimensions = {
        Service = "orchestrator"
      }
    }
  }

  metric_query {
    id = "m2"

    metric {
      metric_name = "Errors"
      namespace   = "JauntDataScout"
      period      = "300"
      stat        = "Sum"

      dimensions = {
        Service = "orchestrator"
      }
    }
  }

  tags = {
    Project = "jaunt-data-scout"
  }
}

resource "aws_cloudwatch_metric_alarm" "budget_cap_nearing" {
  alarm_name          = "jaunt-budget-cap-nearing"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "BudgetCapUtilization"
  namespace           = "JauntDataScout"
  period              = "300"
  statistic           = "Maximum"
  threshold           = var.budget_cap_threshold
  alarm_description   = "This metric monitors budget cap utilization"
  alarm_actions       = var.alarm_email != "" ? [aws_sns_topic.alarms[0].arn] : []

  dimensions = {
    Service = "orchestrator"
  }

  tags = {
    Project = "jaunt-data-scout"
  }
}