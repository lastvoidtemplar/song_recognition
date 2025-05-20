resource "aws_apigatewayv2_api" "http_api" {
  name          = "api-gateway-http"
  protocol_type = "HTTP"
  
  cors_configuration {
    allow_origins = ["*"]
    allow_methods = ["*"]
    allow_headers = ["*"]
    expose_headers = ["*"]
    max_age = 3600
  }
}

resource "aws_apigatewayv2_integration" "proxy_integration" {
  api_id             = aws_apigatewayv2_api.http_api.id

  integration_type   = "HTTP_PROXY"
  integration_uri    = "http://${aws_instance.audio_backend_web_api.public_ip}/{proxy}"
  integration_method = "ANY"
}

resource "aws_apigatewayv2_route" "default_route" {
  api_id    = aws_apigatewayv2_api.http_api.id
  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.proxy_integration.id}"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.http_api.id
  name        = "$default"
  auto_deploy = true
}  


