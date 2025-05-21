resource "aws_ssm_parameter" "cookies_txt" {
  name   = "/audio-backend/cookies.txt"
  type   = "SecureString"
  value  = file("${path.module}/cookies.txt")
  key_id = "alias/aws/ssm"
}

resource "aws_ssm_parameter" "db_username" {
  name  = "/rds/db_username"
  type  = "SecureString"
  value = "sabbac"
}

resource "aws_ssm_parameter" "db_password" {
  name  = "/rds/db_password"
  type  = "SecureString"
  value = "pa$$word1"
}

resource "aws_ssm_parameter" "db_name" {
  name  = "/rds/db_name"
  type  = "String"
  value = "sabbac"
}

resource "aws_ssm_parameter" "rds_host" {
  name  = "/rds/db_host"
  type  = "String"
  value = aws_db_instance.rds_instance.address
}

resource "aws_ssm_parameter" "rds_port" {
  name  = "/rds/db_port"
  type  = "String"
  value = aws_db_instance.rds_instance.port
}