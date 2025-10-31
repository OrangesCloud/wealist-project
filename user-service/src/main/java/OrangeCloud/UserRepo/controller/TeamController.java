package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.MessageApiResponse;
import OrangeCloud.UserRepo.dto.team.CreateTeamRequest;
import OrangeCloud.UserRepo.dto.team.TeamWithMembersResponse;
import OrangeCloud.UserRepo.entity.Team;
import OrangeCloud.UserRepo.entity.UserInfo;
import OrangeCloud.UserRepo.service.TeamService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@RestController
@RequestMapping("/api/teams")
@RequiredArgsConstructor
@Slf4j
public class TeamController {

    private final TeamService teamService;

    // 팀 생성 (UserInfo에 존재하는 사용자만 가능)
    @PostMapping
    public ResponseEntity<MessageApiResponse> createTeam(@Valid @RequestBody CreateTeamRequest request) {
        log.info("Request to create team: {}", request);

        try {
            TeamWithMembersResponse team = teamService.createTeam(request);
            return ResponseEntity.ok(MessageApiResponse.success("팀이 성공적으로 생성되었습니다.", team));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest()
                    .body(MessageApiResponse.failure(e.getMessage()));
        }
    }

    // 활성화된 팀 목록 조회
    @GetMapping
    public ResponseEntity<MessageApiResponse> getAllActiveTeams() {
        log.info("Request to get all active teams");
        List<Team> teams = teamService.getAllActiveTeams();
        return ResponseEntity.ok(MessageApiResponse.success("활성화된 팀 목록을 성공적으로 조회했습니다.", teams));
    }

    // 특정 팀 조회
    @GetMapping("/{teamId}")
    public ResponseEntity<MessageApiResponse> getActiveTeam(@PathVariable UUID teamId) {
        log.info("Request to get team: {}", teamId);

        Optional<Team> team = teamService.getActiveTeamById(teamId);
        if (team.isPresent()) {
            return ResponseEntity.ok(MessageApiResponse.success("팀 정보를 성공적으로 조회했습니다.", team.get()));
        } else {
            return ResponseEntity.notFound().build();
        }
    }

    // 그룹별 활성화된 팀 조회
    @GetMapping("/group/{groupId}")
    public ResponseEntity<MessageApiResponse> getActiveTeamsByGroup(@PathVariable UUID groupId) {
        log.info("Request to get active teams by group: {}", groupId);
        List<Team> teams = teamService.getActiveTeamsByGroupId(groupId);
        return ResponseEntity.ok(MessageApiResponse.success("그룹의 팀 목록을 성공적으로 조회했습니다.", teams));
    }

    // 팀장별 활성화된 팀 조회
    @GetMapping("/leader/{leaderId}")
    public ResponseEntity<MessageApiResponse> getActiveTeamsByLeader(@PathVariable UUID leaderId) {
        log.info("Request to get active teams by leader: {}", leaderId);
        List<Team> teams = teamService.getActiveTeamsByLeaderId(leaderId);
        return ResponseEntity.ok(MessageApiResponse.success("팀장의 팀 목록을 성공적으로 조회했습니다.", teams));
    }

    // 팀명으로 검색 - 메서드 없음으로 주석 처리
    /*
    @GetMapping("/search")
    public ResponseEntity<MessageApiResponse> searchTeamsByName(@RequestParam String teamName) {
        log.info("Request to search teams by name: {}", teamName);
        List<Team> teams = teamService.searchActiveTeamsByName(teamName);
        return ResponseEntity.ok(MessageApiResponse.success("팀 검색 결과를 성공적으로 조회했습니다.", teams));
    }
    */

