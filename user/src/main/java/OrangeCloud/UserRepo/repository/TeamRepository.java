package OrangeCloud.UserRepo.repository;

import OrangeCloud.UserRepo.entity.Team;
import jakarta.transaction.Transactional;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface TeamRepository extends JpaRepository<Team, UUID> {

    Optional<Team> findByTeamIdAndIsActiveTrue(UUID teamId);
    long countByIsActiveTrue();
    long countByGroupIdAndIsActiveTrue(UUID groupId);
    long countByLeaderIdAndIsActiveTrue(UUID leaderId);

    @Query("SELECT t FROM Team t WHERE t.isActive = true ORDER BY t.createdAt DESC")
    List<Team> findAllActiveTeams();

    @Query("SELECT t FROM Team t WHERE t.groupId = :groupId AND t.isActive = true ORDER BY t.createdAt DESC")
    List<Team> findActiveByGroupId(@Param("groupId") UUID groupId);

    @Query("SELECT t FROM Team t WHERE t.leaderId = :leaderId AND t.isActive = true ORDER BY t.createdAt DESC")
    List<Team> findActiveByLeaderId(@Param("leaderId") UUID leaderId);

    // 소프트 삭제 (Team 엔티티에 deletedAt 필드가 있는 경우에만)
    @Modifying
    @Transactional
    @Query("UPDATE Team t SET t.isActive = false, t.deletedAt = CURRENT_TIMESTAMP WHERE t.teamId = :teamId")
    int softDeleteById(@Param("teamId") UUID teamId);
    // 사용하지 않는 메서드들 모두 제거
    // - softDeleteById (deletedAt 필드가 없음)
    // - reactivateById
    // - findActiveByTeamNameContaining
    // - findInactiveTeams
    // - findActiveTeamsCreatedAfter
    // - softDeleteByGroupId
}