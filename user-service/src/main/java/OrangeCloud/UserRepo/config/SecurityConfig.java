package OrangeCloud.UserRepo.config;

import OrangeCloud.UserRepo.filter.JwtAuthenticationFilter;
import OrangeCloud.UserRepo.filter.JwtExceptionFilter;
import OrangeCloud.UserRepo.oauth.CustomOAuth2UserService;
import OrangeCloud.UserRepo.oauth.OAuth2SuccessHandler;
import OrangeCloud.UserRepo.service.AuthService;
import OrangeCloud.UserRepo.util.JwtTokenProvider;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;
import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.CorsConfigurationSource;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;

import java.util.Arrays;

@Configuration
@EnableWebSecurity
public class SecurityConfig {

    @Autowired(required = false)
    private CustomOAuth2UserService customOAuth2UserService;

    @Autowired(required = false)
    private OAuth2SuccessHandler oAuth2SuccessHandler;

    private final ObjectMapper objectMapper;

    public SecurityConfig(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
    }

    @Bean
    public SecurityFilterChain filterChain(
            HttpSecurity http,
            JwtTokenProvider jwtTokenProvider,
            AuthService authService) throws Exception {
        // JWT 필터 생성
        JwtAuthenticationFilter jwtAuthenticationFilter = new JwtAuthenticationFilter(jwtTokenProvider,
                authService);

        // JWT 예외 처리 필터 생성
        JwtExceptionFilter jwtExceptionFilter = new JwtExceptionFilter(objectMapper);

        http
                .csrf(csrf -> csrf.disable())
                .cors(cors -> cors.configurationSource(corsConfigurationSource()))
                .sessionManagement(session -> {
                    // OAuth2가 활성화된 경우 세션 사용, 그렇지 않으면 STATELESS
                    if (customOAuth2UserService != null && oAuth2SuccessHandler != null) {
                        session.sessionCreationPolicy(SessionCreationPolicy.IF_REQUIRED);
                    } else {
                        session.sessionCreationPolicy(SessionCreationPolicy.STATELESS);
                    }
                })
                .authorizeHttpRequests(authz -> authz
                        // 공통 허용 경로 (Health Check, Swagger, Actuator 등)
                        .requestMatchers("/swagger-ui/**", "/v3/api-docs/**", "/swagger-ui.html").permitAll()
                        .requestMatchers("/api/auth/signup", "/api/auth/login", "/api/auth/refresh").permitAll()
                        .requestMatchers("/login/oauth2/**", "/oauth2/**").permitAll()
                        .requestMatchers("/test", "/error", "/").permitAll()
                        .requestMatchers("/actuator/**").permitAll() // Actuator 경로 전체 허용
                        
                        // ************ 나중에 아래 전체 허용 해제 필수 **********
                        .requestMatchers("/**").permitAll() // 모든 요청 허용 (현재 디버깅 목적)
                        // 나머지는 인증 필요
                        .anyRequest().authenticated())
                
                // 🚨 필터 순서 수정: JWT 인증이 AnonymousAuthenticationFilter보다 먼저 실행되도록 합니다.
                // 1. JWT 인증 필터 등록 (인증 처리)
                .addFilterBefore(jwtAuthenticationFilter, UsernamePasswordAuthenticationFilter.class) 
                
                // 2. JWT 예외 처리 필터 등록 (JWT 인증 필터가 던진 예외를 잡도록 그 바로 뒤에 위치)
                // ExceptionFilter는 인증 필터의 예외를 처리해야 하므로, 등록할 때 JWT 필터보다 앞에 오도록 설정합니다.
                .addFilterBefore(jwtExceptionFilter, JwtAuthenticationFilter.class)

                .headers(headers -> headers
                        .frameOptions(frame -> frame.sameOrigin()));

        // OAuth2 로그인 설정 추가
        if (customOAuth2UserService != null && oAuth2SuccessHandler != null) {
            http.oauth2Login(oauth2 -> oauth2
                    .userInfoEndpoint(userInfo -> userInfo
                            .userService(customOAuth2UserService))
                    .authorizationEndpoint(
                            endpoint -> endpoint.baseUri("/oauth2/authorization"))
                    .redirectionEndpoint(
                            endpoint -> endpoint.baseUri("/login/oauth2/code/*"))
                    .successHandler(oAuth2SuccessHandler));
        }

        return http.build();
    }

    @Bean
    public CorsConfigurationSource corsConfigurationSource() {
        CorsConfiguration configuration = new CorsConfiguration();

        // 허용할 Origin 설정 - 개발 환경에서는 모든 Origin 허용
        configuration.setAllowedOriginPatterns(Arrays.asList("*"));

        // 허용할 HTTP 메서드
        configuration.setAllowedMethods(Arrays.asList(
                "GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"));

        // 허용할 헤더
        configuration.setAllowedHeaders(Arrays.asList("*"));

        // 노출할 헤더 (클라이언트에서 접근 가능한 헤더)
        configuration.setExposedHeaders(Arrays.asList(
                "Authorization", "Content-Type", "X-Requested-With"));

        // 인증 정보 포함 허용
        configuration.setAllowCredentials(true);

        // preflight 요청 캐시 시간 (초)
        configuration.setMaxAge(3600L);

        UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
        source.registerCorsConfiguration("/**", configuration);
        return source;
    }
}