    // 팀 멤버 목록 조회 (해당 그룹의 UserInfo들)
    @GetMapping("/{teamId}/members")
    public ResponseEntity<MessageApiResponse> getTeamMembers(@PathVariable UUID teamId) {
        log.info("Request to get team members: {}", teamId);

        try {
            List<UserInfo> members = teamService.getTeamMembers(teamId);
            return ResponseEntity.ok(MessageApiResponse.success("팀 멤버 목록을 성공적으로 조회했습니다.", members));
        } catch (Exception e) {
            return ResponseEntity.badRequest()
                    .body(MessageApiResponse.failure("팀 멤버 조회에 실패했습니다: " + e.getMessage()));
        }
    }

    // 팀 멤버 수 조회
    @GetMapping("/{teamId}/members/count")
    public ResponseEntity<MessageApiResponse> getTeamMemberCount(@PathVariable UUID teamId) {
        log.info("Request to get team member count: {}", teamId);

        try {
            long count = teamService.getTeamMemberCount(teamId);
            return ResponseEntity.ok(MessageApiResponse.success("팀 멤버 수를 성공적으로 조회했습니다.", count));
        } catch (Exception e) {
            return ResponseEntity.badRequest()
                    .body(MessageApiResponse.failure("팀 멤버 수 조회에 실패했습니다: " + e.getMessage()));
        }
    }

    // 특정 그룹에서 사용 가능한 사용자 목록 조회
    @GetMapping("/available-users/{groupId}")
    public ResponseEntity<MessageApiResponse> getAvailableUsersForTeam(@PathVariable UUID groupId) {
        log.info("Request to get available users for team in group: {}", groupId);

        List<UserInfo> availableUsers = teamService.getAvailableUsersForTeam(groupId);
        return ResponseEntity.ok(MessageApiResponse.success("팀에 할당 가능한 사용자 목록을 성공적으로 조회했습니다.", availableUsers));
    }

    // 팀장 권한 확인
    @GetMapping("/{teamId}/leader/check")
    public ResponseEntity<MessageApiResponse> checkTeamLeader(@PathVariable UUID teamId,
                                                              @RequestParam UUID userId) {
        log.info("Request to check if user {} is leader of team: {}", userId, teamId);

        boolean isLeader = teamService.isTeamLeader(teamId, userId);
        String message = isLeader ? "사용자가 해당 팀의 팀장입니다." : "사용자가 해당 팀의 팀장이 아닙니다.";
        return ResponseEntity.ok(MessageApiResponse.success(message, isLeader));
    }

    // 사용자가 속한 팀들 조회
    @GetMapping("/user/{userId}")
    public ResponseEntity<MessageApiResponse> getUserTeams(@PathVariable UUID userId) {
        log.info("Request to get teams for user: {}", userId);

        List<Team> userTeams = teamService.getUserTeams(userId);
        return ResponseEntity.ok(MessageApiResponse.success("사용자의 팀 목록을 성공적으로 조회했습니다.", userTeams));
    }

    // 사용자가 팀장인 팀들 조회
    @GetMapping("/user/{userId}/led-teams")
    public ResponseEntity<MessageApiResponse> getTeamsLedByUser(@PathVariable UUID userId) {
        log.info("Request to get teams led by user: {}", userId);

        List<Team> ledTeams = teamService.getTeamsLedByUser(userId);
        return ResponseEntity.ok(MessageApiResponse.success("사용자가 팀장인 팀 목록을 성공적으로 조회했습니다.", ledTeams));
    }

    // 팀 정보 업데이트
    @PutMapping("/{teamId}")
    public ResponseEntity<MessageApiResponse> updateTeam(@PathVariable UUID teamId,
                                                         @RequestParam UUID requesterId,
                                                         @RequestParam(required = false) String teamName,
                                                         @RequestParam(required = false) String description) {
        log.info("Request to update team: {} by requester: {}", teamId, requesterId);

        try {
            Team updatedTeam = teamService.updateTeam(teamId, requesterId, teamName, description);
            return ResponseEntity.ok(MessageApiResponse.success("팀 정보가 성공적으로 업데이트되었습니다.", updatedTeam));
        } catch (Exception e) {
            return ResponseEntity.badRequest()
                    .body(MessageApiResponse.failure(e.getMessage()));
        }
    }

