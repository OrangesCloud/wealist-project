package OrangeCloud.UserRepo.repository;

import OrangeCloud.UserRepo.entity.Group;
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
public interface GroupRepository extends JpaRepository<Group, UUID> {

    // 1. 그룹 비활성화 (소프트 삭제)
    @Modifying
    @Transactional
    @Query("UPDATE Group g SET g.isActive = false WHERE g.groupId = :groupId")
    int deactivateById(@Param("groupId") UUID groupId);

    // 2. 삭제 시간과 함께 비활성화
    @Modifying
    @Transactional
    @Query("UPDATE Group g SET g.isActive = false, g.deletedAt = :deletedAt WHERE g.groupId = :groupId")
    int deactivateByIdWithTimestamp(@Param("groupId") UUID groupId, @Param("deletedAt") LocalDateTime deletedAt);

    // 3. 현재 시간으로 비활성화
    @Modifying
    @Transactional
    @Query("UPDATE Group g SET g.isActive = false, g.deletedAt = CURRENT_TIMESTAMP WHERE g.groupId = :groupId")
    int softDeleteById(@Param("groupId") UUID groupId);

    // 4. 활성화된 그룹만 조회
    @Query("SELECT g FROM Group g WHERE g.isActive = true")
    List<Group> findAllActiveGroups();

    // 5. ID로 활성화된 그룹 조회
    @Query("SELECT g FROM Group g WHERE g.groupId = :groupId AND g.isActive = true")
    Optional<Group> findActiveById(@Param("groupId") UUID groupId);

    // 6. 메서드 네이밍으로 활성화된 그룹 조회
    List<Group> findByIsActiveTrue();
    Optional<Group> findByGroupIdAndIsActiveTrue(UUID groupId);

//    // 7. 회사명으로 활성화된 그룹 조회
//    @Query("SELECT g FROM Group g WHERE g.companyName = :companyName AND g.isActive = true")
//    List<Group> findActiveByCompanyName(@Param("companyName") String companyName);

    // 8. 그룹명으로 활성화된 그룹 조회
    @Query("SELECT g FROM Group g WHERE g.name LIKE %:name% AND g.isActive = true")
    List<Group> findActiveByNameContaining(@Param("name") String name);

    // 9. 그룹 재활성화
    @Modifying
    @Transactional
    @Query("UPDATE Group g SET g.isActive = true, g.deletedAt = null WHERE g.groupId = :groupId")
    int reactivateById(@Param("groupId") UUID groupId);

    // 10. Native Query 사용 (PostgreSQL 기준)
    @Modifying
    @Transactional
    @Query(value = "UPDATE groups SET is_active = false, deleted_at = NOW() WHERE group_id = ?1",
            nativeQuery = true)
    int deactivateByIdNative(UUID groupId);

    // 11. 활성화된 그룹 수 조회
    @Query("SELECT COUNT(g) FROM Group g WHERE g.isActive = true")
    long countActiveGroups();

    // 12. 비활성화된 그룹 조회 (관리자용)
    @Query("SELECT g FROM Group g WHERE g.isActive = false ORDER BY g.deletedAt DESC")
    List<Group> findInactiveGroups();
    // 회사명으로 활성화된 그룹 조회
    @Query("SELECT g FROM Group g WHERE g.companyName = :companyName AND g.isActive = true ORDER BY g.createdAt ASC")
    List<Group> findActiveByCompanyName(@Param("companyName") String companyName);

    // 회사명과 그룹명이 모두 일치하는 활성화된 그룹 조회
    @Query("SELECT g FROM Group g WHERE g.companyName = :companyName AND g.name = :groupName AND g.isActive = true")
    Optional<Group> findActiveByCompanyNameAndGroupName(@Param("companyName") String companyName, @Param("groupName") String groupName);

    // 회사명 중복 체크
    @Query("SELECT COUNT(g) > 0 FROM Group g WHERE g.companyName = :companyName AND g.isActive = true")
    boolean existsActiveByCompanyName(@Param("companyName") String companyName);
    // 회사별 그룹 수 조회
    @Query("SELECT COUNT(g) FROM Group g WHERE g.companyName = :companyName AND g.isActive = true")
    long countActiveByCompanyName(@Param("companyName") String companyName);
    // 나머지 기존 메서드들...

}