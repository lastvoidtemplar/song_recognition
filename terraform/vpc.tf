resource "aws_vpc" "audio_backend_vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "audio-backend-vpc"
  }
}

resource "aws_subnet" "public_subnet_1" {
  vpc_id                  = aws_vpc.audio_backend_vpc.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = "eu-central-1a"
  map_public_ip_on_launch = true

  tags = {
    Name = "public-subnet-1"
  }
}

resource "aws_subnet" "private_subnet_1" {
  vpc_id            = aws_vpc.audio_backend_vpc.id
  cidr_block        = "10.0.3.0/24"
  availability_zone = "eu-central-1a"

  tags = {
    Name = "private-subnet-1"
  }
}

resource "aws_subnet" "private_subnet_2" {
  vpc_id            = aws_vpc.audio_backend_vpc.id
  cidr_block        = "10.0.4.0/24"
  availability_zone = "eu-central-1b"

  tags = {
    Name = "private-subnet-2"
  }
}

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.audio_backend_vpc.id

  tags = {
    Name = "audio-backend-igw"
  }
}

resource "aws_route_table" "public_route_table" {
  vpc_id = aws_vpc.audio_backend_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    Name = "public-route-table"
  }
}

resource "aws_route_table_association" "public_subnet_1_table_association" {
  subnet_id      = aws_subnet.public_subnet_1.id
  route_table_id = aws_route_table.public_route_table.id
}

resource "aws_network_acl" "public_nacl" {
  vpc_id = aws_vpc.audio_backend_vpc.id

  ingress {
    protocol = "tcp"
    rule_no = 100
    action = "allow"
    cidr_block = "0.0.0.0/0"
    from_port = 80
    to_port = 80
  }

  ingress {
    protocol = "tcp"
    rule_no = 300
    action = "allow"
    cidr_block = "0.0.0.0/0"
    from_port = 1024   
    to_port = 65535
  }


  egress {
    protocol = "all"
    rule_no = 100
    action = "allow"
    cidr_block = "0.0.0.0/0"
    from_port = 0
    to_port = 0
  }

  tags = {
    Name = "public-nacl"
  }
}

resource "aws_network_acl_association" "public_subnet_1_nacl_association" {
  subnet_id      = aws_subnet.public_subnet_1.id
  network_acl_id = aws_network_acl.public_nacl.id
}
