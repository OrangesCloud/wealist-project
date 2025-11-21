package OrangeCloud.UserRepo.service;

import OrangeCloud.UserRepo.dto.userprofile.CreateProfileRequest;
import OrangeCloud.UserRepo.dto.userprofile.UserProfileResponse;
import OrangeCloud.UserRepo.entity.UserProfile;
import OrangeCloud.UserRepo.repository.UserProfileRepository;
import OrangeCloud.UserRepo.exception.UserNotFoundException; // ✅ UserNotFoundException을 사용
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.cache.annotation.CacheEvict;
import org.springframework.cache.annotation.Cacheable;
// import org.springframework.data.crossstore.ChangeSetPersister.NotFoundException; // 🚫 불필요한 임포트 제거
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import OrangeCloud.UserRepo.dto.userprofile.UpdateProfileRequest;

import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
public class UserProfileService {

    private final UserProfileRepository userProfileRepository;
    private static final UUID DEFAULT_WORKSPACE_ID = UUID.fromString("00000000-0000-0000-0000-000000000000");

    @Transactional
    public UserProfileResponse createProfile(CreateProfileRequest request, UUID userId) {
        log.info("Creating profile for user: {}", userId);
        UserProfile userProfile = UserProfile.create(
                request.workspaceId(),
                userId,
                request.nickName(),
                request.email(),
                null
        );
        UserProfile savedProfile = userProfileRepository.save(userProfile);
        return UserProfileResponse.from(savedProfile);
    }

    /**
     * 사용자 프로필을 조회하고 DTO로 반환합니다. (Redis 캐시 적용)
     * 캐시 이름: "userProfile", 키: userId
     */
    @Transactional(readOnly = true)
    @Cacheable(value = "userProfile", key = "#userId")
    // 💡 수정: 반환 타입을 UserProfileResponse DTO로 변경
    public UserProfileResponse getProfile(UUID userId) {
        log.info("[Cacheable] Attempting to retrieve profile from DB for user: {}", userId);
        UUID defaultId = DEFAULT_WORKSPACE_ID;
        // DB 조회 (UserProfile 엔티티)
        UserProfile profile = userProfileRepository.findByWorkspaceIdAndUserId(DEFAULT_WORKSPACE_ID,userId)
                // 💡 수정: 정의된 UserNotFoundException을 사용
                .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));

        // 💡 수정: DTO를 반환하도록 로직을 유지
        return UserProfileResponse.from(profile);
    }
    @Transactional(readOnly = true)
    @Cacheable(value = "userProfile", key = "#userId")
    // 💡 수정: 반환 타입을 UserProfileResponse DTO로 변경
    public UserProfileResponse workSpaceIdGetProfile(UUID workspaceId,UUID userId) {
        log.info("[Cacheable] Attempting to retrieve profile from DB for user: {}", userId);

        // DB 조회 (UserProfile 엔티티)
        UserProfile profile = userProfileRepository.findByWorkspaceIdAndUserId(workspaceId,userId)
                // 💡 수정: 정의된 UserNotFoundException을 사용
                .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));

        // 💡 수정: DTO를 반환하도록 로직을 유지
        return UserProfileResponse.from(profile);
    }

    // 해당 사용자id에 따른 모든 프로필 가져오기
    @Transactional(readOnly = true)
    @Cacheable(value = "userProfiles", key = "#userId")
    public List<UserProfileResponse> getAllProfiles(UUID userId) {
        log.info("[Cacheable] Attempting to retrieve all profiles from DB for user: {}", userId);

        // DB 조회 (해당 사용자의 모든 UserProfile 엔티티)
        List<UserProfile> profiles = userProfileRepository.findAllByUserId(userId);

        if (profiles.isEmpty()) {
            throw new UserNotFoundException("사용자의 프로필을 찾을 수 없습니다.");
        }

        // DTO 변환 후 반환
        return profiles.stream()
                .map(UserProfileResponse::from)
                .collect(Collectors.toList());

    }



    /**
     * 사용자 프로필 닉네임, 이메일 및 이미지 URL을 통합 업데이트하고 캐시를 무효화합니다.
     * @param request
     * @return 업데이트된 UserProfile 엔티티 (Service 내부에서 사용되므로 엔티티 반환 유지)
     */
    @Transactional
    @CacheEvict(value = "userProfile", key = "#request.userId")
    public UserProfileResponse updateProfile(UpdateProfileRequest request) {
        log.info("[CacheEvict] Updating profile for user: userId={}, nickName={}, email={}, imageUrl={}", request.userId(), request.nickName(), request.email(), request.profileImageUrl());

        // 1. UserProfile 조회
        UserProfile profile = userProfileRepository.findByWorkspaceIdAndUserId(request.workspaceId(), request.userId())
                .orElseThrow(() -> new UserNotFoundException("프로필 업데이트 대상 사용자를 찾을 수 없습니다."));

        // 2. 닉네임 업데이트 (값이 존재하고 비어있지 않을 경우에만)
        if (request.nickName() != null && !request.nickName().trim().isEmpty()) {
            profile.updateNickName(request.nickName().trim());
            log.debug("Profile nickName updated to: {}", request.nickName().trim());
        }

        // 3. 이메일 업데이트 (값이 존재하고 비어있지 않을 경우에만)
        if (request.email() != null && !request.email().trim().isEmpty()) {
            profile.updateEmail(request.email().trim());
            log.debug("Profile email updated to: {}", request.email().trim());
        }

        // 4. 이미지 URL 업데이트
        if (request.profileImageUrl() != null) {
            String urlToSave = request.profileImageUrl().trim().isEmpty() ? null : request.profileImageUrl().trim();
            profile.updateProfileImageUrl(urlToSave);
            log.debug("Profile image URL updated to: {}", urlToSave);
        }

        // 5. 변경된 프로필 저장
        UserProfile updatedProfile = userProfileRepository.save(profile);
        return UserProfileResponse.from(updatedProfile);
    }


    /**
     * 사용자 프로필을 삭제하고 캐시를 무효화합니다.
     * @param userId 사용자 ID (UUID)
     * @param workspaceId 워크스페이스 ID (UUID)
     */
    @Transactional
    @CacheEvict(value = "userProfile", key = "#userId")
    public void deleteProfile(UUID userId, UUID workspaceId) {
        log.info("[CacheEvict] Deleting profile for user: userId={}, workspaceId={}", userId, workspaceId);

        UserProfile profile = userProfileRepository.findByWorkspaceIdAndUserId(workspaceId, userId)
                .orElseThrow(() -> new UserNotFoundException("삭제할 프로필을 찾을 수 없습니다."));

        userProfileRepository.delete(profile);
    }

}