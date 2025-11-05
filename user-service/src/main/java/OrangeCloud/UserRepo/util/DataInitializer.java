package OrangeCloud.UserRepo.util;

import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.repository.UserRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.ApplicationRunner;
import org.springframework.stereotype.Component;

import java.util.UUID;

@Component
@RequiredArgsConstructor
public class DataInitializer implements ApplicationRunner {

    private final UserRepository userRepository;
//    private final ExtensionService extensionService;

    @Value("${jwt.secret}")
    private String jwtSecret;


    @Override
    public void run(ApplicationArguments args) throws Exception {
        // 초기 관리자 사용자가 없으면 생성
        if (!userRepository.existsByEmailAndIsActiveTrue("admin@orangecloud.com")) {
            User adminUser = User.builder()
                    .name("System Admin")
                    .email("admin@orangecloud.com")
                    .provider("local")
                    .role("ROLE_ADMIN")
                    .isActive(true)
                    .build();

            userRepository.save(adminUser);
        }

        // 서버 시작 시 확장자 초기화
//        extensionService.saveAllExtensionsFromEnum();
    }
}
