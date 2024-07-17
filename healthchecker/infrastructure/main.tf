locals {
  region = data.aws_region.current.name
  account_id = data.aws_caller_identity.current.account_id

  name = "healthcheck-${random_id.this.hex}"

  schedule = "rate(1 minute)"

  api_key = replace(replace(base64encode(random_password.api_key.result), "+", "-"), "/", "_")

  environment_variables = {
    "WORKER_QUEUE_URL" =  aws_sqs_queue.worker.url
    "HEALTH_CHECK_TABLE" = aws_dynamodb_table.health_check.name
    "HEALTH_CHECK_RESULT_TABLE" = aws_dynamodb_table.health_check_result.name
    "HEALTH_CHECK_RESULT_TABLE_HEALTH_CHECK_ID_INDEX" = "HealthCheckIdIndex"
    "API_KEY" = local.api_key
    "DISPATCH_INTERVAL_SECONDS" = 60
  }
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "random_id" "this" {
  byte_length = 2
}

resource "random_password" "api_key" {
  length           = 32
  special          = false
}

##############
### Shared ###
##############

resource "aws_sqs_queue" "worker" {
  name = "${local.name}-worker"
}

resource "aws_dynamodb_table" "health_check" {
  name           = "${local.name}-health-check"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "healthCheckId"

  attribute {
    name = "healthCheckId"
    type = "S"
  }
}

resource "aws_dynamodb_table" "health_check_result" {
  name           = "${local.name}-health-check-result"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "healthCheckResultId"
  range_key      = "timestamp"

  attribute {
    name = "healthCheckResultId"
    type = "S"
  }

  attribute {
    name = "timestamp"
    type = "S"
  }

  attribute {
    name = "healthCheckId"
    type = "S"
  }

  global_secondary_index {
    name            = "HealthCheckIdIndex"
    hash_key        = "healthCheckId"
    projection_type = "ALL"
  }
}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "lambda" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = ["arn:aws:logs:*:*:*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "dynamodb:BatchGetItem",
      "dynamodb:BatchWriteItem",
      "dynamodb:ConditionCheckItem",
      "dynamodb:CreateTable",
      "dynamodb:DeleteItem",
      "dynamodb:DescribeTable",
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:Query",
      "dynamodb:Scan",
      "dynamodb:UpdateItem",
    ]
    resources = [
      aws_dynamodb_table.health_check.arn,
      "${aws_dynamodb_table.health_check.arn}/index/*",      
      aws_dynamodb_table.health_check_result.arn,
      "${aws_dynamodb_table.health_check_result.arn}/index/*",
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "sqs:CreateQueue",
      "sqs:DeleteQueue",
      "sqs:ListQueues",
      "sqs:SendMessage",
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:PurgeQueue",
      "sqs:GetQueueAttributes",
      "sqs:SetQueueAttributes",
      "sqs:AddPermission",
      "sqs:RemovePermission",
      "sqs:ListQueueTags",
      "sqs:TagQueue",
      "sqs:UntagQueue",
    ]
    resources = [aws_sqs_queue.worker.arn]
  }
}

resource "aws_iam_policy" "lambda" {
  name        = "${local.name}-lambda"
  policy      = data.aws_iam_policy_document.lambda.json
}

