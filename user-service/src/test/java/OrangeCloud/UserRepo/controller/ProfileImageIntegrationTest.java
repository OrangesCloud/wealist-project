package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.userprofile.PresignedUrlRequest;
import OrangeCloud.UserRepo.dto.userprofile.UpdateProfileImageByKeyRequest;
import OrangeCloud.UserRepo.dto.userprofile.UserProfileResponse;
import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.entity.UserProfile;
import OrangeCloud.UserRepo.entity.Workspace;
import OrangeCloud.UserRepo.repository.UserProfileRepository;
import OrangeCloud.UserRepo.repository.UserRepository;
import OrangeCloud.UserRepo.repository.WorkspaceRepository;
import OrangeCloud.UserRepo.service.S3Service;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.security.test.context.support.WithMockUser;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.put;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

/**
 * ProfileImage 통합 테스트 - Presigned URL 플로우
 * **Validates: Requirements 2.3, 2.4, 2.5**
 */
@SpringBootTest(classes = {OrangeCloud.UserRepo.UserRepoApplication.class, OrangeCloud.UserRepo.config.TestS3Config.class})
@AutoConfigureMockMvc
@ActiveProfiles("test")
@Transactional
class ProfileImageIntegrationTest {

    @Autowired
    private MockMvc mockMvc;

    @Autowired
    private ObjectMapper objectMapper;

    @Autowired
    private UserRepository userRepository;

    @Autowired
    private WorkspaceRepository workspaceRepository;

    @Autowired
    private UserProfileRepository userProfileRepository;

    @Autowired
    private S3Service s3Service;

    private User testUser;
    private Workspace testWorkspace;
    private UserProfile testProfile;

    @BeforeEach
    void setUp() {
        // Create test user first
        testUser = User.builder()
                .email("test@example.com")
                .provider("google")
                .googleId("test-google-id")
                .isActive(true)
                .build();
        testUser = userRepository.save(testUser);

        // Create test workspace
        testWorkspace = Workspace.builder()
                .ownerId(testUser.getUserId())
                .workspaceName("Test Workspace")
                .workspaceDescription("Test Description")
                .isPublic(true)
                .needApproved(false)
                .isActive(true)
                .build();
        testWorkspace = workspaceRepository.save(testWorkspace);

        // Create test profile
        testProfile = UserProfile.builder()
                .userId(testUser.getUserId())
                .workspaceId(testWorkspace.getWorkspaceId())
                .nickName("Test Nickname")
                .email(testUser.getEmail())
                .build();
        testProfile = userProfileRepository.save(testProfile);
    }

