# ./terraform/dev/outputs.tf

# 1. EC2 Public IP (SSH 접속 및 백엔드 API 접속용)
output "ec2_public_ip" {
  description = "Application Server Public IP Address"
  value       = aws_instance.app_server.public_ip
}

# 2. CloudFront 도메인 이름 (Frontend 접속용)
output "cloudfront_domain_name" {
  description = "Frontend CDN URL"
  value       = aws_cloudfront_distribution.frontend_cdn.domain_name
}

# 3. Frontend S3 버킷 이름 (Frontend CI/CD Deploy Target)
output "frontend_s3_bucket_name" {
  description = "Frontend S3 Bucket Name"
  value       = aws_s3_bucket.frontend_bucket.bucket
}

# 4. CloudFront Distribution ID (Frontend Cache Invalidation용)
output "cloudfront_distribution_id" {
  description = "CloudFront Distribution ID for cache invalidation"
  value       = aws_cloudfront_distribution.frontend_cdn.id
}