resource "aws_iam_role" "lambda" {
  name               = "${local.name}-lambda"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

resource "aws_iam_role_policy_attachment" "lambda" {
  role       = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.lambda.arn
}

#########################
### API and Dashboard ###
#########################

data "archive_file" "api_lambda" {
  type        = "zip"
  source_dir = "../application/dist/api"
  output_path = "../application/dist/api.zip"
}

data "archive_file" "dashboard_lambda" {
  type        = "zip"
  source_dir = "../application/dist/dashboard"
  output_path = "../application/dist/dashboard.zip"
}

resource "aws_lambda_function" "api" {
  filename      = data.archive_file.api_lambda.output_path
  function_name = "${local.name}-api"
  role          = aws_iam_role.lambda.arn
  handler       = "api.handler"

  source_code_hash = data.archive_file.api_lambda.output_base64sha256

  runtime = "nodejs20.x"
  timeout = 30

  environment {
    variables = local.environment_variables
  }
}

resource "aws_lambda_function" "dashboard" {
  filename      = data.archive_file.dashboard_lambda.output_path
  function_name = "${local.name}-dashboard"
  role          = aws_iam_role.lambda.arn
  handler       = "dashboard.handler"

  source_code_hash = data.archive_file.dashboard_lambda.output_base64sha256

  runtime = "nodejs20.x"
  timeout = 30

  environment {
    variables = local.environment_variables
  }
}

resource "aws_apigatewayv2_api" "api" {
  name          = "${local.name}-api"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "api" {
  api_id = aws_apigatewayv2_api.api.id

  name        = "$default"
  auto_deploy = true
}

resource "aws_apigatewayv2_integration" "api" {
  api_id = aws_apigatewayv2_api.api.id

  integration_uri    = aws_lambda_function.api.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST"

  timeout_milliseconds = 30000
}

resource "aws_apigatewayv2_route" "api" {
  api_id = aws_apigatewayv2_api.api.id

  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.api.id}"
}

resource "aws_lambda_permission" "api" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.api.execution_arn}/*/*"
}

resource "aws_apigatewayv2_api" "dashboard" {
  name          = "${local.name}-dashboard"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "dashboard" {
  api_id = aws_apigatewayv2_api.dashboard.id

  name        = "$default"
  auto_deploy = true
}

resource "aws_apigatewayv2_integration" "dashboard" {
  api_id = aws_apigatewayv2_api.dashboard.id

  integration_uri    = aws_lambda_function.dashboard.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST"

  timeout_milliseconds = 30000
}

resource "aws_apigatewayv2_route" "dashboard" {
  api_id = aws_apigatewayv2_api.dashboard.id

  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.dashboard.id}"
}

resource "aws_lambda_permission" "dashboard" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.dashboard.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.dashboard.execution_arn}/*/*"
}

##################
### Dispatcher ###
##################

data "archive_file" "dispatcher_lambda" {
  type        = "zip"
  source_dir = "../application/dist/dispatcher"
  output_path = "../application/dist/dispatcher.zip"
}

resource "aws_lambda_function" "dispatcher" {
  filename      = data.archive_file.dispatcher_lambda.output_path
  function_name = "${local.name}-dispatcher"
  role          = aws_iam_role.lambda.arn
  handler       = "dispatcher.handler"

  source_code_hash = data.archive_file.dispatcher_lambda.output_base64sha256

  runtime = "nodejs20.x"
  timeout = 30

  environment {
    variables = local.environment_variables
  }
}

resource "aws_cloudwatch_event_rule" "dispatcher" {
  name = "${local.name}-schedule"
  schedule_expression = local.schedule
  # state = "DISABLED"
}

resource "aws_cloudwatch_event_target" "dispatcher" {
  rule = aws_cloudwatch_event_rule.dispatcher.name
  arn = aws_lambda_function.dispatcher.arn
}

resource "aws_lambda_permission" "dispatcher" {
  statement_id = "AllowExecutionFromCloudWatch"
  action = "lambda:InvokeFunction"
  function_name = aws_lambda_function.dispatcher.function_name
  principal = "events.amazonaws.com"
}

##############
### Worker ###
##############

data "archive_file" "worker_lambda" {
  type        = "zip"
  source_dir = "../application/dist/worker"
  output_path = "../application/dist/worker.zip"
}

resource "aws_lambda_function" "worker" {
  filename      = data.archive_file.worker_lambda.output_path
  function_name = "${local.name}-worker"
  role          = aws_iam_role.lambda.arn
  handler       = "worker.handler"

  source_code_hash = data.archive_file.worker_lambda.output_base64sha256

  runtime = "nodejs20.x"
  timeout = 30

  environment {
    variables = local.environment_variables
  }
}

resource "aws_lambda_event_source_mapping" "worker" {
  event_source_arn = aws_sqs_queue.worker.arn
  function_name = aws_lambda_function.worker.arn
}