    @Test
    @DisplayName("통합 테스트: Presigned URL 생성 → 프로필 업데이트 전체 플로우")
    @WithMockUser(username = "test-user-id")
    void testCompletePresignedUrlFlow() throws Exception {
        // Step 1: Request presigned URL
        PresignedUrlRequest presignedRequest = new PresignedUrlRequest(
                testWorkspace.getWorkspaceId(),
                "avatar.jpg",
                512000L,
                "image/jpeg"
        );

        String presignedResponse = mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(presignedRequest)))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.uploadUrl").exists())
                .andExpect(jsonPath("$.fileKey").exists())
                .andExpect(jsonPath("$.expiresIn").value(300))
                .andReturn()
                .getResponse()
                .getContentAsString();

        // Parse presigned URL response
        S3Service.PresignedUrlResponse presignedUrlResponse = objectMapper.readValue(
                presignedResponse,
                S3Service.PresignedUrlResponse.class
        );

        assertThat(presignedUrlResponse.getUploadUrl()).isNotEmpty();
        assertThat(presignedUrlResponse.getFileKey()).isNotEmpty();
        assertThat(presignedUrlResponse.getFileKey()).startsWith("user/" + testWorkspace.getWorkspaceId());
        assertThat(presignedUrlResponse.getExpiresIn()).isEqualTo(300);

        // Step 2: Simulate client uploading to S3 (skipped in test)
        // In real scenario, client would PUT file to presignedUrlResponse.getUploadUrl()

        // Step 3: Update profile with fileKey
        UpdateProfileImageByKeyRequest updateRequest = new UpdateProfileImageByKeyRequest(
                testWorkspace.getWorkspaceId(),
                presignedUrlResponse.getFileKey()
        );

        String updateResponse = mockMvc.perform(put("/api/profiles/me/image")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(updateRequest)))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.profileImageUrl").exists())
                .andExpect(jsonPath("$.userId").value(testUser.getUserId().toString()))
                .andExpect(jsonPath("$.workspaceId").value(testWorkspace.getWorkspaceId().toString()))
                .andReturn()
                .getResponse()
                .getContentAsString();

        // Parse update response
        UserProfileResponse profileResponse = objectMapper.readValue(
                updateResponse,
                UserProfileResponse.class
        );

        assertThat(profileResponse.getProfileImageUrl()).isNotEmpty();
        assertThat(profileResponse.getProfileImageUrl()).contains(presignedUrlResponse.getFileKey());

        // Step 4: Verify profile was updated in database
        UserProfile updatedProfile = userProfileRepository.findById(testProfile.getProfileId()).orElseThrow();
        assertThat(updatedProfile.getProfileImageUrl()).isNotNull();
        assertThat(updatedProfile.getProfileImageUrl()).contains(presignedUrlResponse.getFileKey());
    }

    @Test
    @DisplayName("통합 테스트: 여러 이미지 타입으로 Presigned URL 생성")
    @WithMockUser(username = "test-user-id")
    void testPresignedUrlWithMultipleImageTypes() throws Exception {
        String[][] testCases = {
                {"photo.jpg", "image/jpeg"},
                {"screenshot.png", "image/png"},
                {"animation.gif", "image/gif"},
                {"modern-image.webp", "image/webp"}
        };

        for (String[] testCase : testCases) {
            String fileName = testCase[0];
            String contentType = testCase[1];

            PresignedUrlRequest request = new PresignedUrlRequest(
                    testWorkspace.getWorkspaceId(),
                    fileName,
                    512000L,
                    contentType
            );

            mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isOk())
                    .andExpect(jsonPath("$.uploadUrl").exists())
                    .andExpect(jsonPath("$.fileKey").exists())
                    .andExpect(jsonPath("$.fileKey").value(org.hamcrest.Matchers.containsString("user/" + testWorkspace.getWorkspaceId())));
        }
    }

    @Test
    @DisplayName("통합 테스트: 파일 크기 검증 - 20MB 제한")
    @WithMockUser(username = "test-user-id")
    void testFileSizeValidation() throws Exception {
        // Test exactly 20MB (should succeed)
        PresignedUrlRequest validRequest = new PresignedUrlRequest(
                testWorkspace.getWorkspaceId(),
                "large-image.jpg",
                20 * 1024 * 1024L,
                "image/jpeg"
        );

        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(validRequest)))
                .andExpect(status().isOk());

        // Test 21MB (should fail)
        PresignedUrlRequest invalidRequest = new PresignedUrlRequest(
                testWorkspace.getWorkspaceId(),
                "too-large.jpg",
                21 * 1024 * 1024L,
                "image/jpeg"
        );

        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(invalidRequest)))
                .andExpect(status().isBadRequest());
    }

    @Test
    @DisplayName("통합 테스트: 잘못된 파일 타입 거부")
    @WithMockUser(username = "test-user-id")
    void testInvalidFileTypeRejection() throws Exception {
        String[][] invalidTestCases = {
                {"document.pdf", "application/pdf"},
                {"video.mp4", "video/mp4"},
                {"audio.mp3", "audio/mpeg"},
                {"text.txt", "text/plain"}
        };

        for (String[] testCase : invalidTestCases) {
            String fileName = testCase[0];
            String contentType = testCase[1];

            PresignedUrlRequest request = new PresignedUrlRequest(
                    testWorkspace.getWorkspaceId(),
                    fileName,
                    512000L,
                    contentType
            );

            mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isBadRequest());
        }
    }

    @Test
    @DisplayName("통합 테스트: 잘못된 fileKey로 프로필 업데이트 실패")
    @WithMockUser(username = "test-user-id")
    void testUpdateProfileWithInvalidFileKey() throws Exception {
        // Test with board service file key (should fail)
        UpdateProfileImageByKeyRequest invalidRequest = new UpdateProfileImageByKeyRequest(
                testWorkspace.getWorkspaceId(),
                "board/boards/" + testWorkspace.getWorkspaceId() + "/2024/01/test.jpg"
        );

        mockMvc.perform(put("/api/profiles/me/image")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(invalidRequest)))
                .andExpect(status().isBadRequest());

        // Verify profile was not updated
        UserProfile unchangedProfile = userProfileRepository.findById(testProfile.getProfileId()).orElseThrow();
        assertThat(unchangedProfile.getProfileImageUrl()).isNull();
    }

    @Test
    @DisplayName("통합 테스트: 빈 fileKey로 프로필 업데이트 실패")
    @WithMockUser(username = "test-user-id")
    void testUpdateProfileWithEmptyFileKey() throws Exception {
        UpdateProfileImageByKeyRequest emptyRequest = new UpdateProfileImageByKeyRequest(
                testWorkspace.getWorkspaceId(),
                ""
        );

        mockMvc.perform(put("/api/profiles/me/image")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(emptyRequest)))
                .andExpect(status().isBadRequest());
    }

    @Test
    @DisplayName("통합 테스트: 여러 번 프로필 이미지 업데이트")
    @WithMockUser(username = "test-user-id")
    void testMultipleProfileImageUpdates() throws Exception {
        // First update
        PresignedUrlRequest request1 = new PresignedUrlRequest(
                testWorkspace.getWorkspaceId(),
                "avatar1.jpg",
                512000L,
                "image/jpeg"
        );

        String response1 = mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request1)))
                .andExpect(status().isOk())
                .andReturn()
                .getResponse()
                .getContentAsString();

        S3Service.PresignedUrlResponse presignedResponse1 = objectMapper.readValue(
                response1,
                S3Service.PresignedUrlResponse.class
        );

        UpdateProfileImageByKeyRequest updateRequest1 = new UpdateProfileImageByKeyRequest(
                testWorkspace.getWorkspaceId(),
                presignedResponse1.getFileKey()
        );

        mockMvc.perform(put("/api/profiles/me/image")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(updateRequest1)))
                .andExpect(status().isOk());

        // Second update (should replace first)
        PresignedUrlRequest request2 = new PresignedUrlRequest(
                testWorkspace.getWorkspaceId(),
                "avatar2.png",
                256000L,
                "image/png"
        );

        String response2 = mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request2)))
                .andExpect(status().isOk())
                .andReturn()
                .getResponse()
                .getContentAsString();

        S3Service.PresignedUrlResponse presignedResponse2 = objectMapper.readValue(
                response2,
                S3Service.PresignedUrlResponse.class
        );

        UpdateProfileImageByKeyRequest updateRequest2 = new UpdateProfileImageByKeyRequest(
                testWorkspace.getWorkspaceId(),
                presignedResponse2.getFileKey()
        );

        mockMvc.perform(put("/api/profiles/me/image")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(updateRequest2)))
                .andExpect(status().isOk());

        // Verify final profile has second image
        UserProfile finalProfile = userProfileRepository.findById(testProfile.getProfileId()).orElseThrow();
        assertThat(finalProfile.getProfileImageUrl()).contains(presignedResponse2.getFileKey());
        assertThat(finalProfile.getProfileImageUrl()).doesNotContain(presignedResponse1.getFileKey());
    }

    @Test
    @DisplayName("통합 테스트: 파일명에 특수 문자 포함")
    @WithMockUser(username = "test-user-id")
    void testFileNameWithSpecialCharacters() throws Exception {
        String[] fileNames = {
                "my-avatar.jpg",
                "user_photo.png",
                "profile (1).jpg",
                "image-2024.webp"
        };

        for (String fileName : fileNames) {
            PresignedUrlRequest request = new PresignedUrlRequest(
                    testWorkspace.getWorkspaceId(),
                    fileName,
                    512000L,
                    "image/jpeg"
            );

            mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(objectMapper.writeValueAsString(request)))
                    .andExpect(status().isOk())
                    .andExpect(jsonPath("$.uploadUrl").exists())
                    .andExpect(jsonPath("$.fileKey").exists());
        }
    }

    @Test
    @DisplayName("통합 테스트: 동시에 여러 워크스페이스에서 프로필 이미지 업데이트")
    @WithMockUser(username = "test-user-id")
    void testMultipleWorkspaceProfileUpdates() throws Exception {
        // Create second workspace and profile
        Workspace workspace2 = Workspace.builder()
                .ownerId(testUser.getUserId())
                .workspaceName("Second Workspace")
                .workspaceDescription("Second Description")
                .isPublic(true)
                .needApproved(false)
                .isActive(true)
                .build();
        workspace2 = workspaceRepository.save(workspace2);

        UserProfile profile2 = UserProfile.builder()
                .userId(testUser.getUserId())
                .workspaceId(workspace2.getWorkspaceId())
                .nickName("Second Nickname")
                .email(testUser.getEmail())
                .build();
        profile2 = userProfileRepository.save(profile2);

        // Update first workspace profile
        PresignedUrlRequest request1 = new PresignedUrlRequest(
                testWorkspace.getWorkspaceId(),
                "workspace1-avatar.jpg",
                512000L,
                "image/jpeg"
        );

        String response1 = mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request1)))
                .andExpect(status().isOk())
                .andReturn()
                .getResponse()
                .getContentAsString();

        S3Service.PresignedUrlResponse presignedResponse1 = objectMapper.readValue(
                response1,
                S3Service.PresignedUrlResponse.class
        );

        UpdateProfileImageByKeyRequest updateRequest1 = new UpdateProfileImageByKeyRequest(
                testWorkspace.getWorkspaceId(),
                presignedResponse1.getFileKey()
        );

        mockMvc.perform(put("/api/profiles/me/image")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(updateRequest1)))
                .andExpect(status().isOk());

        // Update second workspace profile
        PresignedUrlRequest request2 = new PresignedUrlRequest(
                workspace2.getWorkspaceId(),
                "workspace2-avatar.png",
                256000L,
                "image/png"
        );

        String response2 = mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request2)))
                .andExpect(status().isOk())
                .andReturn()
                .getResponse()
                .getContentAsString();

        S3Service.PresignedUrlResponse presignedResponse2 = objectMapper.readValue(
                response2,
                S3Service.PresignedUrlResponse.class
        );

        UpdateProfileImageByKeyRequest updateRequest2 = new UpdateProfileImageByKeyRequest(
                workspace2.getWorkspaceId(),
                presignedResponse2.getFileKey()
        );

        mockMvc.perform(put("/api/profiles/me/image")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(updateRequest2)))
                .andExpect(status().isOk());

        // Verify both profiles have different images
        UserProfile updatedProfile1 = userProfileRepository.findById(testProfile.getProfileId()).orElseThrow();
        UserProfile updatedProfile2 = userProfileRepository.findById(profile2.getProfileId()).orElseThrow();

        assertThat(updatedProfile1.getProfileImageUrl()).contains(presignedResponse1.getFileKey());
        assertThat(updatedProfile2.getProfileImageUrl()).contains(presignedResponse2.getFileKey());
        assertThat(updatedProfile1.getProfileImageUrl()).isNotEqualTo(updatedProfile2.getProfileImageUrl());
    }

    @Test
    @DisplayName("통합 테스트: 에러 응답 형식 검증")
    @WithMockUser(username = "test-user-id")
    void testErrorResponseFormat() throws Exception {
        // Test with oversized file
        PresignedUrlRequest oversizedRequest = new PresignedUrlRequest(
                testWorkspace.getWorkspaceId(),
                "large.jpg",
                25 * 1024 * 1024L,
                "image/jpeg"
        );

        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(oversizedRequest)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.error").exists())
                .andExpect(jsonPath("$.error.code").exists())
                .andExpect(jsonPath("$.error.message").exists());

        // Test with invalid file type
        PresignedUrlRequest invalidTypeRequest = new PresignedUrlRequest(
                testWorkspace.getWorkspaceId(),
                "video.mp4",
                512000L,
                "video/mp4"
        );

        mockMvc.perform(post("/api/profiles/me/image/presigned-url")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(invalidTypeRequest)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.error").exists())
                .andExpect(jsonPath("$.error.code").exists())
                .andExpect(jsonPath("$.error.message").exists());
    }
}
