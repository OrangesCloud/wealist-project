//package com.wealist.user.config;
//
//import OrangeCloud.UserRepo.entity.User;
//import OrangeCloud.UserRepo.entity.UserProfile;
//import OrangeCloud.UserRepo.entity.Workspace;
//import OrangeCloud.UserRepo.entity.WorkspaceMember;
//import OrangeCloud.UserRepo.repository.UserProfileRepository;
//import OrangeCloud.UserRepo.repository.UserRepository;
//import OrangeCloud.UserRepo.repository.WorkspaceMemberRepository;
//import OrangeCloud.UserRepo.repository.WorkspaceRepository;
//import lombok.RequiredArgsConstructor;
//import lombok.extern.slf4j.Slf4j;
//
//import org.springframework.boot.CommandLineRunner;
//import org.springframework.context.annotation.Bean;
//import org.springframework.context.annotation.Configuration;
//import org.springframework.core.env.Environment;
//
//import java.util.*;
//
//import static java.rmi.server.LogStream.log;
//
//@Slf4j
//@Configuration
//@RequiredArgsConstructor
//public class TestDataInitializer {
//
//    private final UserRepository userRepository;
//    private final UserProfileRepository userProfileRepository;
//    private final WorkspaceRepository workspaceRepository;
//    private final WorkspaceMemberRepository workspaceMemberRepository;
//    private final Environment environment;
//
//    private static final int WORKSPACE_COUNT = 10;
//    private static final int USER_COUNT = 50;
//    private static final int USERS_PER_WORKSPACE = 5;
//
//    @Bean
//    public CommandLineRunner initializeTestData() {
//        return args -> {
//            String[] activeProfiles = environment.getActiveProfiles();
//            boolean isTestEnvironment = Arrays.asList(activeProfiles).contains("test")
//                    || activeProfiles.length == 0;
//
//            if (!isTestEnvironment) {
//                TestDataInitializer.log.info("⏭️  Not a test environment. Skipping test data initialization.");
//                return;
//            }
//
//            TestDataInitializer.log.info("🚀 Starting test data initialization...");
//
//            if (userRepository.count() > 0 || workspaceRepository.count() > 0) {
//                TestDataInitializer.log.info("✅ Database already contains data. Skipping test data creation.");
//                return;
//            }
//
//            // 1️⃣ 기본 owner 유저 생성
//            User owner = createInitialUser();
//
//            // 2️⃣ 워크스페이스 생성 및 owner 할당
//            List<Workspace> workspaces = createWorkspaces(owner);
//
//            // 3️⃣ 테스트 유저 생성
//            List<User> users = createTestUsers(USER_COUNT);
//
//            // 4️⃣ 생성된 유저를 워크스페이스에 배정 (유저프로필 + 워크스페이스멤버)
//            assignUsersToWorkspaces(users, workspaces);
//
//            TestDataInitializer.log.info("🎉 Test data initialization completed! Created {} users and {} workspaces.",
//                    userRepository.count(), workspaceRepository.count());
//        };
//    }
//
//    private User createInitialUser() {
//        return userRepository.findByEmail("owner@wealist.com")
//                .orElseGet(() -> userRepository.save(
//                        User.builder()
//                                .email("owner@wealist.com")
//                                .provider("google")
//                                .googleId("google-id-owner")
//                                .isActive(true)
//                                .build()
//                ));
//    }
//
//    private List<User> createTestUsers(int count) {
//        List<User> users = new ArrayList<>();
//        for (int i = 1; i <= count; i++) {
//            String email = "testuser" + i + "@wealist.com";
//            if (userRepository.existsByEmailAndIsActiveTrue(email)) continue;
//
//            User user = User.builder()
//                    .email(email)
//                    .provider("google")
//                    .googleId("google-id-" + i)
//                    .isActive(true)
//                    .build();
//            users.add(userRepository.save(user));
//        }
//        return users;
//    }
//
//    private List<Workspace> createWorkspaces(User owner) {
//        List<Workspace> workspaces = new ArrayList<>();
//        for (int i = 1; i <= WORKSPACE_COUNT; i++) {
//            Workspace workspace = Workspace.builder()
//                    .ownerId(owner.getUserId())
//                    .workspaceName("Test Workspace " + i)
//                    .workspaceDescription("Auto-generated workspace number " + i)
//                    .isPublic(i % 2 == 0)
//                    .needApproved(i % 3 != 0)
//                    .build();
//            Workspace savedWorkspace = workspaceRepository.save(workspace);
//            workspaces.add(savedWorkspace);
//
//            // 워크스페이스 생성 시 Owner를 멤버로 자동 등록
//            // 첫 번째 워크스페이스를 기본값으로 설정
//            boolean isDefaultForOwner = (i == 1);
//            createWorkspaceMember(owner, savedWorkspace, WorkspaceMember.WorkspaceRole.OWNER, isDefaultForOwner);
//            createProfileIfNotExists(owner, savedWorkspace);
//        }
//        return workspaces;
//    }
//
//    private void assignUsersToWorkspaces(List<User> users, List<Workspace> workspaces) {
//        if (workspaces.isEmpty()) return;
//
//        int workspaceIndex = 0;
//
//        for (int i = 0; i < users.size(); i += USERS_PER_WORKSPACE) {
//            Workspace workspace = workspaces.get(workspaceIndex);
//
//            // USERS_PER_WORKSPACE 만큼 잘라서 해당 워크스페이스에 배정
//            int end = Math.min(i + USERS_PER_WORKSPACE, users.size());
//            List<User> subList = users.subList(i, end);
//
//            for (User user : subList) {
//                boolean hasDefaultWorkspace = workspaceMemberRepository.existsByUserIdAndIsDefaultTrue(user.getUserId());
//                createProfileIfNotExists(user, workspace);
//                createWorkspaceMember(user, workspace, WorkspaceMember.WorkspaceRole.MEMBER, !hasDefaultWorkspace);
//            }
//
//            // 다음 워크스페이스로 이동
//            workspaceIndex++;
//            if (workspaceIndex >= workspaces.size()) {
//                workspaceIndex = 0; // 워크스페이스 부족 시 다시 처음으로 순환
//            }
//        }
//    }
//
//    private void createProfileIfNotExists(User user, Workspace workspace) {
//        if (userProfileRepository.existsByUserIdAndWorkspaceId(user.getUserId(), workspace.getWorkspaceId())) {
//            return;
//        }
//        UserProfile profile = UserProfile.builder()
//                .userId(user.getUserId())
//                .workspaceId(workspace.getWorkspaceId())
//                .nickName(user.getEmail().split("@")[0])
//                .email(user.getEmail())
//                .profileImageUrl("https://i.pravatar.cc/150?u=" + user.getUserId())
//                .build();
//        userProfileRepository.save(profile);
//    }
//
//    private void createWorkspaceMember(User user, Workspace workspace, WorkspaceMember.WorkspaceRole role, boolean isDefault) {
//        if (workspaceMemberRepository.existsByWorkspaceIdAndUserId(workspace.getWorkspaceId(), user.getUserId())) {
//            return;
//        }
//        WorkspaceMember member = WorkspaceMember.builder()
//                .workspaceId(workspace.getWorkspaceId())
//                .userId(user.getUserId())
//                .role(role)
//                .isDefault(isDefault)
//                .isActive(true)
//                .build();
//        workspaceMemberRepository.save(member);
//        log.info("Test workSpaceMember 잘 저장중");
//    }
//}
