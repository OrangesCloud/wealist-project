package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.*;
import OrangeCloud.UserRepo.dto.auth.AuthResponse;
import OrangeCloud.UserRepo.dto.auth.LoginRequest;
import OrangeCloud.UserRepo.dto.auth.SignupRequest;
import OrangeCloud.UserRepo.dto.userinfo.UserInfoResponse;
import OrangeCloud.UserRepo.service.AuthService;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.validation.Valid;
import io.swagger.v3.oas.annotations.Operation;
import OrangeCloud.UserRepo.exception.InvalidTokenException;
import OrangeCloud.UserRepo.exception.ErrorCode;
import OrangeCloud.UserRepo.service.RateLimitingService;
import java.util.UUID;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/auth")
@CrossOrigin(origins = "*", maxAge = 3600)
@Tag(name = "Authentication", description = "인증 관련 API")
public class AuthController {

    private static final Logger logger = LoggerFactory.getLogger(AuthController.class);

    private final AuthService authService;
    private final RateLimitingService rateLimitingService;

    @Autowired
    public AuthController(AuthService authService, RateLimitingService rateLimitingService) {
        this.authService = authService;
        this.rateLimitingService = rateLimitingService;
    }

    // 회원가입
//    @PostMapping("/signup")
//    @Operation(summary = "회원가입", description = "새로운 사용자 계정을 생성합니다.")
//    @io.swagger.v3.oas.annotations.responses.ApiResponse(responseCode = "200", description = "회원가입 성공")
//    @io.swagger.v3.oas.annotations.responses.ApiResponse(responseCode = "400", description = "잘못된 요청")
//    public ResponseEntity<AuthResponse> signup(@Valid @RequestBody SignupRequest signupRequest, HttpServletRequest request) {
//        String clientIp = request.getRemoteAddr();
//        rateLimitingService.checkRateLimit(clientIp + ":signup", 5, 60); // 1분당 5회 제한
//
//        logger.debug("Received signup request for email: {}", signupRequest.getEmail());
//        AuthResponse authResponse = authService.signup(signupRequest);
//        logger.info(MessageCode.SIGNUP_SUCCESS.getMessage() + ": {}", signupRequest.getEmail());
//        return ResponseEntity.ok(authResponse);
//    }

    // 로그인
//    @PostMapping("/login")
//    @Operation(summary = "로그인", description = "사용자 인증을 수행하고 JWT 토큰을 발급합니다.")
//    @io.swagger.v3.oas.annotations.responses.ApiResponse(responseCode = "200", description = "로그인 성공")
//    @io.swagger.v3.oas.annotations.responses.ApiResponse(responseCode = "400", description = "인증 실패")
//    public ResponseEntity<AuthResponse> login(@Valid @RequestBody LoginRequest loginRequest, HttpServletRequest request) {
//        String clientIp = request.getRemoteAddr();
//        rateLimitingService.checkRateLimit(clientIp + ":login", 5, 60); // 1분당 5회 제한
//
//        logger.debug("Received login request for email: {}", loginRequest.getEmail());
//        AuthResponse authResponse = authService.login(loginRequest);
//        logger.info(MessageCode.LOGIN_SUCCESS.getMessage() + ": {}", loginRequest.getEmail());
//        return ResponseEntity.ok(authResponse);
//    }

    // 로그아웃
    @PostMapping("/logout")
    @Operation(summary = "로그아웃", description = "현재 세션을 종료하고 토큰을 무효화합니다.")
    @io.swagger.v3.oas.annotations.responses.ApiResponse(responseCode = "200", description = "로그아웃 성공")
    @io.swagger.v3.oas.annotations.responses.ApiResponse(responseCode = "401", description = "인증되지 않은 사용자")
    public ResponseEntity<MessageApiResponse> logout(HttpServletRequest request) {
        logger.debug("Received logout request.");
        String token = extractTokenFromRequest(request);
        authService.logout(token);
        logger.info(MessageCode.LOGOUT_SUCCESS.getMessage());
        return ResponseEntity.ok(new MessageApiResponse(true, MessageCode.LOGOUT_SUCCESS.getMessage()));
    }

    // 토큰 갱신
    @PostMapping("/refresh")
    @Operation(summary = "토큰 갱신", description = "Refresh Token을 사용하여 Access Token을 갱신합니다.")
    @io.swagger.v3.oas.annotations.responses.ApiResponse(responseCode = "200", description = "토큰 갱신 성공")
    @io.swagger.v3.oas.annotations.responses.ApiResponse(responseCode = "400", description = "유효하지 않은 토큰")
    public ResponseEntity<AuthResponse> refresh(@Valid @RequestBody RefreshTokenRequest refreshRequest) {
        logger.debug("Received refresh token request.");
        AuthResponse authResponse = authService.refreshToken(refreshRequest.getRefreshToken());
        logger.info(MessageCode.TOKEN_REFRESH_SUCCESS.getMessage());
        return ResponseEntity.ok(authResponse);
    }

    // 내 정보 조회
    @GetMapping("/me")
    @Operation(summary = "내 정보 조회", description = "현재 인증된 사용자의 정보를 조회합니다.")
    @io.swagger.v3.oas.annotations.responses.ApiResponse(responseCode = "200", description = "조회 성공")
    @io.swagger.v3.oas.annotations.responses.ApiResponse(responseCode = "401", description = "인증되지 않은 사용자")
    public ResponseEntity<UserInfoResponse> getCurrentUser(Authentication authentication) {
        UUID userId = UUID.fromString(authentication.getName()); // Convert principal name (UUID string) to UUID
        logger.debug("Received request to get current user info for ID: {}", userId);
        UserInfoResponse userInfo = authService.getCurrentUserInfo(userId); // Pass UUID
        logger.info(MessageCode.USER_INFO_RETRIEVED_SUCCESS.getMessage() + ": {}", userId);
        return ResponseEntity.ok(userInfo);
    }
    @GetMapping("/test")
    public ResponseEntity<AuthResponse> testUser(){
        return ResponseEntity.ok(authService.TestLogin());
    }

    // 헬퍼 메서드: Request에서 토큰 추출
    private String extractTokenFromRequest(HttpServletRequest request) {
        logger.debug("Attempting to extract token from request.");
        String bearerToken = request.getHeader("Authorization");
        if (bearerToken != null && bearerToken.startsWith("Bearer ")) {
            String token = bearerToken.substring(7);
            logger.debug("Token extracted successfully.");
            return token;
        }
        logger.warn(ErrorCode.TOKEN_NOT_FOUND.getMessage());
        throw new InvalidTokenException("Authorization 헤더에서 토큰을 찾을 수 없습니다.");
    }
}