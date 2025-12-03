package OrangeCloud.UserRepo.config;

import OrangeCloud.UserRepo.repository.*;
import org.mockito.Mockito;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import org.springframework.context.annotation.Profile;

@Configuration
@Profile("staging")
public class StagingMockConfig {

    @Bean
    @Primary
    public AttachmentRepository attachmentRepository() {
        return Mockito.mock(AttachmentRepository.class);
    }

    @Bean
    @Primary
    public UserRepository userRepository() {
        return Mockito.mock(UserRepository.class);
    }

    @Bean
    @Primary
    public UserProfileRepository userProfileRepository() {
        return Mockito.mock(UserProfileRepository.class);
    }

    @Bean
    @Primary
    public WorkspaceRepository workspaceRepository() {
        return Mockito.mock(WorkspaceRepository.class);
    }

    @Bean
    @Primary
    public WorkspaceMemberRepository workspaceMemberRepository() {
        return Mockito.mock(WorkspaceMemberRepository.class);
    }

    @Bean
    @Primary
    public WorkspaceJoinRequestRepository workspaceJoinRequestRepository() {
        return Mockito.mock(WorkspaceJoinRequestRepository.class);
    }
}