    // 팀장 변경
    @PutMapping("/{teamId}/leader")
    public ResponseEntity<MessageApiResponse> changeTeamLeader(@PathVariable UUID teamId,
                                                               @RequestParam UUID currentLeaderId,
                                                               @RequestParam UUID newLeaderId) {
        log.info("Request to change team leader: {} from {} to {}", teamId, currentLeaderId, newLeaderId);

        try {
            Team updatedTeam = teamService.changeTeamLeader(teamId, currentLeaderId, newLeaderId);
            return ResponseEntity.ok(MessageApiResponse.success("팀장이 성공적으로 변경되었습니다.", updatedTeam));
        } catch (Exception e) {
            return ResponseEntity.badRequest()
                    .body(MessageApiResponse.failure(e.getMessage()));
        }
    }

    // 팀 삭제 - softDeleteTeam 메서드 없음으로 하드 삭제 사용
    @DeleteMapping("/{teamId}")
    public ResponseEntity<MessageApiResponse> deleteTeam(@PathVariable UUID teamId,
                                                         @RequestParam UUID requesterId) {
        log.info("Request to delete team: {} by requester: {}", teamId, requesterId);

        try {
            // 팀장 권한 확인 후 삭제
            if (!teamService.isTeamLeader(teamId, requesterId)) {
                return ResponseEntity.badRequest()
                        .body(MessageApiResponse.failure("팀 삭제 권한이 없습니다. 팀장만 팀을 삭제할 수 있습니다."));
            }

            boolean deleted = teamService.disbandTeam(teamId, requesterId);
            if (deleted) {
                return ResponseEntity.ok(MessageApiResponse.success("팀이 성공적으로 삭제되었습니다."));
            } else {
                return ResponseEntity.badRequest()
                        .body(MessageApiResponse.failure("팀 삭제에 실패했습니다."));
            }
        } catch (Exception e) {
            return ResponseEntity.badRequest()
                    .body(MessageApiResponse.failure("팀 삭제에 실패했습니다: " + e.getMessage()));
        }
    }

    // 팀 해산 (팀 삭제, UserInfo는 그대로 유지)
    @DeleteMapping("/{teamId}/disband")
    public ResponseEntity<MessageApiResponse> disbandTeam(@PathVariable UUID teamId,
                                                          @RequestParam UUID requesterId) {
        log.info("Request to disband team: {} by requester: {}", teamId, requesterId);

        try {
            boolean disbanded = teamService.disbandTeam(teamId, requesterId);
            if (disbanded) {
                return ResponseEntity.ok(MessageApiResponse.success("팀이 성공적으로 해산되었습니다."));
            } else {
                return ResponseEntity.badRequest().body(MessageApiResponse.failure("팀 해산에 실패했습니다."));
            }
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest()
                    .body(MessageApiResponse.failure(e.getMessage()));
        }
    }
    // 팀에 멤버 추가 (팀장만 가능)
    @PostMapping("/{teamId}/members")
    public ResponseEntity<MessageApiResponse> addMemberToTeam(
            @PathVariable UUID teamId,
            @RequestParam UUID requesterId,
            @RequestParam UUID userId,
            @RequestParam(required = false, defaultValue = "팀원") String role) {

        log.info("Request to add user {} to team {} by requester {}", userId, teamId, requesterId);

        try {
            UserInfo addedMember = teamService.addMemberToTeam(teamId, requesterId, userId, role);
            return ResponseEntity.ok(MessageApiResponse.success("팀원이 성공적으로 추가되었습니다.", addedMember));
        } catch (Exception e) {
            return ResponseEntity.badRequest()
                    .body(MessageApiResponse.failure("팀원 추가에 실패했습니다: " + e.getMessage()));
        }
    }

