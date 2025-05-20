output "cloudfront_url" {
  value = "https://${aws_cloudfront_distribution.frontend_cloudfront.domain_name}"
}

output "ec2_public_ip" {
  value = aws_instance.audio_backend_web_api.public_ip
}

output "api_gateway_url" {
    value = aws_apigatewayv2_api.http_api.api_endpoint
}