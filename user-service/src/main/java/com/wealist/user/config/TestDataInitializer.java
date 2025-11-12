package com.wealist.user.config;

import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.entity.UserProfile;
import OrangeCloud.UserRepo.entity.Workspace;
import OrangeCloud.UserRepo.repository.UserProfileRepository;
import OrangeCloud.UserRepo.repository.UserRepository;
import OrangeCloud.UserRepo.repository.WorkspaceRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.CommandLineRunner;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.env.Environment;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.UUID;

@Slf4j
@Configuration
@RequiredArgsConstructor
public class TestDataInitializer {

    private final UserRepository userRepository;
    private final UserProfileRepository userProfileRepository;
    private final WorkspaceRepository workspaceRepository;
    private final Environment environment;

    @Bean
    public CommandLineRunner initializeTestData() {
        return args -> {
            String[] activeProfiles = environment.getActiveProfiles();
            boolean isTestEnvironment = Arrays.asList(activeProfiles).contains("test")
                    || activeProfiles.length == 0;

            if (!isTestEnvironment) {
                log.info("⏭️  Not a test environment. Skipping test data initialization.");
                return;
            }

            if (userRepository.count() > 0) {
                log.info("✅ Database already has users. Skipping initialization.");
                return;
            }

            log.info("🚀 Starting test data initialization...");

            // 1. Create a default workspace
            User owner = createInitialUser();
            Workspace workspace = createDefaultWorkspace(owner);
            log.info("✅ Created default workspace: {}", workspace.getWorkspaceName());

            // 2. Create test users
            List<User> users = createTestUsers(owner);
            users.forEach(user -> {
                createUserProfile(user, workspace);
                log.info("✅ Created user: {} ({})", user.getEmail(), user.getUserId());
            });

            log.info("🎉 Test data initialization completed! Created {} users.", users.size());
        };
    }

    private User createInitialUser() {
        if (userRepository.findByEmail("owner@wealist.com").isPresent()) {
            return userRepository.findByEmail("owner@wealist.com").get();
        }
        User owner = User.builder()
                .email("owner@wealist.com")
                .provider("google")
                .googleId("google-id-owner")
                .isActive(true)
                .build();
        return userRepository.save(owner);
    }

    private Workspace createDefaultWorkspace(User owner) {
        if (workspaceRepository.count() > 0) {
            return workspaceRepository.findAll().get(0);
        }
        Workspace workspace = Workspace.builder()
                .ownerId(owner.getUserId())
                .workspaceName("Default Workspace")
                .workspaceDescription("This is a default workspace for testing.")
                .isPublic(true)
                .needApproved(false)
                .build();
        return workspaceRepository.save(workspace);
    }

    private List<User> createTestUsers(User owner) {
        List<User> users = new ArrayList<>();
        users.add(owner);

        List<DummyUserData> dummyUsers = createDummyUserData();

        for (DummyUserData data : dummyUsers) {
            if (userRepository.existsByEmailAndIsActiveTrue(data.email)) {
                log.debug("⏭️  User already exists: {}", data.email);
                continue;
            }
            User user = User.builder()
                    .email(data.email)
                    .provider("google")
                    .googleId(data.googleId)
                    .isActive(true)
                    .build();
            users.add(userRepository.save(user));
        }
        return users;
    }

    private void createUserProfile(User user, Workspace workspace) {
        UserProfile profile = UserProfile.builder()
                .userId(user.getUserId())
                .workspaceId(workspace.getWorkspaceId())
                .nickName(user.getEmail().split("@")[0])
                .email(user.getEmail())
                .profileImageUrl("https://i.pravatar.cc/150?u=" + user.getUserId())
                .build();
        userProfileRepository.save(profile);
    }

    private List<DummyUserData> createDummyUserData() {
        List<DummyUserData> users = new ArrayList<>();
        users.add(new DummyUserData("chulsoo.kim@example.com", "google-id-001"));
        users.add(new DummyUserData("younghee.lee@example.com", "google-id-002"));
        users.add(new DummyUserData("minsu.park@example.com", "google-id-003"));
        users.add(new DummyUserData("sujin.jung@example.com", "google-id-004"));
        users.add(new DummyUserData("dongwook.choi@example.com", "google-id-005"));
        return users;
    }

    private static class DummyUserData {
        String email;
        String googleId;

        public DummyUserData(String email, String googleId) {
            this.email = email;
            this.googleId = googleId;
        }
    }
}
