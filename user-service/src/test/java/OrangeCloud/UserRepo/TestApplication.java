package OrangeCloud.UserRepo;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.SpringBootConfiguration;
import org.springframework.boot.autoconfigure.EnableAutoConfiguration;
import org.springframework.boot.autoconfigure.domain.EntityScan;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.FilterType;

@SpringBootConfiguration
@EnableAutoConfiguration
@ComponentScan(
    basePackages = "OrangeCloud.UserRepo",
    excludeFilters = {
        @ComponentScan.Filter(
            type = FilterType.ASSIGNABLE_TYPE,
            classes = UserRepoApplication.class  // ✅ 메인 앱 클래스 제외
        )
    }
)

@EntityScan(basePackages = "OrangeCloud.UserRepo.entity") // Entity 위치
// @EnableJpaRepositories(basePackages = "OrangeCloud.UserRepo.repository") // Repository 위치

public class TestApplication {
    // 테스트 전용 애플리케이션 컨텍스트
    public static void main(String[] args) {
        SpringApplication.run(TestApplication.class, args);
    }
}
