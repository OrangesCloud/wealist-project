package OrangeCloud.UserRepo.service;

import OrangeCloud.UserRepo.dto.workspace.*;
import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.entity.UserProfile;
import OrangeCloud.UserRepo.entity.Workspace;
import OrangeCloud.UserRepo.entity.WorkspaceJoinRequest;
import OrangeCloud.UserRepo.entity.WorkspaceMember;
import OrangeCloud.UserRepo.exception.UserNotFoundException;
import OrangeCloud.UserRepo.repository.UserProfileRepository;
import OrangeCloud.UserRepo.repository.UserRepository;
import OrangeCloud.UserRepo.repository.WorkspaceJoinRequestRepository;
import OrangeCloud.UserRepo.repository.WorkspaceMemberRepository;
import OrangeCloud.UserRepo.repository.WorkspaceRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;



@Service
@RequiredArgsConstructor
@Transactional
@Slf4j
public class WorkspaceService {

    private final WorkspaceRepository workspaceRepository;
    private final WorkspaceMemberRepository workspaceMemberRepository;
    private final WorkspaceJoinRequestRepository workspaceJoinRequestRepository;
    private final UserRepository userRepository;
    private final UserProfileRepository userProfileRepository;
    private static final UUID DEFAULT_WORKSPACE_ID = UUID.fromString("00000000-0000-0000-0000-000000000000");
    // ============================================================================
    // Workspace 생성/수정/삭제
    // ============================================================================

    /**
     * 새로운 Workspace 생성 (생성자가 자동으로 OWNER)
     */
    public WorkspaceResponse createWorkspace(CreateWorkspaceRequest request, UUID creatorId) {
        log.info("Creating workspace: name={}, creator={}", request.getWorkspaceName(), creatorId);

        User creator = userRepository.findById(creatorId)
                .orElseThrow(() -> {
                    log.warn("User not found: {}", creatorId);
                    return new UserNotFoundException("사용자를 찾을 수 없습니다.");
                });

        Workspace workspace = Workspace.builder()
                .ownerId(creatorId)
                .workspaceName(request.getWorkspaceName())
                .workspaceDescription(request.getWorkspaceDescription())
                .isPublic(request.getIsPublic())
                .needApproved(true)
                .isActive(true)
                .build();

        Workspace savedWorkspace = workspaceRepository.save(workspace);
        log.info("Workspace created: workspaceId={}", savedWorkspace.getWorkspaceId());

        WorkspaceMember ownerMember = WorkspaceMember.builder()
                .workspaceId(savedWorkspace.getWorkspaceId())
                .userId(creatorId)
                .role(WorkspaceMember.WorkspaceRole.OWNER)
                .isDefault(true)
                .isActive(true)
                .build();

        workspaceMemberRepository.save(ownerMember);
        log.info("Creator added as OWNER: workspaceId={}, userId={}", savedWorkspace.getWorkspaceId(), creatorId);

        UserProfile creatorProfile = userProfileRepository.findByWorkspaceIdAndUserId(DEFAULT_WORKSPACE_ID, creatorId)
                .orElseThrow(() -> {
                    log.warn("Profile not found for user: {}", creatorId);
                    return new UserNotFoundException("프로필을 찾을 수 없습니다.");
                });

        return convertToWorkspaceResponse(savedWorkspace, creator, creatorProfile);
    }

    /**
     * 테스트용 또는 신규 가입 사용자를 위한 기본 Workspace 생성
     */
//    public UUID createDefaultWorkspaceForUser(User user) {
//        log.info("Creating default workspace for user: {}", user.getUserId());
//
//        String defaultWorkspaceName = user.getNickName() != null && !user.getNickName().isEmpty() ?
//                user.getNickName() + "'s Workspace" :
//                user.getEmail().split("@")[0] + "'s Workspace";
//
//        Workspace workspace = Workspace.builder()
//                .ownerId(user.getUserId())
//                .workspaceName(defaultWorkspaceName)
//                .workspaceDescription("Default personal workspace")
//                .isPublic(false) // 기본적으로 비공개
//                .needApproved(false) // 승인 필요 없음
//                .isActive(true)
//                .build();
//
//        Workspace savedWorkspace = workspaceRepository.save(workspace);
//        log.info("Default workspace created: workspaceId={}", savedWorkspace.getWorkspaceId());
//
//        WorkspaceMember ownerMember = WorkspaceMember.builder()
//                .workspaceId(savedWorkspace.getWorkspaceId())
//                .userId(user.getUserId())
//                .role(WorkspaceMember.WorkspaceRole.OWNER)
//                .isDefault(true)
//                .isActive(true)
//                .build();
//
//        workspaceMemberRepository.save(ownerMember);
//        log.info("User {} added as OWNER of default workspace {}", user.getUserId(), savedWorkspace.getWorkspaceId());
//
//        return savedWorkspace.getWorkspaceId();
//    }

