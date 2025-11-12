# ./terraform/dev/rds.tf

# RDS용 보안 그룹 (EC2에서만 5432 포트 접근 허용)
resource "aws_security_group" "rds_sg" {
  vpc_id = aws_vpc.main.id
  name   = "${var.project_name}-rds-sg"

  ingress {
    from_port   = 5432 # PostgreSQL 포트
    to_port     = 5432
    protocol    = "tcp"
    # 소스: EC2 보안 그룹만 허용 (EC2에서만 접속 가능)
    security_groups = [aws_security_group.dev_sg.id] 
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# 1. RDS 마스터 비밀번호 생성 (임시)
resource "random_password" "db_password" {
  length  = 16
  special = false
  upper   = true
  lower   = true
  number  = true
}

# 2. RDS 인스턴스 생성
resource "aws_db_instance" "postgres" {
  identifier           = "${var.project_name}-postgres"
  allocated_storage    = 20
  engine               = "postgres"
  engine_version       = "15.4"
  instance_class       = var.db_instance_class
  username             = "wealist_admin"
  password             = random_password.db_password.result # 임시 비밀번호 사용
  skip_final_snapshot  = true
  db_subnet_group_name = aws_db_subnet_group.default.name
  vpc_security_group_ids = [aws_security_group.rds_sg.id]
}

# RDS Subnet Group (RDS가 Subnet을 선택하도록)
resource "aws_db_subnet_group" "default" {
  name       = "${var.project_name}-rds-sng"
  subnet_ids = [aws_subnet.public.id] 
  tags = {
    Name = "RDS Subnet Group"
  }
}