provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "main" {
  cidr_block       = "10.0.0.0/16"
  instance_tenancy = "default"

  tags = {
    Name = "main"
  }
}

resource "aws_subnet" "example" {
cidr_block = "10.0.1.0/24"
  vpc_id      = aws_vpc.main.id

  tags = {
    Name = "example"
  }
}
