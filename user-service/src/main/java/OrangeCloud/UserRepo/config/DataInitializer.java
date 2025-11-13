package OrangeCloud.UserRepo.config;

import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.entity.UserProfile;
import OrangeCloud.UserRepo.entity.Workspace;
import OrangeCloud.UserRepo.entity.WorkspaceMember;
import OrangeCloud.UserRepo.repository.UserProfileRepository;
import OrangeCloud.UserRepo.repository.UserRepository;
import OrangeCloud.UserRepo.repository.WorkspaceMemberRepository;
import OrangeCloud.UserRepo.repository.WorkspaceRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.CommandLineRunner;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.env.Environment;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.UUID;

@Slf4j
@Configuration
@RequiredArgsConstructor
public class DataInitializer {

    private final UserRepository userRepository;
    private final UserProfileRepository userProfileRepository;
    private final WorkspaceRepository workspaceRepository;
    private final WorkspaceMemberRepository workspaceMemberRepository;
    private final Environment environment;

    @Bean
    @Transactional
    public CommandLineRunner initializeData() {
        return args -> {
            // 1️⃣ 개발 환경 체크
            String[] activeProfiles = environment.getActiveProfiles();
            boolean isDev = Arrays.asList(activeProfiles).contains("dev")
                    || Arrays.asList(activeProfiles).contains("local")
                    || activeProfiles.length == 0;

            if (!isDev) {
                log.info("⏭️ Production environment detected. Skipping dummy data initialization.");
                return;
            }

            // 2️⃣ 충분한 데이터 존재 여부 체크
            if (userRepository.count() >= 50 && workspaceRepository.count() >= 10 && userProfileRepository.count() >= 50) {
                log.info("✅ Database already has sufficient data. Skipping initialization.");
                return;
            }

            log.info("🚀 Starting dummy data initialization...");

            // 3️⃣ 사용자 생성
            List<User> users = new ArrayList<>();
            for (int i = 1; i <= 50; i++) {
                String email = "user" + i + "@example.com";
                if (userRepository.existsByEmailAndIsActiveTrue(email)) continue;

                User user = User.builder()
                        .email(email)
                        .provider("google")
                        .googleId("google-id-" + String.format("%03d", i))
                        .isActive(true)
                        .build();
                users.add(user);
            }
            userRepository.saveAll(users);
            List<User> allUsers = userRepository.findAll();
            log.info("✅ Created {} users.", allUsers.size());

            // 4️⃣ 워크스페이스 생성
            List<Workspace> workspaces = new ArrayList<>();
            for (int i = 0; i < 10; i++) {
                User owner = allUsers.get(i);
                Workspace ws = Workspace.builder()
                        .workspaceName("테스트 워크스페이스 " + (i + 1))
                        .workspaceDescription("이것은 테스트 워크스페이스 " + (i + 1) + "입니다.")
                        .ownerId(owner.getUserId())
                        .build();
                workspaces.add(ws);
            }
            workspaceRepository.saveAll(workspaces);
            List<Workspace> allWorkspaces = workspaceRepository.findAll();
            log.info("✅ Created {} workspaces.", allWorkspaces.size());

            // 5️⃣ UserProfile + WorkspaceMember 생성 (5명씩 워크스페이스 배정)
            List<UserProfile> profiles = new ArrayList<>();
            List<WorkspaceMember> members = new ArrayList<>();

            for (int i = 0; i < allUsers.size(); i++) {
                User user = allUsers.get(i);
                Workspace targetWorkspace = allWorkspaces.get(i / 5); // 5명씩 배정

                // ➤ UserProfile 생성
                if (!userProfileRepository.existsByUserIdAndWorkspaceId(user.getUserId(), targetWorkspace.getWorkspaceId())) {
                    UserProfile profile = UserProfile.builder()
                            .userId(user.getUserId())
                            .workspaceId(targetWorkspace.getWorkspaceId())
                            .nickName("테스터" + (i + 1))
                            .email(user.getEmail())
                            .profileImageUrl("https://i.pravatar.cc/150?img=" + (i + 1))
                            .build();
                    profiles.add(profile);
                }

                // ➤ WorkspaceMember 생성
                if (!workspaceMemberRepository.existsByUserIdAndWorkspaceId(user.getUserId(), targetWorkspace.getWorkspaceId())) {
                    WorkspaceMember member = WorkspaceMember.builder()
                            .userId(user.getUserId())
                            .workspaceId(targetWorkspace.getWorkspaceId())
                            .role(WorkspaceMember.WorkspaceRole.MEMBER)
                            .isDefault(i % 5 == 0) // 각 그룹의 첫 번째만 기본
                            .build();
                    members.add(member);
                }
            }

            userProfileRepository.saveAll(profiles);
            workspaceMemberRepository.saveAll(members);
            log.info("✅ Created {} user profiles.", profiles.size());
            log.info("✅ Created {} workspace members.", members.size());

            log.info("🎉 Data initialization finished successfully.");
        };
    }
}
