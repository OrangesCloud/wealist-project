package OrangeCloud.UserRepo.config;


import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class JacksonConfig {

    /**
     * Spring Boot 애플리케이션에서 ObjectMapper Bean을 정의하고 설정합니다.
     * 이 Bean은 애플리케이션 전체에서 JSON 직렬화/역직렬화에 사용됩니다.
     */
    @Bean
    public ObjectMapper objectMapper() {
        // ObjectMapper 인스턴스를 생성합니다.
        ObjectMapper objectMapper = new ObjectMapper();

        // 💡 핵심: Java 8 날짜/시간 타입(java.time.LocalDateTime 등) 처리를 위한 모듈을 등록합니다.
        // 이 모듈이 있어야 LocalDateTime 객체를 ISO 8601 문자열로 변환할 수 있습니다.
        objectMapper.registerModule(new JavaTimeModule());
        
        // 참고: 필요에 따라 다른 설정을 추가할 수 있습니다 (예: 특정 필드 무시 등)

        return objectMapper;
    }
}