    // 팀에서 멤버 제거 (팀장만 가능)
    @DeleteMapping("/{teamId}/members/{userId}")
    public ResponseEntity<MessageApiResponse> removeMemberFromTeam(
            @PathVariable UUID teamId,
            @PathVariable UUID userId,
            @RequestParam UUID requesterId) {

        log.info("Request to remove user {} from team {} by requester {}", userId, teamId, requesterId);

        try {
            boolean removed = teamService.removeMemberFromTeam(teamId, requesterId, userId);
            if (removed) {
                return ResponseEntity.ok(MessageApiResponse.success("팀원이 성공적으로 제거되었습니다."));
            } else {
                return ResponseEntity.badRequest()
                        .body(MessageApiResponse.failure("팀원 제거에 실패했습니다."));
            }
        } catch (Exception e) {
            return ResponseEntity.badRequest()
                    .body(MessageApiResponse.failure("팀원 제거에 실패했습니다: " + e.getMessage()));
        }
    }




    // === 통계 및 카운트 API들 ===

//    // 활성화된 팀 수 조회
//    @GetMapping("/count")
//    public ResponseEntity<MessageApiResponse> getActiveTeamCount() {
//        long count = teamService.getActiveTeamCount();
//        return ResponseEntity.ok(MessageApiResponse.success("활성화된 팀 수를 성공적으로 조회했습니다.", count));
//    }

//    // 그룹별 활성화된 팀 수 조회
//    @GetMapping("/group/{groupId}/count")
//    public ResponseEntity<MessageApiResponse> getActiveTeamCountByGroup(@PathVariable UUID groupId) {
//        long count = teamService.getActiveTeamCountByGroupId(groupId);
//        return ResponseEntity.ok(MessageApiResponse.success("그룹의 팀 수를 성공적으로 조회했습니다.", count));
//    }

//    // 팀장별 활성화된 팀 수 조회
//    @GetMapping("/leader/{leaderId}/count")
//    public ResponseEntity<MessageApiResponse> getActiveTeamCountByLeader(@PathVariable UUID leaderId) {
//        long count = teamService.getActiveTeamCountByLeaderId(leaderId);
//        return ResponseEntity.ok(MessageApiResponse.success("팀장의 팀 수를 성공적으로 조회했습니다.", count));
//    }

    // === 관리자용 API들 - 해당 메서드들이 없으므로 주석 처리 ===

    // 비활성화된 팀 조회 (관리자용) - getInactiveTeams 메서드 없음
    /*
    @GetMapping("/inactive")
    public ResponseEntity<MessageApiResponse> getInactiveTeams() {
        log.info("Request to get inactive teams");
        List<Team> teams = teamService.getInactiveTeams();
        return ResponseEntity.ok(MessageApiResponse.success("비활성화된 팀 목록을 성공적으로 조회했습니다.", teams));
    }
    */

    // 최근 생성된 팀 조회 - getRecentActiveTeams 메서드 없음
    /*
    @GetMapping("/recent")
    public ResponseEntity<MessageApiResponse> getRecentTeams(@RequestParam(defaultValue = "7") int days) {
        log.info("Request to get recent teams for {} days", days);
        List<Team> teams = teamService.getRecentActiveTeams(days);
        return ResponseEntity.ok(MessageApiResponse.success("최근 생성된 팀 목록을 성공적으로 조회했습니다.", teams));
    }
    */

    // 그룹의 모든 팀 비활성화 (관리자용) - softDeleteTeamsByGroupId 메서드 없음
    /*
    @DeleteMapping("/group/{groupId}/all")
    public ResponseEntity<MessageApiResponse> deleteAllTeamsByGroup(@PathVariable UUID groupId) {
        log.info("Request to delete all teams in group: {}", groupId);

        boolean deleted = teamService.softDeleteTeamsByGroupId(groupId);
        if (deleted) {
            return ResponseEntity.ok(MessageApiResponse.success("그룹의 모든 팀이 성공적으로 삭제되었습니다."));
        } else {
            return ResponseEntity.ok(MessageApiResponse.success("삭제할 팀이 없습니다."));
        }
    }
    */
}