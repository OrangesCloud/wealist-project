package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.userprofile.PresignedUrlRequest;
import OrangeCloud.UserRepo.exception.CustomException;
import OrangeCloud.UserRepo.exception.ErrorCode;
import OrangeCloud.UserRepo.service.S3Service;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.MediaType;
import org.springframework.security.core.Authentication;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;

import java.security.Principal;
import java.util.UUID;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.*;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

/**
 * ProfileImageController 단위 테스트
 * Requirements: 2.1, 2.2
 */
@ExtendWith(MockitoExtension.class)
class ProfileImageControllerTest {

    @Mock
    private S3Service s3Service;

    @InjectMocks
    private ProfileImageController profileImageController;

    private MockMvc mockMvc;
    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        mockMvc = MockMvcBuilders.standaloneSetup(profileImageController).build();
        objectMapper = new ObjectMapper();
    }

    private Principal createMockPrincipal(UUID userId) {
        Authentication authentication = mock(Authentication.class);
        when(authentication.getName()).thenReturn(userId.toString());
        return authentication;
    }

    @Test
    @DisplayName("유효한 요청으로 Presigned URL 생성 성공")
    void generatePresignedUrl_Success() throws Exception {
        // Given
        UUID userId = UUID.randomUUID();
        UUID workspaceId = UUID.randomUUID();
        PresignedUrlRequest request = new PresignedUrlRequest(
                workspaceId,
                "avatar.jpg",
                512000L,
                "image/jpeg"
        );

        S3Service.PresignedUrlResponse s3Response = new S3Service.PresignedUrlResponse(
                "https://test-bucket.s3.amazonaws.com/user/test/2024/01/user_123456789.jpg?signature=xyz",
                "user/" + workspaceId + "/2024/01/" + userId + "_123456789.jpg",
                300
        );

        when(s3Service.generatePresignedUrl(eq(workspaceId), eq(userId), eq("avatar.jpg"), eq("image/jpeg")))
                .thenReturn(s3Response);

        Principal principal = createMockPrincipal(userId);

        // When & Then
        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .principal(principal)
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.uploadUrl").value(s3Response.getUploadUrl()))
                .andExpect(jsonPath("$.fileKey").value(s3Response.getFileKey()))
                .andExpect(jsonPath("$.expiresIn").value(300));

        verify(s3Service, times(1)).generatePresignedUrl(
                eq(workspaceId), eq(userId), eq("avatar.jpg"), eq("image/jpeg"));
    }

    @Test
    @DisplayName("파일 크기 초과 시 400 에러 반환")
    void generatePresignedUrl_FileTooLarge_Returns400() throws Exception {
        // Given
        UUID userId = UUID.randomUUID();
        UUID workspaceId = UUID.randomUUID();
        long fileSizeOver20MB = 21 * 1024 * 1024L; // 21MB

        PresignedUrlRequest request = new PresignedUrlRequest(
                workspaceId,
                "large-image.jpg",
                fileSizeOver20MB,
                "image/jpeg"
        );

        Principal principal = createMockPrincipal(userId);

        // When & Then
        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .principal(principal)
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isBadRequest());

        // S3Service가 호출되지 않았는지 검증
        verify(s3Service, never()).generatePresignedUrl(any(), any(), any(), any());
    }

    @Test
    @DisplayName("잘못된 이미지 타입 - application/pdf - 400 에러 반환")
    void generatePresignedUrl_InvalidImageType_Pdf_Returns400() throws Exception {
        // Given
        UUID userId = UUID.randomUUID();
        UUID workspaceId = UUID.randomUUID();

        PresignedUrlRequest request = new PresignedUrlRequest(
                workspaceId,
                "document.pdf",
                512000L,
                "application/pdf"
        );

        Principal principal = createMockPrincipal(userId);

        // When & Then
        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .principal(principal)
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isBadRequest());

        verify(s3Service, never()).generatePresignedUrl(any(), any(), any(), any());
    }

    @Test
    @DisplayName("잘못된 이미지 타입 - video/mp4 - 400 에러 반환")
    void generatePresignedUrl_InvalidImageType_Video_Returns400() throws Exception {
        // Given
        UUID userId = UUID.randomUUID();
        UUID workspaceId = UUID.randomUUID();

        PresignedUrlRequest request = new PresignedUrlRequest(
                workspaceId,
                "video.mp4",
                512000L,
                "video/mp4"
        );

        Principal principal = createMockPrincipal(userId);

        // When & Then
        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .principal(principal)
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isBadRequest());

        verify(s3Service, never()).generatePresignedUrl(any(), any(), any(), any());
    }

    @Test
    @DisplayName("잘못된 파일 확장자 - .txt - 400 에러 반환")
    void generatePresignedUrl_InvalidFileExtension_Returns400() throws Exception {
        // Given
        UUID userId = UUID.randomUUID();
        UUID workspaceId = UUID.randomUUID();

        PresignedUrlRequest request = new PresignedUrlRequest(
                workspaceId,
                "file.txt",
                512000L,
                "image/jpeg" // Content-Type은 이미지지만 확장자가 다름
        );

        Principal principal = createMockPrincipal(userId);

        // When & Then
        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .principal(principal)
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isBadRequest());

        verify(s3Service, never()).generatePresignedUrl(any(), any(), any(), any());
    }

    @Test
    @DisplayName("허용된 모든 이미지 타입 검증 - jpg, jpeg, png, gif, webp")
    void generatePresignedUrl_AllAllowedImageTypes_Success() throws Exception {
        // Given
        UUID userId = UUID.randomUUID();
        UUID workspaceId = UUID.randomUUID();

        String[][] testCases = {
                {"image.jpg", "image/jpeg"},
                {"image.jpeg", "image/jpeg"},
                {"image.png", "image/png"},
                {"image.gif", "image/gif"},
                {"image.webp", "image/webp"}
        };

        for (String[] testCase : testCases) {
            String fileName = testCase[0];
            String contentType = testCase[1];

            PresignedUrlRequest request = new PresignedUrlRequest(
                    workspaceId,
                    fileName,
                    512000L,
                    contentType
            );

            S3Service.PresignedUrlResponse s3Response = new S3Service.PresignedUrlResponse(
                    "https://test-bucket.s3.amazonaws.com/test",
                    "user/test/2024/01/test.jpg",
                    300
            );

            when(s3Service.generatePresignedUrl(any(), any(), any(), any()))
                    .thenReturn(s3Response);

            Principal principal = createMockPrincipal(userId);

            // When & Then
            mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                            .principal(principal)
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isOk());
        }

        // 5개의 이미지 타입 모두 테스트되었는지 검증
        verify(s3Service, times(5)).generatePresignedUrl(any(), any(), any(), any());
    }

    @Test
    @DisplayName("정확히 20MB 파일은 허용됨")
    void generatePresignedUrl_Exactly20MB_Success() throws Exception {
        // Given
        UUID userId = UUID.randomUUID();
        UUID workspaceId = UUID.randomUUID();
        long exactly20MB = 20 * 1024 * 1024L;

        PresignedUrlRequest request = new PresignedUrlRequest(
                workspaceId,
                "large-image.jpg",
                exactly20MB,
                "image/jpeg"
        );

        S3Service.PresignedUrlResponse s3Response = new S3Service.PresignedUrlResponse(
                "https://test-bucket.s3.amazonaws.com/test",
                "user/test/2024/01/test.jpg",
                300
        );

        when(s3Service.generatePresignedUrl(any(), any(), any(), any()))
                .thenReturn(s3Response);

        Principal principal = createMockPrincipal(userId);

        // When & Then
        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .principal(principal)
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isOk());

        verify(s3Service, times(1)).generatePresignedUrl(any(), any(), any(), any());
    }

    @Test
    @DisplayName("필수 필드 누락 시 400 에러 반환 - workspaceId")
    void generatePresignedUrl_MissingWorkspaceId_Returns400() throws Exception {
        // Given
        UUID userId = UUID.randomUUID();
        String requestJson = """
                {
                    "fileName": "avatar.jpg",
                    "fileSize": 512000,
                    "contentType": "image/jpeg"
                }
                """;

        Principal principal = createMockPrincipal(userId);

        // When & Then
        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .principal(principal)
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson))
                .andExpect(status().isBadRequest());

        verify(s3Service, never()).generatePresignedUrl(any(), any(), any(), any());
    }

    @Test
    @DisplayName("필수 필드 누락 시 400 에러 반환 - fileName")
    void generatePresignedUrl_MissingFileName_Returns400() throws Exception {
        // Given
        UUID userId = UUID.randomUUID();
        UUID workspaceId = UUID.randomUUID();
        String requestJson = String.format("""
                {
                    "workspaceId": "%s",
                    "fileSize": 512000,
                    "contentType": "image/jpeg"
                }
                """, workspaceId);

        Principal principal = createMockPrincipal(userId);

        // When & Then
        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .principal(principal)
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson))
                .andExpect(status().isBadRequest());

        verify(s3Service, never()).generatePresignedUrl(any(), any(), any(), any());
    }

    @Test
    @DisplayName("필수 필드 누락 시 400 에러 반환 - fileSize")
    void generatePresignedUrl_MissingFileSize_Returns400() throws Exception {
        // Given
        UUID userId = UUID.randomUUID();
        UUID workspaceId = UUID.randomUUID();
        String requestJson = String.format("""
                {
                    "workspaceId": "%s",
                    "fileName": "avatar.jpg",
                    "contentType": "image/jpeg"
                }
                """, workspaceId);

        Principal principal = createMockPrincipal(userId);

        // When & Then
        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .principal(principal)
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson))
                .andExpect(status().isBadRequest());

        verify(s3Service, never()).generatePresignedUrl(any(), any(), any(), any());
    }

    @Test
    @DisplayName("필수 필드 누락 시 400 에러 반환 - contentType")
    void generatePresignedUrl_MissingContentType_Returns400() throws Exception {
        // Given
        UUID userId = UUID.randomUUID();
        UUID workspaceId = UUID.randomUUID();
        String requestJson = String.format("""
                {
                    "workspaceId": "%s",
                    "fileName": "avatar.jpg",
                    "fileSize": 512000
                }
                """, workspaceId);

        Principal principal = createMockPrincipal(userId);

        // When & Then
        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .principal(principal)
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson))
                .andExpect(status().isBadRequest());

        verify(s3Service, never()).generatePresignedUrl(any(), any(), any(), any());
    }

    @Test
    @DisplayName("음수 파일 크기는 거부됨")
    void generatePresignedUrl_NegativeFileSize_Returns400() throws Exception {
        // Given
        UUID userId = UUID.randomUUID();
        UUID workspaceId = UUID.randomUUID();
        String requestJson = String.format("""
                {
                    "workspaceId": "%s",
                    "fileName": "avatar.jpg",
                    "fileSize": -1000,
                    "contentType": "image/jpeg"
                }
                """, workspaceId);

        Principal principal = createMockPrincipal(userId);

        // When & Then
        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .principal(principal)
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson))
                .andExpect(status().isBadRequest());

        verify(s3Service, never()).generatePresignedUrl(any(), any(), any(), any());
    }

    @Test
    @DisplayName("대소문자 구분 없이 Content-Type 검증")
    void generatePresignedUrl_CaseInsensitiveContentType_Success() throws Exception {
        // Given
        UUID userId = UUID.randomUUID();
        UUID workspaceId = UUID.randomUUID();

        PresignedUrlRequest request = new PresignedUrlRequest(
                workspaceId,
                "avatar.jpg",
                512000L,
                "IMAGE/JPEG" // 대문자
        );

        S3Service.PresignedUrlResponse s3Response = new S3Service.PresignedUrlResponse(
                "https://test-bucket.s3.amazonaws.com/test",
                "user/test/2024/01/test.jpg",
                300
        );

        when(s3Service.generatePresignedUrl(any(), any(), any(), any()))
                .thenReturn(s3Response);

        Principal principal = createMockPrincipal(userId);

        // When & Then
        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .principal(principal)
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isOk());

        verify(s3Service, times(1)).generatePresignedUrl(any(), any(), any(), any());
    }
}
