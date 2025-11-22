package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.userprofile.PresignedUrlRequest;
import OrangeCloud.UserRepo.dto.userprofile.PresignedUrlResponse;
import OrangeCloud.UserRepo.exception.CustomException;
import OrangeCloud.UserRepo.exception.ErrorCode;
import OrangeCloud.UserRepo.service.S3Service;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.security.Principal;
import java.util.Set;
import java.util.UUID;

/**
 * 프로필 이미지 컨트롤러
 * Presigned URL 기반 프로필 이미지 업로드를 처리합니다.
 */
@RestController
@RequestMapping("/api/profiles/me/image")
@RequiredArgsConstructor
@Tag(name = "ProfileImage", description = "프로필 이미지 업로드 API")
@Slf4j
public class ProfileImageController {

    private static final long MAX_FILE_SIZE = 20 * 1024 * 1024; // 20MB
    private static final Set<String> ALLOWED_IMAGE_TYPES = Set.of(
            "image/jpeg",
            "image/png",
            "image/gif",
            "image/webp"
    );
    private static final Set<String> ALLOWED_IMAGE_EXTENSIONS = Set.of(
            ".jpg", ".jpeg", ".png", ".gif", ".webp"
    );

    private final S3Service s3Service;

    /**
     * 인증된 사용자 ID 추출
     */
    private UUID extractUserId(Principal principal) {
        if (principal instanceof Authentication authentication) {
            return UUID.fromString(authentication.getName());
        }
        throw new IllegalStateException("인증된 사용자 정보를 찾을 수 없습니다.");
    }

    /**
     * Presigned URL 생성
     *
     * @param request   Presigned URL 요청
     * @param principal 인증된 사용자 정보
     * @return Presigned URL 응답
     */
    @PostMapping("/presigned-url")
    @Operation(
            summary = "프로필 이미지 업로드를 위한 Presigned URL 생성",
            description = "클라이언트가 S3에 직접 업로드할 수 있는 Presigned URL을 생성합니다. " +
                    "이미지 파일만 허용되며, 최대 20MB까지 업로드 가능합니다."
    )
    public ResponseEntity<PresignedUrlResponse> generatePresignedUrl(
            @Valid @RequestBody PresignedUrlRequest request,
            Principal principal) {

        UUID userId = extractUserId(principal);
        log.info("Presigned URL 요청 - userId: {}, workspaceId: {}, fileName: {}, fileSize: {}, contentType: {}",
                userId, request.workspaceId(), request.fileName(), request.fileSize(), request.contentType());

        // 파일 메타데이터 검증
        validateFileMetadata(request);

        // Presigned URL 생성
        S3Service.PresignedUrlResponse s3Response = s3Service.generatePresignedUrl(
                request.workspaceId(),
                userId,
                request.fileName(),
                request.contentType()
        );

        PresignedUrlResponse response = new PresignedUrlResponse(
                s3Response.getUploadUrl(),
                s3Response.getFileKey(),
                s3Response.getExpiresIn()
        );

        log.info("Presigned URL 생성 완료 - userId: {}, fileKey: {}", userId, response.fileKey());
        return ResponseEntity.ok(response);
    }

    /**
     * 파일 메타데이터 검증
     * - 파일 크기: 20MB 이하
     * - 파일 타입: 이미지만 허용 (jpg, jpeg, png, gif, webp)
     * - 파일 확장자: Content-Type과 일치
     */
    private void validateFileMetadata(PresignedUrlRequest request) {
        // 파일 크기 검증
        if (request.fileSize() > MAX_FILE_SIZE) {
            log.warn("파일 크기 초과 - fileName: {}, fileSize: {}, maxSize: {}",
                    request.fileName(), request.fileSize(), MAX_FILE_SIZE);
            throw new CustomException(ErrorCode.FILE_TOO_LARGE,
                    String.format("파일 크기는 %dMB를 초과할 수 없습니다.", MAX_FILE_SIZE / (1024 * 1024)));
        }

        // Content-Type 검증 (이미지만 허용)
        if (!ALLOWED_IMAGE_TYPES.contains(request.contentType().toLowerCase())) {
            log.warn("지원하지 않는 파일 타입 - fileName: {}, contentType: {}",
                    request.fileName(), request.contentType());
            throw new CustomException(ErrorCode.INVALID_FILE_TYPE,
                    "이미지 파일만 업로드 가능합니다. 지원 형식: jpg, jpeg, png, gif, webp");
        }

        // 파일 확장자 검증
        String fileName = request.fileName().toLowerCase();
        boolean hasValidExtension = ALLOWED_IMAGE_EXTENSIONS.stream()
                .anyMatch(fileName::endsWith);

        if (!hasValidExtension) {
            log.warn("지원하지 않는 파일 확장자 - fileName: {}", request.fileName());
            throw new CustomException(ErrorCode.INVALID_FILE_TYPE,
                    "지원하지 않는 파일 확장자입니다. 지원 형식: jpg, jpeg, png, gif, webp");
        }

        log.debug("파일 메타데이터 검증 완료 - fileName: {}, fileSize: {}, contentType: {}",
                request.fileName(), request.fileSize(), request.contentType());
    }
}
