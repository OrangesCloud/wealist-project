package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.userprofile.CreateProfileRequest;
import OrangeCloud.UserRepo.dto.userprofile.UpdateProfileRequest;
import OrangeCloud.UserRepo.dto.userprofile.UserProfileResponse;
import OrangeCloud.UserRepo.service.UserProfileService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.security.Principal;
import java.util.UUID;

@RestController
@RequestMapping("/api/profiles")
@RequiredArgsConstructor
@Tag(name = "UserProfile", description = "사용자 프로필 조회 및 수정 API")
@Slf4j
public class UserProfileController {

    private final UserProfileService userProfileService;

    private UUID extractUserId(Principal principal) {
        if (principal instanceof Authentication authentication) {
            return UUID.fromString(authentication.getName());
        }
        throw new IllegalStateException("인증된 사용자 정보를 찾을 수 없습니다.");
    }

    @PostMapping
    @Operation(summary = "프로필 생성", description = "새로운 프로필을 생성합니다.")
    public ResponseEntity<UserProfileResponse> createProfile(
            Principal principal,
            @Valid @RequestBody CreateProfileRequest request) {
        UUID userId = extractUserId(principal);
        log.info("Creating profile for user: {}", userId);
        UserProfileResponse response = userProfileService.createProfile(request, userId);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/me")
    @Operation(summary = "내 프로필 조회", description = "내 프로필을 조회합니다.")
    public ResponseEntity<UserProfileResponse> getMyProfile(Principal principal) {
        UUID userId = extractUserId(principal);
        UserProfileResponse response = userProfileService.getProfile(userId);
        return ResponseEntity.ok(response);
    }

    @PutMapping("/me")
    @Operation(summary = "내 프로필 정보 통합 업데이트", description = "인증된 사용자의 이름 또는 프로필 이미지 URL을 업데이트합니다.")
    public ResponseEntity<UserProfileResponse> updateMyProfile(
            Principal principal,
            @Valid @RequestBody UpdateProfileRequest request
    ) {
        UUID userId = extractUserId(principal);
        log.info("Received integrated profile update request for user: {}", userId);

        // Ensure the userId in the request matches the authenticated user
        if (!request.userId().equals(userId)) {
            throw new IllegalArgumentException("User ID in request does not match authenticated user.");
        }

        UserProfileResponse updatedProfile = userProfileService.updateProfile(request);
        return ResponseEntity.ok(updatedProfile);
    }

    @DeleteMapping("/{workspaceId}")
    @Operation(summary = "프로필 삭제", description = "특정 워크스페이스의 프로필을 삭제합니다.")
    public ResponseEntity<Void> deleteProfile(
            Principal principal,
            @PathVariable UUID workspaceId) {
        UUID userId = extractUserId(principal);
        log.info("Deleting profile for user: {} in workspace: {}", userId, workspaceId);
        userProfileService.deleteProfile(userId, workspaceId);
        return ResponseEntity.noContent().build();
    }
}