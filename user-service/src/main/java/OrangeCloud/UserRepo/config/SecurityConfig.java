package OrangeCloud.UserRepo.config;

import OrangeCloud.UserRepo.filter.JwtAuthenticationFilter;
import OrangeCloud.UserRepo.filter.JwtExceptionFilter;
import OrangeCloud.UserRepo.oauth.CustomOAuth2UserService;
import OrangeCloud.UserRepo.oauth.OAuth2SuccessHandler;
import OrangeCloud.UserRepo.service.AuthService;
import OrangeCloud.UserRepo.util.JwtTokenProvider;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
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
@RequiredArgsConstructor
public class SecurityConfig {

        private final CustomOAuth2UserService customOAuth2UserService;
        private final OAuth2SuccessHandler oAuth2SuccessHandler;
        private final ObjectMapper objectMapper;

        @Bean
        public SecurityFilterChain filterChain(
                        HttpSecurity http,
                        JwtTokenProvider jwtTokenProvider,
                        AuthService authService) throws Exception {
                // JWT 필터 생성
                JwtAuthenticationFilter jwtAuthenticationFilter = new JwtAuthenticationFilter(jwtTokenProvider,
                                authService);

                JwtExceptionFilter jwtExceptionFilter = new JwtExceptionFilter(objectMapper);

                return http
                                .csrf(csrf -> csrf.disable())
                                .cors(cors -> cors.configurationSource(corsConfigurationSource()))
                                .sessionManagement(session -> session
                                                .sessionCreationPolicy(SessionCreationPolicy.STATELESS))
                                .authorizeHttpRequests(authz -> authz
                                                // Swagger UI 경로 허용
                                                .requestMatchers("/swagger-ui/**").permitAll()
                                                .requestMatchers("/swagger-ui.html").permitAll()
                                                .requestMatchers("/v3/api-docs/**").permitAll()
                                                .requestMatchers("/swagger-resources/**").permitAll()
                                                // 인증 API 허용 (회원가입, 로그인)
                                                .requestMatchers("/api/auth/signup").permitAll()
                                                .requestMatchers("/api/auth/login").permitAll()
                                                .requestMatchers("/api/auth/refresh").permitAll()
                                                // OAuth2 로그인 경로 허용 (인증 요청 및 리디렉션)
                                                .requestMatchers("/login/oauth2/**").permitAll()
                                                .requestMatchers("/oauth2/**").permitAll()
                                                // 테스트 엔드포인트 허용
                                                .requestMatchers("/test").permitAll()
                                                .requestMatchers("/error").permitAll()
                                                .requestMatchers("/").permitAll()
                                                .requestMatchers("/actuator/health").permitAll()
                                                // ************ 나중에 아래 전체 허용 해제 필수 **********
                                                .requestMatchers("/**").permitAll()
                                                // 나머지는 인증 필요
                                                .anyRequest().authenticated())

                                // 🚨 필터 순서 (예외 -> 인증)
                                // 1. JWT 예외 처리 필터를 JWT 인증 필터 앞에 둡니다.
                                .addFilterBefore(jwtExceptionFilter, JwtAuthenticationFilter.class)
                                // 2. JWT 인증 필터를 UsernamePasswordAuthenticationFilter 앞에 둡니다.
                                .addFilterBefore(jwtAuthenticationFilter, UsernamePasswordAuthenticationFilter.class)

                                // ----------------------------------------------------
                                // 🔑 OAuth2 로그인 설정 (Endpoint 명시적 추가)
                                // ----------------------------------------------------
                                .oauth2Login(oauth2 -> oauth2
                                                // 💡 User Service 등록
                                                .userInfoEndpoint(userInfo -> userInfo
                                                                .userService(customOAuth2UserService))
                                                // 💡 인증 시작 경로 명시 (ex: /api/users/oauth2/authorization/google)
                                                .authorizationEndpoint(
                                                                endpoint -> endpoint.baseUri("/oauth2/authorization"))
                                                // 💡 리디렉션 응답 처리 경로 명시 (ex: /api/users/login/oauth2/code/google)
                                                .redirectionEndpoint(
                                                                endpoint -> endpoint.baseUri("/login/oauth2/code/*"))
                                                // 💡 성공/실패 핸들러 등록
                                                .successHandler(oAuth2SuccessHandler))
                                .headers(headers -> headers
                                                .frameOptions(frame -> frame.sameOrigin()))
                                .build();
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