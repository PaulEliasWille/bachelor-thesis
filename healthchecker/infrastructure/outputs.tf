output "api_url" {
  value = aws_apigatewayv2_stage.api.invoke_url
}

output "dashboard_url" {
  value = aws_apigatewayv2_stage.dashboard.invoke_url
}

output "api_key" {
  value = replace(replace(base64encode(random_password.api_key.result), "+", "-"), "/", "_")
  sensitive = true
}