package OrangeCloud.UserRepo.service;

import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.repository.UserRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
@RequiredArgsConstructor
@Transactional
@Slf4j
public class UserService {

    private final UserRepository userRepository;
    private final PasswordEncoder passwordEncoder; // 비밀번호 암호화용

    // 소프트 삭제
    public boolean softDeleteUser(UUID userId) {
        log.info("Soft deleting user with ID: {}", userId);
        int updatedRows = userRepository.softDeleteById(userId);
        boolean success = updatedRows > 0;
        log.info("User soft delete result: {}", success);
        return success;
    }

    // 사용자 재활성화
    public boolean reactivateUser(UUID userId) {
        log.info("Reactivating user with ID: {}", userId);
        int updatedRows = userRepository.reactivateById(userId);
        boolean success = updatedRows > 0;
        log.info("User reactivation result: {}", success);
        return success;
    }

    // 활성화된 모든 사용자 조회
    @Transactional(readOnly = true)
    public List<User> getAllActiveUsers() {
        log.debug("Fetching all active users");
        return userRepository.findAllActiveUsers();
    }

    // ID로 활성화된 사용자 조회
    @Transactional(readOnly = true)
    public Optional<User> getActiveUserById(UUID userId) {
        log.debug("Fetching active user by ID: {}", userId);
        return userRepository.findByUserIdAndIsActiveTrue(userId);
    }

    // 이메일로 활성화된 사용자 조회
//    @Transactional(readOnly = true)
//    public Optional<User> getActiveUserByEmail(String email) {
//        log.debug("Fetching active user by email: {}", email);
//        return userRepository.findByEmailAndIsActiveTrue(email);
//    }
    @Transactional(readOnly = true)
    public List<User> getUsersByUserIds(List<UUID> userIds) {
        return userRepository.findAllByUserIdIn(userIds);
    }

    // 이름으로 사용자 검색
    @Transactional(readOnly = true)
    public List<User> searchActiveUsersByName(String name) {
        log.debug("Searching active users by name: {}", name);
        return userRepository.findActiveByNameContaining(name);
    }

    // 사용자 정보 수정
//    public Optional<User> updateUser(UUID userId, String name, String email) {
//        log.info("Updating user: {} with name: {} and email: {}", userId, name, email);
//
//        return userRepository.findByUserIdAndIsActiveTrue(userId)
//                .map(user -> {
//                    // 이메일 변경 시 중복 체크
//                    if (!user.getEmail().equals(email) && userRepository.existsByEmailAndIsActiveTrue(email)) {
//                        throw new IllegalArgumentException("이미 사용 중인 이메일입니다: " + email);
//                    }
//
//                    user.setName(name);
//                    user.setEmail(email);
//                    User updatedUser = userRepository.save(user);
//                    log.info("Successfully updated user: {}", userId);
//                    return updatedUser;
//                });
//    }

    // 이메일 or 사람 이름으로 검색
    @Transactional(readOnly = true)
    public List<User> searchUsers(String query) {
        // 이메일 형식이면 이메일로 검색
        if (isEmailFormat(query)) {
            return userRepository.findAllByEmailAndIsActiveTrue(query);
        }
        // 아니면 이름 부분 검색
        return userRepository.findActiveByNameContaining(query);
    }
    // 로그인 검증
    @Transactional(readOnly = true)
    public Optional<User> authenticateUser(String email) {
        log.debug("Authenticating user with email: {}", email);

        Optional<User> userOpt = userRepository.findActiveByEmail(email);
        if (userOpt.isPresent()) {
            User user = userOpt.get();
                return Optional.of(user);
        }

        log.warn("Authentication failed for email: {}", email);
        return Optional.empty();
    }

    // 활성화된 사용자 수 조회
    @Transactional(readOnly = true)
    public long getActiveUserCount() {
        return userRepository.countByIsActiveTrue();
    }

    // 비활성화된 사용자 조회 (관리자용)
    @Transactional(readOnly = true)
    public List<User> getInactiveUsers() {
        return userRepository.findInactiveUsers();
    }

    // 최근 가입한 활성화된 사용자 조회
    @Transactional(readOnly = true)
    public List<User> getRecentActiveUsers(int days) {
        LocalDateTime startDate = LocalDateTime.now().minusDays(days);
        return userRepository.findActiveUsersCreatedAfter(startDate);
    }

    // 이메일 중복 체크
    @Transactional(readOnly = true)
    public boolean isEmailAvailable(String email) {
        return !userRepository.existsByEmailAndIsActiveTrue(email);
    }

    private boolean isEmailFormat(String query) {
        return query.matches("^[A-Za-z0-9+_.-]+@[A-Za-z0-9.-]+$");
    }
}