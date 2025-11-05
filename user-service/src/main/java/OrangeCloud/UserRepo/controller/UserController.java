package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.MessageApiResponse;
import OrangeCloud.UserRepo.dto.auth.LoginRequest;
import OrangeCloud.UserRepo.dto.user.CountResponse;
import OrangeCloud.UserRepo.dto.user.*;
import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.service.UserService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;

@RestController
@RequestMapping("/api/users")
@RequiredArgsConstructor
@Slf4j
public class UserController {

    private final UserService userService;


    // 사용자 소프트 삭제
    @DeleteMapping("/{userId}")
    public ResponseEntity<MessageApiResponse> deleteUser(@PathVariable UUID userId) {
        log.info("Request to delete user: {}", userId);

        boolean deleted = userService.softDeleteUser(userId);
        if (deleted) {
            return ResponseEntity.ok(MessageApiResponse.builder()
                    .success(true)
                    .message("사용자가 성공적으로 삭제되었습니다.")
                    .build());
        } else {
            return ResponseEntity.notFound().build();
        }
    }

    // 활성화된 사용자 목록 조회
    @GetMapping
    public ResponseEntity<List<User>> getAllActiveUsers() {
        log.info("Request to get all active users");
        List<User> users = userService.getAllActiveUsers();
        return ResponseEntity.ok(users);
    }

    // 특정 사용자 조회
    @GetMapping("/{userId}")
    public ResponseEntity<User> getActiveUser(@PathVariable UUID userId) {
        log.info("Request to get user: {}", userId);

        Optional<User> user = userService.getActiveUserById(userId);
        return user.map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

//    // 이메일로 사용자 조회
//    @GetMapping("/email/{email}")
//    public ResponseEntity<User> getActiveUserByEmail(@PathVariable String email) {
//        log.info("Request to get user by email: {}", email);
//
//        Optional<User> user = userService.getActiveUserByEmail(email);
//        return user.map(ResponseEntity::ok)
//                .orElse(ResponseEntity.notFound().build());
//    }
//
//    // 이름으로 사용자 검색
//    @GetMapping("/search")
//    public ResponseEntity<List<User>> searchUsersByName(@RequestParam String name) {
//        log.info("Request to search users by name: {}", name);
//        List<User> users = userService.searchActiveUsersByName(name);
//        return ResponseEntity.ok(users);
//    }
    // 단일 이름검색 및 단일 이메일 검색
    @GetMapping("/search")
    public ResponseEntity<List<User>> searchUsers(@RequestParam String query) {
        log.info("Searching users by query: {}", query);

        if (query == null || query.trim().isEmpty()) {
            return ResponseEntity.badRequest().build();
        }

        List<User> users = userService.searchUsers(query);
        return ResponseEntity.ok(users);
}

    // 요청하는 사용자 id 기반 모든 사용자 정보
    @PostMapping("/ids")
    public ResponseEntity<List<User>> getUsersByIds(@RequestBody UserIdRequest request) {
        return ResponseEntity.ok(userService.getUsersByUserIds(request.getUserIds()));
    }
    //유저 프로필 이미지 설정



//    // 사용자 정보 수정
//    @PutMapping("/{userId}")
//    public ResponseEntity<User> updateUser(@PathVariable UUID userId,
//                                           @Valid @RequestBody UpdateUserRequest request) {
//        log.info("Request to update user: {} with data: {}", userId, request);
//
//        try {
//            Optional<User> updatedUser = userService.updateUser(userId, request.getName(), request.getEmail());
//            return updatedUser.map(ResponseEntity::ok)
//                    .orElse(ResponseEntity.notFound().build());
//        } catch (IllegalArgumentException e) {
//            return ResponseEntity.badRequest().build();
//        }
//    }

    // 비밀번호 변경 메서드 수정
//    @PatchMapping("/{userId}/password")
//    public ResponseEntity<MessageApiResponse> changePassword(@PathVariable UUID userId,
//                                                             @Valid @RequestBody ChangePasswordRequest request) {
//        log.info("Request to change password for user: {}", userId);
//
//        boolean changed = userService.changePassword(userId, request.getCurrentPassword(), request.getNewPassword());
//        if (changed) {
//            return ResponseEntity.ok(MessageApiResponse.success("비밀번호가 성공적으로 변경되었습니다."));
//        } else {
//            return ResponseEntity.badRequest().body(MessageApiResponse.failure("현재 비밀번호가 올바르지 않습니다."));
//        }
//    }

    // 사용자 재활성화
    @PutMapping("/{userId}/reactivate")
    public ResponseEntity<MessageApiResponse> reactivateUser(@PathVariable UUID userId) {
        log.info("Request to reactivate user: {}", userId);

        boolean reactivated = userService.reactivateUser(userId);
        if (reactivated) {
            return ResponseEntity.ok(MessageApiResponse.success("사용자가 성공적으로 재활성화되었습니다."));
        } else {
            return ResponseEntity.notFound().build();
        }
    }

    // 로그인
    @PostMapping("/login")
    public ResponseEntity<User> login(@Valid @RequestBody LoginRequest request) {
        log.info("Login request for email: {}", request.getEmail());

        Optional<User> user = userService.authenticateUser(request.getEmail());
        return user.map(ResponseEntity::ok)
                .orElse(ResponseEntity.status(401).build()); // 이렇게 수정
    }

    // 이메일 중복 체크
    @GetMapping("/check-email")
    public ResponseEntity<MessageApiResponse> checkEmailAvailability(@RequestParam String email) {
        boolean available = userService.isEmailAvailable(email);
        return ResponseEntity.ok(MessageApiResponse.success(
                available ? "사용 가능한 이메일입니다." : "이미 사용 중인 이메일입니다.",
                available
        ));
    }







    // 활성화된 사용자 수 조회
    @GetMapping("/count")
    public ResponseEntity<CountResponse> getActiveUserCount() {
        long count = userService.getActiveUserCount();
        return ResponseEntity.ok(CountResponse.builder()
                .count(count)
                .message("활성화된 사용자 수")
                .build());
    }


}