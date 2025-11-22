package OrangeCloud.UserRepo.service;

import OrangeCloud.UserRepo.config.S3Config;
import OrangeCloud.UserRepo.exception.CustomException;
import OrangeCloud.UserRepo.exception.ErrorCode;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import software.amazon.awssdk.services.s3.model.PutObjectRequest;
import software.amazon.awssdk.services.s3.presigner.S3Presigner;
import software.amazon.awssdk.services.s3.presigner.model.PresignedPutObjectRequest;
import software.amazon.awssdk.services.s3.presigner.model.PutObjectPresignRequest;

import java.time.Duration;
import java.time.LocalDateTime;
import java.util.UUID;

/**
 * S3 서비스
 * Presigned URL 생성 및 파일 키 관리를 담당합니다.
 */
@Service
public class S3Service {

    private static final Logger logger = LoggerFactory.getLogger(S3Service.class);
    private static final Duration PRESIGNED_URL_EXPIRATION = Duration.ofMinutes(5);

    private final S3Presigner s3Presigner;
    private final S3Config s3Config;

    public S3Service(S3Presigner s3Presigner, S3Config s3Config) {
        this.s3Presigner = s3Presigner;
        this.s3Config = s3Config;
    }

    /**
     * Presigned URL 생성
     *
     * @param workspaceId 워크스페이스 ID
     * @param userId      사용자 ID
     * @param fileName    파일명
     * @param contentType Content-Type
     * @return Presigned URL과 파일 키를 포함한 응답
     */
    public PresignedUrlResponse generatePresignedUrl(
            UUID workspaceId,
            UUID userId,
            String fileName,
            String contentType) {

        // 파라미터 검증
        validateParameters(workspaceId, userId, fileName, contentType);

        try {
            // 파일 키 생성
            String fileKey = generateFileKey(workspaceId, userId, fileName);

            // PutObjectRequest 생성
            PutObjectRequest putObjectRequest = PutObjectRequest.builder()
                    .bucket(s3Config.getBucket())
                    .key(fileKey)
                    .contentType(contentType)
                    .build();

            // Presigned URL 생성
            PutObjectPresignRequest presignRequest = PutObjectPresignRequest.builder()
                    .signatureDuration(PRESIGNED_URL_EXPIRATION)
                    .putObjectRequest(putObjectRequest)
                    .build();

            PresignedPutObjectRequest presignedRequest = s3Presigner.presignPutObject(presignRequest);
            String presignedUrl = presignedRequest.url().toString();

            logger.info("Presigned URL 생성 성공 - workspaceId: {}, userId: {}, fileKey: {}",
                    workspaceId, userId, fileKey);

            return new PresignedUrlResponse(presignedUrl, fileKey, (int) PRESIGNED_URL_EXPIRATION.getSeconds());

        } catch (Exception e) {
            logger.error("Presigned URL 생성 실패 - workspaceId: {}, userId: {}, error: {}",
                    workspaceId, userId, e.getMessage(), e);
            throw new CustomException(ErrorCode.S3_UPLOAD_FAILED, "Presigned URL 생성에 실패했습니다.");
        }
    }

    /**
     * 파일 키 생성
     * 형식: user/{workspaceId}/{year}/{month}/{userId}_{timestamp}.ext
     *
     * @param workspaceId 워크스페이스 ID
     * @param userId      사용자 ID
     * @param fileName    파일명
     * @return 생성된 파일 키
     */
    private String generateFileKey(UUID workspaceId, UUID userId, String fileName) {
        LocalDateTime now = LocalDateTime.now();
        String year = String.valueOf(now.getYear());
        String month = String.format("%02d", now.getMonthValue());
        long timestamp = System.currentTimeMillis();

        // 파일 확장자 추출
        String extension = "";
        int lastDotIndex = fileName.lastIndexOf('.');
        if (lastDotIndex > 0 && lastDotIndex < fileName.length() - 1) {
            extension = fileName.substring(lastDotIndex);
        }

        return String.format("user/%s/%s/%s/%s_%d%s",
                workspaceId, year, month, userId, timestamp, extension);
    }

    /**
     * 파일 키로부터 S3 URL 생성
     *
     * @param fileKey S3 파일 키
     * @return S3 파일 URL
     */
    public String generateS3Url(String fileKey) {
        if (fileKey == null || fileKey.trim().isEmpty()) {
            throw new CustomException(ErrorCode.INVALID_INPUT_VALUE, "파일 키는 필수입니다.");
        }

        // fileKey 형식 검증 (user/ 로 시작해야 함)
        if (!fileKey.startsWith("user/")) {
            logger.warn("Invalid fileKey format: {}", fileKey);
            throw new CustomException(ErrorCode.INVALID_INPUT_VALUE, "잘못된 파일 키 형식입니다.");
        }

        // S3 URL 생성
        // 형식: https://{bucket}.s3.{region}.amazonaws.com/{fileKey}
        String s3Url = String.format("https://%s.s3.%s.amazonaws.com/%s",
                s3Config.getBucket(),
                s3Config.getRegion(),
                fileKey);

        logger.debug("Generated S3 URL from fileKey: {} -> {}", fileKey, s3Url);
        return s3Url;
    }

    /**
     * 파라미터 검증
     */
    private void validateParameters(UUID workspaceId, UUID userId, String fileName, String contentType) {
        if (workspaceId == null) {
            throw new CustomException(ErrorCode.INVALID_INPUT_VALUE, "워크스페이스 ID는 필수입니다.");
        }
        if (userId == null) {
            throw new CustomException(ErrorCode.INVALID_INPUT_VALUE, "사용자 ID는 필수입니다.");
        }
        if (fileName == null || fileName.trim().isEmpty()) {
            throw new CustomException(ErrorCode.INVALID_INPUT_VALUE, "파일명은 필수입니다.");
        }
        if (contentType == null || contentType.trim().isEmpty()) {
            throw new CustomException(ErrorCode.INVALID_INPUT_VALUE, "Content-Type은 필수입니다.");
        }
    }

    /**
     * Presigned URL 응답 DTO
     */
    public static class PresignedUrlResponse {
        private final String uploadUrl;
        private final String fileKey;
        private final int expiresIn;

        public PresignedUrlResponse(String uploadUrl, String fileKey, int expiresIn) {
            this.uploadUrl = uploadUrl;
            this.fileKey = fileKey;
            this.expiresIn = expiresIn;
        }

        public String getUploadUrl() {
            return uploadUrl;
        }

        public String getFileKey() {
            return fileKey;
        }

        public int getExpiresIn() {
            return expiresIn;
        }
    }
}
