package OrangeCloud.UserRepo.repository;

import OrangeCloud.UserRepo.entity.User;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface UserRepository extends JpaRepository<User, UUID> {

    // 1. 소프트 삭제
    @Modifying
    @Transactional
    @Query("UPDATE User u SET u.isActive = false, u.deletedAt = CURRENT_TIMESTAMP WHERE u.userId = :userId")
    int softDeleteById(@Param("userId") UUID userId);

    // 2. 사용자 재활성화
    @Modifying
    @Transactional
    @Query("UPDATE User u SET u.isActive = true, u.deletedAt = null WHERE u.userId = :userId")
    int reactivateById(@Param("userId") UUID userId);

    // 3. 활성화된 사용자만 조회
    @Query("SELECT u FROM User u WHERE u.isActive = true")
    List<User> findAllActiveUsers();

    // 4. ID로 활성화된 사용자 조회
    Optional<User> findByUserIdAndIsActiveTrue(UUID userId);

    // 5. 이메일로 활성화된 사용자 조회
    Optional<User> findByEmailAndIsActiveTrue(String email);

    // 6. 이름으로 활성화된 사용자 검색
    @Query("SELECT u FROM User u WHERE u.name LIKE %:name% AND u.isActive = true")
    List<User> findActiveByNameContaining(@Param("name") String name);

    // 7. 이메일 중복 체크 (활성화된 사용자만)
    boolean existsByEmailAndIsActiveTrue(String email);

    // 8. 활성화된 사용자 수 조회
    long countByIsActiveTrue();

    // 9. 비활성화된 사용자 조회 (관리자용)
    @Query("SELECT u FROM User u WHERE u.isActive = false ORDER BY u.deletedAt DESC")
    List<User> findInactiveUsers();

    // 10. 생성일 기준 활성화된 사용자 조회
    @Query("SELECT u FROM User u WHERE u.createdAt >= :startDate AND u.isActive = true ORDER BY u.createdAt DESC")
    List<User> findActiveUsersCreatedAfter(@Param("startDate") LocalDateTime startDate);

    // 11. Native Query 사용
    @Modifying
    @Transactional
    @Query(value = "UPDATE users SET is_active = false, deleted_at = NOW() WHERE user_id = ?1",
            nativeQuery = true)
    int softDeleteByIdNative(UUID userId);

    // 12. 이메일과 비밀번호로 활성화된 사용자 조회 (로그인용)
    @Query("SELECT u FROM User u WHERE u.email = :email AND u.isActive = true")
    Optional<User> findActiveByEmail(@Param("email") String email);
    // 13. uuid 기반 해당하는 모든 사용자 정보
    List<User> findAllByUserIdIn(List<UUID> ids);
    // 14. 이메일로만 검색
    List<User> findAllByEmailAndIsActiveTrue(String email);
}