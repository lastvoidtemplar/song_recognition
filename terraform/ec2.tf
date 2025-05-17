resource "aws_instance" "audio_backend_web_api" {
  ami           = "ami-0ef32de3e8ab0640e"
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.public_subnet_1.id

  vpc_security_group_ids = [aws_security_group.web_api_security_group.id]
  iam_instance_profile = aws_iam_instance_profile.ssm_instance_profile.name

  user_data = file("${path.module}/user_data/user_data.sh")

  associate_public_ip_address = true

  tags = {
    Name = "audio-backend-web-api"
  }
}


resource "aws_security_group" "web_api_security_group" {
  name        = "web-api-security-group"
  description = "Allow ingress HTTP and egress "
  vpc_id      = aws_vpc.audio_backend_vpc.id

  ingress {
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 80
    to_port     = 80
  }

  egress {
    protocol    = "all"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    to_port     = 0
  }

  tags = {
    Name = "web-api-security-group"
  }
}

resource "aws_iam_role" "ssm_role" {
  name = "ec2-ssm-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Principal = {
          Service = "ec2.amazonaws.com"
        },
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = {
    Name = "ec2-ssm-role"
  }
}

resource "aws_iam_policy" "ssm_read_cookies_policy" {
  name   = "ReadWelcomeTxt"
  policy = jsonencode({
    Version : "2012-10-17",
    Statement : [{
      Effect   : "Allow",
      Action   : "ssm:GetParameter",
      Resource : aws_ssm_parameter.cookies_txt.arn
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ssm_policy_ssm_managed_attachment" {
  role       = aws_iam_role.ssm_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_role_policy_attachment" "ssm_policy_ssm_cookies_attachment" {
  role       = aws_iam_role.ssm_role.name
  policy_arn = aws_iam_policy.ssm_read_cookies_policy.arn
}

resource "aws_iam_instance_profile" "ssm_instance_profile" {
  name = "ec2-ssm-instance-profile"
  role = aws_iam_role.ssm_role.name
}