    /**
     * Workspace 정보 수정 (OWNER만)
     */
    public WorkspaceResponse updateWorkspace(UUID workspaceId, UpdateWorkspaceRequest request, UUID requesterId) {
        log.info("Updating workspace: workspaceId={}, requester={}", workspaceId, requesterId);

        checkWorkspaceOwner(workspaceId, requesterId);

        Workspace workspace = workspaceRepository.findById(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace not found");
                });

        if (request.getWorkspaceName() != null && !request.getWorkspaceName().isEmpty()) {
            workspace.setWorkspaceName(request.getWorkspaceName());
        }

        if (request.getWorkspaceDescription() != null && !request.getWorkspaceDescription().isEmpty()) {
            workspace.setWorkspaceDescription(request.getWorkspaceDescription());
        }

        Workspace updated = workspaceRepository.save(workspace);

        WorkspaceMember ownerMember = workspaceMemberRepository.findOwnerByWorkspaceId(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace owner not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace owner not found");
                });

        User owner = userRepository.findById(ownerMember.getUserId())
                .orElseThrow(() -> new UserNotFoundException("사용자를 찾을 수 없습니다."));

        UserProfile ownerProfile = userProfileRepository.findByUserId(ownerMember.getUserId())
                .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));

        log.info("Workspace updated: workspaceId={}", workspaceId);
        return convertToWorkspaceResponse(updated, owner, ownerProfile);
    }

    /**
     * Workspace 소프트 삭제 (OWNER만)
     */
    public void deleteWorkspace(UUID workspaceId, UUID requesterId) {
        log.info("Deleting workspace: workspaceId={}, requester={}", workspaceId, requesterId);

        checkWorkspaceOwner(workspaceId, requesterId);

        Workspace workspace = workspaceRepository.findById(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace not found");
                });

        workspace.softDelete();
        workspaceRepository.save(workspace);
        log.info("Workspace deleted: workspaceId={}", workspaceId);
    }

    // ============================================================================
    // Workspace 설정 관리 (신규 추가)
    // ============================================================================

    /**
     * 특정 Workspace의 설정 정보를 조회합니다. (멤버만 가능)
     */
    @Transactional(readOnly = true)
    public WorkspaceSettingsResponse getWorkspaceSettings(UUID workspaceId, UUID requesterId) {
        log.debug("Fetching workspace settings: workspaceId={}, requester={}", workspaceId, requesterId);

        checkWorkspaceMember(workspaceId, requesterId);

        Workspace workspace = workspaceRepository.findById(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace not found");
                });

        return convertToWorkspaceSettingsResponse(workspace);
    }

    /**
     * Workspace 설정 정보를 수정합니다. (OWNER/ADMIN만 가능)
     */
    public WorkspaceSettingsResponse updateWorkspaceSettings(
            UUID workspaceId,
            UpdateWorkspaceSettingsRequest request,
            UUID requesterId) {
        log.info("Updating workspace settings: workspaceId={}, requester={}", workspaceId, requesterId);

        checkWorkspaceAdminOrOwner(workspaceId, requesterId);

        Workspace workspace = workspaceRepository.findById(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace not found");
                });

        if (request.getWorkspaceName() != null && !request.getWorkspaceName().isEmpty()) {
            workspace.setWorkspaceName(request.getWorkspaceName());
        }
        if (request.getWorkspaceDescription() != null) {
            workspace.setWorkspaceDescription(request.getWorkspaceDescription());
        }
        if (request.getIsPublic() != null) {
            workspace.setIsPublic(request.getIsPublic());
        }
        if (request.getNeedApproved() != null) {
            workspace.setNeedApproved(request.getNeedApproved());
        }

        Workspace updated = workspaceRepository.save(workspace);
        log.info("Workspace settings updated: workspaceId={}", workspaceId);
        return convertToWorkspaceSettingsResponse(updated);
    }

    // ============================================================================
    // Workspace 조회
    // ============================================================================

    /**
     * Workspace 조회 (멤버만 가능)
     */
    @Transactional(readOnly = true)
    public WorkspaceResponse getWorkspace(UUID workspaceId, UUID requesterId) {
        log.debug("Fetching workspace: workspaceId={}", workspaceId);

        checkWorkspaceMember(workspaceId, requesterId);

        Workspace workspace = workspaceRepository.findById(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace not found");
                });

        WorkspaceMember ownerMember = workspaceMemberRepository.findOwnerByWorkspaceId(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace owner not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace owner not found");
                });

        User owner = userRepository.findById(ownerMember.getUserId())
                .orElseThrow(() -> new UserNotFoundException("사용자를 찾을 수 없습니다."));

        UserProfile ownerProfile = userProfileRepository.findByUserId(ownerMember.getUserId())
                .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));

        return convertToWorkspaceResponse(workspace, owner, ownerProfile);
    }
    @Transactional(readOnly = true)
    public WorkspaceResponse getWorkspaceByQuery(String workspaceName,UUID workspaceId, UUID requesterId) {
        log.debug("Fetching workspace: workspaceId={}", workspaceId);

        checkWorkspaceMember(workspaceId, requesterId);

        Workspace workspace = workspaceRepository.findById(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace not found");
                });

        WorkspaceMember ownerMember = workspaceMemberRepository.findOwnerByWorkspaceId(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace owner not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace owner not found");
                });

        User owner = userRepository.findById(ownerMember.getUserId())
                .orElseThrow(() -> new UserNotFoundException("사용자를 찾을 수 없습니다."));

        UserProfile ownerProfile = userProfileRepository.findByUserId(ownerMember.getUserId())
                .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));

        return convertToWorkspaceResponse(workspace, owner, ownerProfile);
    }

    /**
     * 사용자가 속한 모든 Workspace 조회
     */
    @Transactional(readOnly = true)
    public List<WorkspaceResponse> getUserWorkspaces(UUID userId) {
        log.debug("Fetching workspaces for user: userId={}", userId);

        List<WorkspaceMember> members = workspaceMemberRepository.findActiveByUserId(userId);

        return members.stream()
                .map(member -> {
                    Workspace workspace = workspaceRepository.findById(member.getWorkspaceId())
                            .orElseThrow(() -> new IllegalArgumentException("Workspace not found"));
                    WorkspaceMember owner = workspaceMemberRepository.findOwnerByWorkspaceId(member.getWorkspaceId())
                            .orElseThrow(() -> new IllegalArgumentException("Workspace owner not found"));
                    User ownerUser = userRepository.findById(owner.getUserId())
                            .orElseThrow(() -> new UserNotFoundException("사용자를 찾을 수 없습니다."));
                    UserProfile ownerProfile = userProfileRepository.findByWorkspaceIdAndUserId(DEFAULT_WORKSPACE_ID, ownerUser.getUserId())
                            .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));
                    return convertToWorkspaceResponse(workspace, ownerUser, ownerProfile);
                })
                .collect(Collectors.toList());
    }


    @Transactional(readOnly = true)
    public List<WorkspaceResponse> SearchgetUserWorkspaces(UUID userId) {
        log.debug("Fetching workspaces for user: userId={}", userId);

        List<WorkspaceMember> members = workspaceMemberRepository.findActiveByUserId(userId);

        return members.stream()
                .map(member -> {
                    Workspace workspace = workspaceRepository.findById(member.getWorkspaceId())
                            .orElseThrow(() -> new IllegalArgumentException("Workspace not found"));
                    WorkspaceMember owner = workspaceMemberRepository.findOwnerByWorkspaceId(member.getWorkspaceId())
                            .orElseThrow(() -> new IllegalArgumentException("Workspace owner not found"));
                    User ownerUser = userRepository.findById(owner.getUserId())
                            .orElseThrow(() -> new UserNotFoundException("사용자를 찾을 수 없습니다."));
                    UserProfile ownerProfile = userProfileRepository.findByWorkspaceIdAndUserId(DEFAULT_WORKSPACE_ID, ownerUser.getUserId())
                            .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));
                    return convertToWorkspaceResponse(workspace, ownerUser, ownerProfile);
                })
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public List<WorkspaceResponse> searchPublicWorkspaces(String query) {
        log.debug("Searching public workspaces with query: {}", query);

        List<Workspace> publicWorkspaces;

        if (query != null && !query.trim().isEmpty()) {
            // 쿼리가 있는 경우 이름으로 검색
            publicWorkspaces = workspaceRepository.findPublicWorkspacesByNameContaining(query.trim());
        } else {
            // 쿼리가 없는 경우 모든 공개 워크스페이스 조회
            publicWorkspaces = workspaceRepository.findAllPublicWorkspaces();
        }

        return publicWorkspaces.stream()
                .map(workspace -> {
                    WorkspaceMember owner = workspaceMemberRepository.findOwnerByWorkspaceId(workspace.getWorkspaceId())
                            .orElseThrow(() -> new IllegalArgumentException("Workspace owner not found"));
                    User ownerUser = userRepository.findById(owner.getUserId())
                            .orElseThrow(() -> new UserNotFoundException("사용자를 찾을 수 없습니다."));
                    UserProfile ownerProfile = userProfileRepository.findByWorkspaceIdAndUserId(DEFAULT_WORKSPACE_ID, ownerUser.getUserId())
                            .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));
                    return convertToWorkspaceResponse(workspace, ownerUser, ownerProfile);
                })
                .collect(Collectors.toList());
    }

    /**
     * 기본 Workspace 설정
     */
    public void setDefaultWorkspace(UUID workspaceId, UUID userId) {
        log.info("Setting default workspace: workspaceId={}, userId={}", workspaceId, userId);

        checkWorkspaceMember(workspaceId, userId);

        List<WorkspaceMember> userMembers = workspaceMemberRepository.findActiveByUserId(userId);
        userMembers.forEach(member -> {
            member.setDefault(false);
            workspaceMemberRepository.save(member);
        });

        WorkspaceMember member = workspaceMemberRepository.findByWorkspaceIdAndUserId(workspaceId, userId)
                .orElseThrow(() -> {
                    log.warn("User is not a member of workspace: workspaceId={}, userId={}", workspaceId, userId);
                    return new IllegalArgumentException("User is not a member of this workspace");
                });

        member.setDefault(true);
        workspaceMemberRepository.save(member);
        log.info("Default workspace set: workspaceId={}", workspaceId);
    }

    // ============================================================================
    // Workspace 멤버 관리
    // ============================================================================

    /**
     * 워크스페이스에 사용자 초대
     */
    @Transactional
    public WorkspaceMemberResponse inviteUser(UUID workspaceId, InviteUserRequest request, UUID requesterId) {
        String query = request.getQuery().trim();
        log.info("Inviting user to workspace: workspaceId={}, query={}, requester={}", workspaceId, query, requesterId);

        checkWorkspaceAdminOrOwner(workspaceId, requesterId);

        // ✅ 이메일 형식인지 검사
        boolean isEmail = query.contains("@");

        // ✅ 사용자 조회
        Optional<User> targetUserOpt;

        targetUserOpt = userRepository.findByEmail(query);


        User targetUser = targetUserOpt
                .orElseThrow(() -> new UserNotFoundException("해당 " + (isEmail ? "이메일" : "이름") + "의 사용자를 찾을 수 없습니다."));

        // ✅ 이미 멤버인지 확인
        if (workspaceMemberRepository.existsByWorkspaceIdAndUserId(workspaceId, targetUser.getUserId())) {
            log.warn("User is already a member of workspace: workspaceId={}, userId={}", workspaceId, targetUser.getUserId());
            throw new IllegalArgumentException("해당 사용자는 이미 워크스페이스 멤버입니다.");
        }

        // ✅ 새 멤버 등록
        WorkspaceMember newMember = WorkspaceMember.builder()
                .workspaceId(workspaceId)
                .userId(targetUser.getUserId())
                .role(WorkspaceMember.WorkspaceRole.MEMBER)
                .isDefault(false)
                .isActive(true)
                .build();

        WorkspaceMember savedMember = workspaceMemberRepository.save(newMember);
        log.info("User invited and added as member: workspaceId={}, userId={}", workspaceId, targetUser.getUserId());

        return convertToWorkspaceMemberResponse(savedMember);
    }



    /**
     * 특정 워크스페이스의 모든 멤버 목록을 조회합니다 (UserProfile 포함).
     */
    @Transactional(readOnly = true)
    public List<WorkspaceMemberResponse> getWorkspaceMembers(UUID workspaceId, UUID currentUserId) {
        // 권한 확인 (멤버인지 확인) - 생략 가능하나 안정성을 위해 checkWorkspaceMember(workspaceId,
        // currentUserId);

        List<WorkspaceMember> members = workspaceMemberRepository.findAllByWorkspaceId(workspaceId);

        return members.stream()
                .<WorkspaceMemberResponse>map(member -> {
                    Optional<User> userOpt = userRepository.findById(member.getUserId());
                    Optional<UserProfile> profileOpt = userProfileRepository.findByUserId(member.getUserId());

                    User user = userOpt.orElseGet(
                            () -> User.builder().userId(member.getUserId()).email("unknown@user.com").build());
                    UserProfile profile = profileOpt.orElseGet(
                            () -> UserProfile.builder().nickName("Deleted User").userId(member.getUserId()).build());

                    return convertToWorkspaceMemberResponse(
                            member,
                            profile.getNickName(),
                            profile.getProfileImageUrl(),
                            user);
                })
                .collect(Collectors.toList());
    }

    /**
     * 멤버 역할 변경 (OWNER만)
     */
    public WorkspaceMemberResponse updateMemberRole(UUID workspaceId, UUID memberId, UpdateMemberRoleRequest request,
                                                    UUID requesterId) {
        log.info("Updating member role: workspaceId={}, memberId={}, newRole={}, requester={}",
                workspaceId, memberId, request.getRoleName(), requesterId);

        checkWorkspaceOwner(workspaceId, requesterId);

        WorkspaceMember member = workspaceMemberRepository.findById(memberId)
                .orElseThrow(() -> {
                    log.warn("Member not found: {}", memberId);
                    return new IllegalArgumentException("Member not found");
                });

        if (!member.getWorkspaceId().equals(workspaceId)) {
            log.warn("Member does not belong to workspace: memberId={}, workspaceId={}", memberId, workspaceId);
            throw new IllegalArgumentException("Member does not belong to this workspace");
        }

        member.setRole(WorkspaceMember.WorkspaceRole.valueOf(request.getRoleName()));
        WorkspaceMember updated = workspaceMemberRepository.save(member);

        return convertToWorkspaceMemberResponse(updated);
    }

    /**
     * 멤버 제거 (OWNER/ADMIN만, OWNER는 제거 불가)
     */
    public void removeMember(UUID workspaceId, UUID memberId, UUID requesterId) {
        log.info("Removing member: workspaceId={}, memberId={}, requester={}", workspaceId, memberId, requesterId);

        checkWorkspaceAdminOrOwner(workspaceId, requesterId);

        WorkspaceMember member = workspaceMemberRepository.findById(memberId)
                .orElseThrow(() -> {
                    log.warn("Member not found: {}", memberId);
                    return new IllegalArgumentException("Member not found");
                });

        if (!member.getWorkspaceId().equals(workspaceId)) {
            log.warn("Member does not belong to workspace: memberId={}, workspaceId={}", memberId, workspaceId);
            throw new IllegalArgumentException("Member does not belong to this workspace");
        }

        if (member.getRole() == WorkspaceMember.WorkspaceRole.OWNER) {
            log.warn("Cannot remove workspace owner: memberId={}", memberId);
            throw new IllegalArgumentException("Cannot remove workspace owner");
        }

        // requesterId가 memberId와 일치하는지 확인 (스스로를 제거하는 행위 방지)
        if (member.getUserId().equals(requesterId)) {
            log.warn("User cannot remove themselves: userId={}", requesterId);
            throw new IllegalArgumentException("Cannot remove yourself");
        }

        member.setIsActive(false);
        workspaceMemberRepository.save(member);
        log.info("Member removed: workspaceId={}, memberId={}", workspaceId, memberId);
    }

    // ============================================================================
    // 가입 신청 관리
    // ============================================================================

    /**
     * Workspace 가입 신청
     */
    public JoinRequestResponse createJoinRequest(UUID workspaceId, UUID userId) {
        log.info("Creating join request: workspaceId={}, userId={}", workspaceId, userId);

        User user = userRepository.findById(userId)
                .orElseThrow(() -> {
                    log.warn("User not found: {}", userId);
                    return new UserNotFoundException("사용자를 찾을 수 없습니다.");
                });

        if (workspaceMemberRepository.existsByWorkspaceIdAndUserId(workspaceId, userId)) {
            log.warn("User is already a member of workspace: workspaceId={}, userId={}", workspaceId, userId);
            throw new IllegalArgumentException("User is already a member of this workspace");
        }

        // TODO: 이미 PENDING 상태의 요청이 있는지 확인하는 로직 추가 필요

        WorkspaceJoinRequest request = WorkspaceJoinRequest.builder()
                .workspaceId(workspaceId)
                .userId(userId)
                .status(WorkspaceJoinRequest.JoinRequestStatus.PENDING)
                .build();

        WorkspaceJoinRequest saved = workspaceJoinRequestRepository.save(request);
        log.info("Join request created: requestId={}", saved.getJoinRequestId());

        UserProfile userProfile = userProfileRepository.findByUserId(userId)
                .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));

        return convertToJoinRequestResponse(saved, user, userProfile);
    }

    /**
     * 가입 신청 승인 (POST /{userId}/approve)
     */
    public void approveJoinRequest(UUID workspaceId, UUID userId, UUID responderId) {
        log.info("Approving join request: workspaceId={}, userId={}, responder={}", workspaceId, userId, responderId);

        checkWorkspaceAdminOrOwner(workspaceId, responderId);

        // ✅ 개선된 코드: Repository 메서드를 사용하여 DB 레벨에서 정확히 조회
        Optional<WorkspaceJoinRequest> pendingRequestOpt = workspaceJoinRequestRepository
                .findByWorkspaceIdAndUserIdAndStatus(
                        workspaceId,
                        userId,
                        WorkspaceJoinRequest.JoinRequestStatus.PENDING);
        WorkspaceJoinRequest joinRequest = pendingRequestOpt.orElseThrow(() -> {
            log.warn("Pending join request not found for user: workspaceId={}, userId={}", workspaceId, userId);
            throw new IllegalArgumentException("Pending join request not found");
        });

        // 멤버 추가
        WorkspaceMember newMember = WorkspaceMember.builder()
                .workspaceId(workspaceId)
                .userId(userId)
                .role(WorkspaceMember.WorkspaceRole.MEMBER)
                .isDefault(false)
                .isActive(true)
                .build();
        workspaceMemberRepository.save(newMember);
        log.info("User approved and added as member: workspaceId={}, userId={}", workspaceId, userId);

        // 신청 상태 업데이트
        joinRequest.setStatus(WorkspaceJoinRequest.JoinRequestStatus.APPROVED);
        workspaceJoinRequestRepository.save(joinRequest);
    }

    /**
     * 가입 신청 거절 (POST /{userId}/reject)
     */
    public void rejectJoinRequest(UUID workspaceId, UUID userId, UUID responderId) {
        log.info("Rejecting join request: workspaceId={}, userId={}, responder={}", workspaceId, userId, responderId);

        checkWorkspaceAdminOrOwner(workspaceId, responderId);

        // ✅ 개선된 코드: Repository 메서드를 사용하여 DB 레벨에서 정확히 조회
        Optional<WorkspaceJoinRequest> pendingRequestOpt = workspaceJoinRequestRepository
                .findByWorkspaceIdAndUserIdAndStatus(
                        workspaceId,
                        userId,
                        WorkspaceJoinRequest.JoinRequestStatus.PENDING);
        WorkspaceJoinRequest joinRequest = pendingRequestOpt.orElseThrow(() -> {
            log.warn("Pending join request not found for user: workspaceId={}, userId={}", workspaceId, userId);
            throw new IllegalArgumentException("Pending join request not found");
        });

        // 신청 상태 업데이트 (거절)
        joinRequest.setStatus(WorkspaceJoinRequest.JoinRequestStatus.REJECTED);
        workspaceJoinRequestRepository.save(joinRequest);
        log.info("Join request rejected: workspaceId={}, userId={}", workspaceId, userId);
    }

    /**
     * 가입 신청 승인/거절 (기존 PUT /joinRequests/{requestId} 매핑)
     */
    public JoinRequestResponse updateJoinRequest(UUID workspaceId, UUID requestId,
                                                 UpdateJoinRequestRequest updateRequest, UUID responderId) {
        log.info("Updating join request: workspaceId={}, requestId={}, status={}, responder={}",
                workspaceId, requestId, updateRequest.getStatus(), responderId);

        checkWorkspaceAdminOrOwner(workspaceId, responderId);

        WorkspaceJoinRequest joinRequest = workspaceJoinRequestRepository.findById(requestId)
                .orElseThrow(() -> {
                    log.warn("Join request not found: {}", requestId);
                    return new IllegalArgumentException("Join request not found");
                });

        if (!joinRequest.getWorkspaceId().equals(workspaceId)) {
            log.warn("Join request does not belong to workspace: requestId={}, workspaceId={}", requestId, workspaceId);
            throw new IllegalArgumentException("Join request does not belong to this workspace");
        }

        WorkspaceJoinRequest.JoinRequestStatus newStatus = WorkspaceJoinRequest.JoinRequestStatus
                .valueOf(updateRequest.getStatus());

        if (newStatus == WorkspaceJoinRequest.JoinRequestStatus.APPROVED) {
            // 멤버 추가 로직
            WorkspaceMember newMember = WorkspaceMember.builder()
                    .workspaceId(workspaceId)
                    .userId(joinRequest.getUserId())
                    .role(WorkspaceMember.WorkspaceRole.MEMBER)
                    .isDefault(false)
                    .isActive(true)
                    .build();
            workspaceMemberRepository.save(newMember);
            log.info("User approved and added as member: workspaceId={}, userId={}", workspaceId,
                    joinRequest.getUserId());
        }

        joinRequest.setStatus(newStatus);
        WorkspaceJoinRequest updated = workspaceJoinRequestRepository.save(joinRequest);

        User user = userRepository.findById(joinRequest.getUserId())
                .orElseThrow(() -> new UserNotFoundException("사용자를 찾을 수 없습니다."));

        UserProfile userProfile = userProfileRepository.findByUserId(joinRequest.getUserId())
                .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));

        return convertToJoinRequestResponse(updated, user, userProfile);
    }

    /**
     * Workspace의 가입 신청 목록 조회 (OWNER/ADMIN만)
     */
    @Transactional(readOnly = true)
    public List<JoinRequestResponse> getJoinRequests(UUID workspaceId, UUID requesterId, String status) {
        log.debug("Fetching join requests: workspaceId={}, status={}", workspaceId, status);

        checkWorkspaceAdminOrOwner(workspaceId, requesterId);

        List<WorkspaceJoinRequest> requests;
        if (status != null && !status.isEmpty()) {
            // TODO: Repository에 findByWorkspaceIdAndStatus 메서드 필요
            requests = workspaceJoinRequestRepository.findByWorkspaceId(workspaceId);

        } else {
            requests = workspaceJoinRequestRepository.findByWorkspaceId(workspaceId);
        }

        return requests.stream()
                .map(req -> {
                    User user = userRepository.findById(req.getUserId())
                            .orElseThrow(() -> new UserNotFoundException("사용자를 찾을 수 없습니다."));
                    UserProfile userProfile = userProfileRepository.findByUserId(req.getUserId())
                            .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));
                    return convertToJoinRequestResponse(req, user, userProfile);
                })
                .collect(Collectors.toList());
    }




    // ============================================================================
    // 권한 확인 (Private Methods)
    // ============================================================================

    /**
     * Workspace 멤버 확인
     */
    private void checkWorkspaceMember(UUID workspaceId, UUID userId) {
        if (!workspaceMemberRepository.existsByWorkspaceIdAndUserId(workspaceId, userId)) {
            log.warn("User is not a member of workspace: workspaceId={}, userId={}", workspaceId, userId);
            throw new IllegalArgumentException("User is not a member of this workspace");
        }
    }

    /**
     * Workspace OWNER 확인
     */
    private void checkWorkspaceOwner(UUID workspaceId, UUID userId) {
        WorkspaceMember member = workspaceMemberRepository.findByWorkspaceIdAndUserId(workspaceId, userId)
                .orElseThrow(() -> {
                    log.warn("User is not a member of workspace: workspaceId={}, userId={}", workspaceId, userId);
                    return new IllegalArgumentException("User is not a member of this workspace");
                });

        if (member.getRole() != WorkspaceMember.WorkspaceRole.OWNER) {
            log.warn("User is not OWNER of workspace: workspaceId={}, userId={}", workspaceId, userId);
            throw new IllegalArgumentException("Only workspace owner can perform this action");
        }
    }

    /**
     * Workspace OWNER 또는 ADMIN 확인
     */
    private void checkWorkspaceAdminOrOwner(UUID workspaceId, UUID userId) {
        WorkspaceMember member = workspaceMemberRepository.findByWorkspaceIdAndUserId(workspaceId, userId)
                .orElseThrow(() -> {
                    log.warn("User is not a member of workspace: workspaceId={}, userId={}", workspaceId, userId);
                    return new IllegalArgumentException("User is not a member of this workspace");
                });

        if (member.getRole() != WorkspaceMember.WorkspaceRole.OWNER &&
                member.getRole() != WorkspaceMember.WorkspaceRole.ADMIN) {
            log.warn("User is not OWNER or ADMIN of workspace: workspaceId={}, userId={}", workspaceId, userId);
            throw new IllegalArgumentException("Only workspace owner or admin can perform this action");
        }
    }

    // ============================================================================
    // DTO 변환 (Private Methods)
    // ============================================================================

    private WorkspaceResponse convertToWorkspaceResponse(Workspace workspace, User owner, UserProfile ownerProfile) {
        return WorkspaceResponse.builder()
                .workspaceId(workspace.getWorkspaceId())
                .workspaceName(workspace.getWorkspaceName())
                .workspaceDescription(workspace.getWorkspaceDescription())
                .ownerId(owner.getUserId())
                .ownerName(ownerProfile.getNickName())
                .ownerEmail(owner.getEmail())
                .isPublic(workspace.getIsPublic())
                .needApproved(workspace.getNeedApproved())
                .createdAt(workspace.getCreatedAt())
                .build();
    }

    /**
     * WorkspaceSettingsResponse DTO로 변환
     */
    private WorkspaceSettingsResponse convertToWorkspaceSettingsResponse(Workspace workspace) {
        return WorkspaceSettingsResponse.builder()
                .workspaceId(workspace.getWorkspaceId())
                .workspaceName(workspace.getWorkspaceName())
                .workspaceDescription(workspace.getWorkspaceDescription())
                .isPublic(workspace.getIsPublic())
                .requiresApproval(workspace.getNeedApproved()) // BE needApproved -> FE requiresApproval 매핑
                .onlyOwnerCanInvite(false) // BE 엔티티에 필드 없음. 임시값
                .build();
    }

    private WorkspaceMemberResponse convertToWorkspaceMemberResponse(WorkspaceMember member, String userName,
                                                                     String profileImageUrl, User user) {
        return WorkspaceMemberResponse.builder()
                .id(member.getId())
                .workspaceId(member.getWorkspaceId())
                .userId(member.getUserId())
                .userName(userName)
                .userEmail(user.getEmail())
                .profileImageUrl(profileImageUrl)
                .roleName(member.getRole().name())
                .isDefault(member.isDefault())
                .joinedAt(member.getJoinedAt())
                .build();
    }

    private WorkspaceMemberResponse convertToWorkspaceMemberResponse(WorkspaceMember member) {
        User user = userRepository.findById(member.getUserId())
                .orElse(null);

        UserProfile userProfile = userProfileRepository.findByUserId(member.getUserId())
                .orElse(null);

        String userName = userProfile != null ? userProfile.getNickName() : "Deleted/Missing User";
        String userEmail = user != null ? user.getEmail() : "missing@user.com";
        String profileImageUrl = userProfile != null ? userProfile.getProfileImageUrl() : null;

        return WorkspaceMemberResponse.builder()
                .id(member.getId())
                .workspaceId(member.getWorkspaceId())
                .userId(member.getUserId())
                .userName(userName)
                .userEmail(userEmail)
                .profileImageUrl(profileImageUrl)
                .roleName(member.getRole().name())
                .isDefault(member.isDefault())
                .joinedAt(member.getJoinedAt())
                .build();
    }

    private JoinRequestResponse convertToJoinRequestResponse(WorkspaceJoinRequest request, User user,
                                                             UserProfile userProfile) {
        return JoinRequestResponse.builder()
                .id(request.getJoinRequestId())
                .workspaceId(request.getWorkspaceId())
                .userId(request.getUserId())
                .userName(userProfile.getNickName())
                .userEmail(user.getEmail())
                .status(request.getStatus().name())
                .requestedAt(request.getRequestedAt())
                .updatedAt(request.getUpdatedAt())
                .build();
    }
}