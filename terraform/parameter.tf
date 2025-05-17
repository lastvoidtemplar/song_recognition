resource "aws_ssm_parameter" "cookies_txt" {
  name   = "/audio-backend/cookies.txt"
  type   = "SecureString"
  value  = file("${path.module}/cookies.txt")
  key_id = "alias/aws/ssm"
}