resource "aws_security_group" "rds_security_group" {
  name        = "rds-security-group"
  vpc_id      = aws_vpc.audio_backend_vpc.id

  ingress {
    from_port   = 3306
    to_port     = 3306
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_subnet_group" "rds_subnet_group" {
  name       = "rds-subnet-group"
  subnet_ids = [
    aws_subnet.private_subnet_1.id,
    aws_subnet.private_subnet_2.id
  ]

  tags = {
    Name = "rds-subnet-group"
  }
}

resource "aws_db_instance" "rds_instance" {
  identifier        = "rds-instance"
  engine            = "mysql"
  engine_version    = "8.0"
  instance_class    = "db.t3.micro"
  allocated_storage = 20
  storage_type      = "gp2"

  username = aws_ssm_parameter.db_username.value
  password = aws_ssm_parameter.db_password.value
  db_name  = aws_ssm_parameter.db_name.value

  db_subnet_group_name = aws_db_subnet_group.rds_subnet_group.name

  port                 = 3306
  publicly_accessible  = false
  skip_final_snapshot  = true
  vpc_security_group_ids = [aws_security_group.rds_security_group.id]

  tags = {
    Name = "rds-instance"
  }
}