package OrangeCloud.UserRepo.service;

import OrangeCloud.UserRepo.entity.Group;
import OrangeCloud.UserRepo.entity.Team;
import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.dto.team.CreateTeamRequest;
import OrangeCloud.UserRepo.dto.team.TeamWithMembersResponse;
import OrangeCloud.UserRepo.entity.UserInfo;
import OrangeCloud.UserRepo.repository.TeamRepository;
import OrangeCloud.UserRepo.repository.UserInfoRepository;
import OrangeCloud.UserRepo.repository.UserRepository;
import jakarta.persistence.EntityNotFoundException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Transactional
@Slf4j
public class TeamService {

    private final TeamRepository teamRepository;
    private final UserInfoRepository userInfoRepository;
    private final GroupService groupService;

    // 팀 생성 시 팀장을 멤버로 자동 추가
    public TeamWithMembersResponse createTeam(CreateTeamRequest request) {
        log.info("Creating team: {}", request);

        // 1. 팀장이 UserInfo에 존재하는지 확인
        UserInfo leader = userInfoRepository.findByUserIdAndIsActiveTrue(request.getLeaderId())
                .orElseThrow(() -> new IllegalArgumentException("팀장을 찾을 수 없습니다. User ID: " + request.getLeaderId()));

        // 2. 그룹 찾거나 생성
        Group group = groupService.findOrCreateGroupByCompanyName(
                request.getCompanyName(),
                request.getGroupName() != null ? request.getGroupName() : request.getCompanyName()
        );

        // 3. 팀 생성
        Team team = Team.builder()
                .teamName(request.getTeamName())
                .groupId(group.getGroupId())
                .leaderId(request.getLeaderId())
                .description(request.getDescription())
                .isActive(true)
                .build();

        // 4. 팀장을 멤버로 자동 추가
        team.addMember(request.getLeaderId());

        Team savedTeam = teamRepository.save(team);
        log.info("Created team with ID: {} by user: {}", savedTeam.getTeamId(), request.getLeaderId());

        // 5. 응답 생성
        List<UserInfo> teamMembers = List.of(leader);

        return TeamWithMembersResponse.builder()
                .team(savedTeam)
                .members(teamMembers)
                .group(group)
                .memberCount(1)
                .build();
    }

    // 팀에 멤버 추가 (다중 팀 가입 허용)
    public UserInfo addMemberToTeam(UUID teamId, UUID requesterId, UUID userId, String role) {
        log.info("Adding user {} to team {} by requester {}", userId, teamId, requesterId);

        // 1. 팀 확인
        Team team = teamRepository.findById(teamId)
                .orElseThrow(() -> new IllegalArgumentException("팀을 찾을 수 없습니다. Team ID: " + teamId));

        // 2. 팀장 권한 확인
        if (!team.getLeaderId().equals(requesterId)) {
            throw new IllegalArgumentException("팀원 추가 권한이 없습니다. 팀장만 팀원을 추가할 수 있습니다.");
        }

        // 3. 추가할 사용자 확인
        UserInfo userInfo = userInfoRepository.findByUserIdAndIsActiveTrue(userId)
                .orElseThrow(() -> new IllegalArgumentException("추가할 사용자를 찾을 수 없습니다. User ID: " + userId));

        // 4. 이미 해당 팀에 속해있는지 확인
        if (team.hasMember(userId)) {
            throw new IllegalArgumentException("사용자가 이미 해당 팀에 속해있습니다.");
        }

        // 5. 팀에 멤버 추가
        team.addMember(userId);
        teamRepository.save(team);

        log.info("Successfully added user {} to team {} (user can be in multiple teams)", userId, teamId);
        return userInfo;
    }

    // 팀에서 멤버 제거 (해당 팀에서만 제거)
    public boolean removeMemberFromTeam(UUID teamId, UUID requesterId, UUID userId) {
        log.info("Removing user {} from team {} by requester {}", userId, teamId, requesterId);

        // 1. 팀 확인
        Team team = teamRepository.findById(teamId)
                .orElseThrow(() -> new IllegalArgumentException("팀을 찾을 수 없습니다. Team ID: " + teamId));

        // 2. 팀장 권한 확인
        if (!team.getLeaderId().equals(requesterId)) {
            throw new IllegalArgumentException("팀원 제거 권한이 없습니다. 팀장만 팀원을 제거할 수 있습니다.");
        }

        // 3. 팀장 자신은 제거할 수 없음
        if (team.getLeaderId().equals(userId)) {
            throw new IllegalArgumentException("팀장은 자신을 팀에서 제거할 수 없습니다.");
        }

        // 4. 해당 팀에 속해있는지 확인
        if (!team.hasMember(userId)) {
            throw new IllegalArgumentException("사용자가 해당 팀에 속해있지 않습니다.");
        }

        // 5. 팀에서 멤버 제거 (다른 팀에는 영향 없음)
        team.removeMember(userId);
        teamRepository.save(team);

        log.info("Successfully removed user {} from team {} (other team memberships unaffected)", userId, teamId);
        return true;
    }

