# ./terraform/dev/ec2.tf

# 보안 그룹 (SSH, Frontend, Backend 포트 허용)
resource "aws_security_group" "dev_sg" {
  vpc_id = aws_vpc.main.id
  name   = "${var.project_name}-app-sg"

  # SSH (22)
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] # 어디서든 접속 가능 (Dev 환경 단순화)
  }
  # Frontend (3000)
  ingress {
    from_port   = 3000
    to_port     = 3000
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  # User Service (8080)
  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  # Board Service (8000)
  ingress {
    from_port   = 8000
    to_port     = 8000
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

# EC2 인스턴스 (Docker 자동 설치)
resource "aws_instance" "app_server" {
  ami           = data.aws_ami.amazon_linux.id # Amazon Linux 2 AMI ID 사용
  instance_type = var.instance_type
  subnet_id     = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.dev_sg.id]
  key_name      = "your-ssh-key-name" # ⭐ 사용자의 SSH 키 이름으로 변경
  associate_public_ip_address = true

  # Docker 및 Docker Compose 자동 설치 스크립트
  user_data = <<-EOF
              #!/bin/bash
              sudo yum update -y
              sudo amazon-linux-extras install docker -y
              sudo service docker start
              sudo usermod -a -G docker ec2-user
              
              # Docker Compose 설치 (v2 사용)
              sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.5/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
              sudo chmod +x /usr/local/bin/docker-compose
              
              # 환경 준비 완료! (이제 GitHub Actions에서 SSH로 접속하여 'docker compose up' 실행 가능)
              EOF

  tags = {
    Name = "${var.project_name}-app-server"
  }
}

# 현재 AWS 리전에서 Amazon Linux 2 AMI ID를 동적으로 가져옴
data "aws_ami" "amazon_linux" {
  most_recent = true
  owners = ["amazon"]
  filter {
    name = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}