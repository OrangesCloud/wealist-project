package OrangeCloud.UserRepo.service;

import OrangeCloud.UserRepo.entity.UserInfo;
import OrangeCloud.UserRepo.repository.UserInfoRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
@RequiredArgsConstructor
@Transactional
@Slf4j
public class UserInfoService {

    private final UserInfoRepository userInfoRepository;

    // 소프트 삭제
    public boolean softDeleteUser(UUID userId) {
        log.info("Soft deleting user with ID: {}", userId);
        int updatedRows = userInfoRepository.softDeleteById(userId);
        boolean success = updatedRows > 0;
        log.info("User soft delete result: {}", success);
        return success;
    }

    // 활성화된 모든 사용자 조회
    @Transactional(readOnly = true)
    public List<UserInfo> getAllActiveUsers() {
        log.debug("Fetching all active users");
        return userInfoRepository.findAllActiveUsers();
    }

    // ID로 활성화된 사용자 조회
    @Transactional(readOnly = true)
    public Optional<UserInfo> getActiveUserById(UUID userId) {
        log.debug("Fetching active user by ID: {}", userId);
        return userInfoRepository.findByUserIdAndIsActiveTrue(userId);
    }

    // 그룹별 활성화된 사용자 조회
    @Transactional(readOnly = true)
    public List<UserInfo> getActiveUsersByGroupId(UUID groupId) {
        log.debug("Fetching active users by group ID: {}", groupId);
        return userInfoRepository.findActiveByGroupId(groupId);
    }

    // 역할별 활성화된 사용자 조회
    @Transactional(readOnly = true)
    public List<UserInfo> getActiveUsersByRole(String role) {
        log.debug("Fetching active users by role: {}", role);
        return userInfoRepository.findByRoleAndIsActiveTrue(role);
    }

    // 사용자 재활성화
    public boolean reactivateUser(UUID userId) {
        log.info("Reactivating user with ID: {}", userId);
        int updatedRows = userInfoRepository.reactivateById(userId);
        boolean success = updatedRows > 0;
        log.info("User reactivation result: {}", success);
        return success;
    }

    // 그룹의 모든 사용자 비활성화
    public boolean softDeleteUsersByGroupId(UUID groupId) {
        log.info("Soft deleting all users in group: {}", groupId);
        int updatedRows = userInfoRepository.softDeleteByGroupId(groupId);
        boolean success = updatedRows > 0;
        log.info("Group users soft delete result: {} users affected", updatedRows);
        return success;
    }

    // 활성화된 사용자 수 조회
    @Transactional(readOnly = true)
    public long getActiveUserCount() {
        return userInfoRepository.countByIsActiveTrue();
    }

    // 사용자 생성
    public UserInfo createUser(UUID groupId, String Role,UUID userId) {
        log.info("Creating new user with groupId: {} and role: {}", groupId, Role);
        UserInfo userInfo = UserInfo.builder()
                .groupId(groupId)
                .userId(userId)
                .role(Role)
                .isActive(true)
                .build();

        UserInfo savedUser = userInfoRepository.save(userInfo);
        log.info("Created user with ID: {}", savedUser.getUserId());
        return savedUser;
    }

    // 사용자 정보 수정
    public Optional<UserInfo> updateUser(UUID userId, UUID groupId, String role) {
        log.info("Updating user: {} with groupId: {} and role: {}", userId, groupId, role);

        return userInfoRepository.findByUserIdAndIsActiveTrue(userId)
                .map(userInfo -> {
                    userInfo.setGroupId(groupId);
                    userInfo.setRole(role);
                    UserInfo updatedUser = userInfoRepository.save(userInfo);
                    log.info("Successfully updated user: {}", userId);
                    return updatedUser;
                });
    }
}