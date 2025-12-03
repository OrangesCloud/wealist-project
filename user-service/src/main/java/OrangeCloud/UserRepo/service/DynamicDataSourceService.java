package OrangeCloud.UserRepo.service;

import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.core.env.ConfigurableEnvironment;
import org.springframework.core.env.MapPropertySource;
import org.springframework.stereotype.Service;

import java.util.HashMap;
import java.util.Map;

@Service
public class DynamicDataSourceService {

    private final ConfigurableApplicationContext context;
    private final ConfigurableEnvironment environment;

    public DynamicDataSourceService(ConfigurableApplicationContext context, ConfigurableEnvironment environment) {
        this.context = context;
        this.environment = environment;
    }

    public void connect(String url, String username, String password) {
        // 1. 데이터베이스 연결 정보를 Environment에 추가
        Map<String, Object> dbProperties = new HashMap<>();
        dbProperties.put("spring.datasource.url", url);
        dbProperties.put("spring.datasource.username", username);
        dbProperties.put("spring.datasource.password", password);
        
        environment.getPropertySources().addFirst(new MapPropertySource("dynamicDataSource", dbProperties));

        // 2. "with-database" 프로파일 활성화
        environment.addActiveProfile("with-database");

        // 3. 애플리케이션 컨텍스트 새로고침
        // 이렇게 하면 변경된 환경과 활성화된 프로파일을 기반으로 스프링이 컨텍스트를 다시 로드합니다.
        // `DatabaseConfig`가 이제 활성화되어 필요한 빈들을 생성합니다.
        Thread refreshThread = new Thread(() -> context.refresh());
        refreshThread.setDaemon(false);
        refreshThread.start();
    }
}
