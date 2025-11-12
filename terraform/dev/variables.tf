# ./terraform/dev/variables.tf

variable "aws_region" {
  description = "AWS 리전"
  type        = string
  default     = "ap-northeast-2"
}

variable "project_name" {
  description = "프로젝트 이름 태그"
  type        = string
  default     = "wealist-dev"
}

# 네트워킹
variable "vpc_cidr" {
  description = "VPC CIDR 블록"
  type        = string
  default     = "10.0.0.0/16"
}

# EC2 설정
variable "instance_type" {
  description = "EC2 인스턴스 타입"
  type        = string
  default     = "t2.micro" 
}

# RDS 설정
variable "db_instance_class" {
  description = "RDS 인스턴스 클래스 (가장 저렴한 t3.micro 권장)"
  type        = string
  default     = "db.t3.micro" 
}