// src/main/java/OrangeCloud/UserRepo/repository/UserInfoRepository.java
package OrangeCloud.UserRepo.repository;

import OrangeCloud.UserRepo.entity.UserInfo;
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
public interface UserInfoRepository extends JpaRepository<UserInfo, UUID> {

    // 소프트 삭제
    @Modifying
    @Transactional
    @Query("UPDATE UserInfo u SET u.isActive = false, u.deletedAt = CURRENT_TIMESTAMP WHERE u.userId = :userId")
    int softDeleteById(@Param("userId") UUID userId);

    // 활성화된 사용자만 조회
    @Query("SELECT u FROM UserInfo u WHERE u.isActive = true")
    List<UserInfo> findAllActiveUsers();

    // ID로 활성화된 사용자 조회
    Optional<UserInfo> findByUserIdAndIsActiveTrue(UUID userId);

    // 그룹ID로 활성화된 사용자들 조회
    @Query("SELECT u FROM UserInfo u WHERE u.groupId = :groupId AND u.isActive = true")
    List<UserInfo> findActiveByGroupId(@Param("groupId") UUID groupId);

    // 역할별 활성화된 사용자 조회
    List<UserInfo> findByRoleAndIsActiveTrue(String role);

    // 그룹ID와 역할로 활성화된 사용자 조회
    List<UserInfo> findByGroupIdAndRoleAndIsActiveTrue(UUID groupId, String role);

    // 사용자 재활성화
    @Modifying
    @Transactional
    @Query("UPDATE UserInfo u SET u.isActive = true, u.deletedAt = null WHERE u.userId = :userId")
    int reactivateById(@Param("userId") UUID userId);

    // 그룹의 모든 사용자 비활성화
    @Modifying
    @Transactional
    @Query("UPDATE UserInfo u SET u.isActive = false, u.deletedAt = CURRENT_TIMESTAMP WHERE u.groupId = :groupId")
    int softDeleteByGroupId(@Param("groupId") UUID groupId);

    // 활성화된 사용자 수 조회
    long countByIsActiveTrue();

    // 그룹별 활성화된 사용자 수 조회
    long countByGroupIdAndIsActiveTrue(UUID groupId);

    // 비활성화된 사용자 조회
    @Query("SELECT u FROM UserInfo u WHERE u.isActive = false ORDER BY u.deletedAt DESC")
    List<UserInfo> findInactiveUsers();

    // 팀 역할이 없는 그룹의 활성화된 사용자 조회 (teamId -> teamRole IS NULL로 변경)
//    @Query("SELECT u FROM UserInfo u WHERE u.groupId = :groupId AND u.teamRole IS NULL AND u.isActive = true")
//    List<UserInfo> findActiveByGroupIdWithoutTeam(@Param("groupId") UUID groupId);

    // 그룹별 활성화된 사용자 수 조회 (teamId -> groupId로 변경)
//    @Query("SELECT COUNT(u) FROM UserInfo u WHERE u.groupId = :groupId AND u.isActive = true")
//    long countActiveByGroupId(@Param("groupId") UUID groupId);

    // 특정 팀 역할을 가진 그룹의 활성화된 사용자에서 팀 역할 제거 (teamId 관련 부분 제거)
//    @Modifying
//    @Transactional
//    @Query("UPDATE UserInfo u SET u.teamRole = null WHERE u.userId = :userId AND u.groupId = :groupId")
//    int removeTeamRoleFromUser(@Param("userId") UUID userId, @Param("groupId") UUID groupId);
//
//    // 그룹 내 모든 사용자의 팀 역할 제거 (teamId -> groupId로 변경)
//    @Modifying
//    @Transactional

    // 그룹에서 팀 역할이 있는 활성 사용자 수 조회 (추가)
//    long countByGroupIdAndIsActiveTrueAndTeamRoleIsNotNull(UUID groupId);

    // 또는 더 구체적으로 특정 팀 역할을 가진 사용자 수
    //long countByGroupIdAndIsActiveTrueAndTeamRole(UUID groupId, String teamRole);

    // 기존의 removeAllTeamRolesFromGroup 메서드 유지
//    @Modifying
//    @Transactional
//    @Query("UPDATE UserInfo u SET u.teamRole = null WHERE u.groupId = :groupId")
//    int removeAllTeamRolesFromGroup(@Param("groupId") UUID groupId);
}