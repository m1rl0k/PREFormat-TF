provider "aws" {
region="us-west-2"
}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"

tags = {
Name="example_vpc"
}
}

resource "aws_subnet" "example" {
  vpc_id            = aws_vpc.example.id
  cidr_block        = "10.0.1.0/24"

  tags = {
  Name = "example_subnet"
  }
}

resource "aws_security_group" "example" {
  name        = "example"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
