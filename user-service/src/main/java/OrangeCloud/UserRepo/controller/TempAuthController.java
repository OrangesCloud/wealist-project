
package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.util.JwtTokenProvider;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;

@RestController
@RequestMapping("/temp")
@RequiredArgsConstructor
@Tag(name = "Temporary Auth", description = "임시 인증 API (테스트용)")
public class TempAuthController {

    private final JwtTokenProvider jwtTokenProvider;
    private static final Map<String, TempUser> tempUserStore = new ConcurrentHashMap<>();
    private final PasswordEncoder passwordEncoder = new BCryptPasswordEncoder();

    @Data
    static class TempUser {
        private UUID id;
        private String email;
        private String password;
        private String name;

        TempUser(String email, String password, String name) {
            this.id = UUID.randomUUID();
            this.email = email;
            this.password = password;
            this.name = name;
        }
    }

    @Data
    static class TempSignUpRequest {
        @NotBlank @Email
        private String email;
        @NotBlank
        private String password;
        @NotBlank
        private String name;
    }

    @Data
    static class TempLoginRequest {
        @NotBlank @Email
        private String email;
        @NotBlank
        private String password;
    }

    @Data
    static class AuthResponse {
        private String accessToken;
        private UUID userId;

        AuthResponse(String accessToken, UUID userId) {
            this.accessToken = accessToken;
            this.userId = userId;
        }
    }

    @Operation(summary = "임시 회원가입", description = "테스트용 임시 계정을 생성합니다.")
    @PostMapping("/signup")
    public ResponseEntity<?> tempSignUp(@Valid @RequestBody TempSignUpRequest request) {
        if (tempUserStore.containsKey(request.getEmail())) {
            return ResponseEntity.status(HttpStatus.CONFLICT).body("Email already exists");
        }
        String encodedPassword = passwordEncoder.encode(request.getPassword());
        TempUser newUser = new TempUser(request.getEmail(), encodedPassword, request.getName());
        tempUserStore.put(newUser.getEmail(), newUser);
        return ResponseEntity.status(HttpStatus.CREATED).body("User created successfully");
    }

    @Operation(summary = "임시 로그인", description = "테스트용 임시 계정으로 로그인하고 JWT를 발급받습니다.")
    @PostMapping("/login")
    public ResponseEntity<?> tempLogin(@Valid @RequestBody TempLoginRequest request) {
        TempUser user = tempUserStore.get(request.getEmail());

        if (user == null || !passwordEncoder.matches(request.getPassword(), user.getPassword())) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body("Invalid credentials");
        }

        String token = jwtTokenProvider.generateToken(user.getId());
        return ResponseEntity.ok(new AuthResponse(token, user.getId()));
    }
}