    // 특정 팀의 멤버 목록 조회
    public List<UserInfo> getTeamMembers(UUID teamId) {
        Team team = teamRepository.findById(teamId)
                .orElseThrow(() -> new EntityNotFoundException("팀을 찾을 수 없습니다."));

        List<UUID> memberIds = team.getMemberIdsList();
        if (memberIds.isEmpty()) {
            return List.of();
        }

        return userInfoRepository.findAllById(memberIds).stream()
                .filter(user -> user.getIsActive())
                .collect(Collectors.toList());
    }

    // 팀 멤버 수 조회
    public long getTeamMemberCount(UUID teamId) {
        Team team = teamRepository.findById(teamId)
                .orElseThrow(() -> new EntityNotFoundException("팀을 찾을 수 없습니다."));

        return team.getMemberIdsList().size();
    }

    // 사용자가 속한 모든 팀 조회
    public List<Team> getUserTeams(UUID userId) {
        // 모든 활성 팀에서 해당 사용자를 멤버로 가진 팀들 찾기
        return teamRepository.findAllActiveTeams().stream()
                .filter(team -> team.hasMember(userId))
                .collect(Collectors.toList());
    }

    // 사용자가 팀장인 팀들 조회
    public List<Team> getTeamsLedByUser(UUID userId) {
        return teamRepository.findActiveByLeaderId(userId);
    }

    // 모든 활성 사용자를 팀에 추가 가능 (중복 가입 허용)
    public List<UserInfo> getAvailableUsersForTeam(UUID teamId) {
        return userInfoRepository.findAllActiveUsers();
    }

    // 사용자가 특정 팀의 멤버인지 확인
    public boolean isTeamMember(UUID teamId, UUID userId) {
        Optional<Team> teamOpt = teamRepository.findById(teamId);
        return teamOpt.map(team -> team.hasMember(userId)).orElse(false);
    }

    // 나머지 기존 메서드들...
    public List<Team> getAllActiveTeams() {
        return teamRepository.findAllActiveTeams();
    }

    public Optional<Team> getActiveTeamById(UUID teamId) {
        return teamRepository.findByTeamIdAndIsActiveTrue(teamId);
    }

    public List<Team> getActiveTeamsByGroupId(UUID groupId) {
        return teamRepository.findActiveByGroupId(groupId);
    }

    public List<Team> getActiveTeamsByLeaderId(UUID leaderId) {
        return teamRepository.findActiveByLeaderId(leaderId);
    }

    public boolean isTeamLeader(UUID teamId, UUID userId) {
        Optional<Team> team = getActiveTeamById(teamId);
        return team.isPresent() && team.get().getLeaderId().equals(userId);
    }

    public Team updateTeam(UUID teamId, UUID requesterId, String newTeamName, String newDescription) {
        Team team = teamRepository.findById(teamId)
                .orElseThrow(() -> new EntityNotFoundException("팀을 찾을 수 없습니다."));

        if (!team.getLeaderId().equals(requesterId)) {
            throw new IllegalArgumentException("팀 수정 권한이 없습니다.");
        }

        if (newTeamName != null) {
            team.setTeamName(newTeamName);
        }
        if (newDescription != null) {
            team.setDescription(newDescription);
        }

        return teamRepository.save(team);
    }

    public Team changeTeamLeader(UUID teamId, UUID currentLeaderId, UUID newLeaderId) {
        Team team = teamRepository.findById(teamId)
                .orElseThrow(() -> new EntityNotFoundException("팀을 찾을 수 없습니다."));

        if (!team.getLeaderId().equals(currentLeaderId)) {
            throw new IllegalArgumentException("팀장 변경 권한이 없습니다.");
        }

        // 새 팀장이 활성 사용자인지 확인
        UserInfo newLeader = userInfoRepository.findByUserIdAndIsActiveTrue(newLeaderId)
                .orElseThrow(() -> new IllegalArgumentException("새 팀장을 찾을 수 없습니다."));

        // 새 팀장을 팀 멤버로 추가 (아직 멤버가 아니라면)
        if (!team.hasMember(newLeaderId)) {
            team.addMember(newLeaderId);
        }

        team.setLeaderId(newLeaderId);
        return teamRepository.save(team);
    }

    public boolean disbandTeam(UUID teamId, UUID requesterId) {
        log.info("Disbanding team {} by requester {}", teamId, requesterId);

        Optional<Team> teamOpt = getActiveTeamById(teamId);
        if (teamOpt.isEmpty()) {
            throw new IllegalArgumentException("팀을 찾을 수 없습니다. Team ID: " + teamId);
        }

        Team team = teamOpt.get();

        if (!team.getLeaderId().equals(requesterId)) {
            throw new IllegalArgumentException("팀 해산 권한이 없습니다. 팀장만 팀을 해산할 수 있습니다.");
        }

        try {
            teamRepository.deleteById(teamId);
            log.info("Successfully disbanded team {} (UserInfo members remain unaffected)", teamId);
            return true;
        } catch (Exception e) {
            log.error("Error disbanding team {}", teamId, e);
            return false;
        }
    }
}