package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.MessageApiResponse;
import OrangeCloud.UserRepo.dto.userinfo.CreateUserInfoRequest;
import OrangeCloud.UserRepo.dto.userinfo.UpdateUserInfoRequest;
import OrangeCloud.UserRepo.entity.UserInfo;
import OrangeCloud.UserRepo.service.UserInfoService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@RestController
@RequestMapping("/api/userinfo")
@RequiredArgsConstructor
public class UserInfoController {

    private final UserInfoService userInfoService;

    // 사용자 정보 소프트 삭제
    @DeleteMapping("/{userId}")
    public ResponseEntity<MessageApiResponse> deleteUser(@PathVariable UUID userId) {
        boolean deleted = userInfoService.softDeleteUser(userId);
        if (deleted) {
            return ResponseEntity.ok(MessageApiResponse.success("사용자가 성공적으로 삭제되었습니다."));
        } else {
            return ResponseEntity.notFound().build();
        }
    }

    // 활성화된 사용자 정보 목록 조회
    @GetMapping
    public ResponseEntity<List<UserInfo>> getAllActiveUsers() {
        List<UserInfo> users = userInfoService.getAllActiveUsers();
        return ResponseEntity.ok(users);
    }

    // 그룹별 활성화된 사용자 조회
    @GetMapping("/group/{groupId}")
    public ResponseEntity<List<UserInfo>> getActiveUsersByGroup(@PathVariable UUID groupId) {
        List<UserInfo> users = userInfoService.getActiveUsersByGroupId(groupId);
        return ResponseEntity.ok(users);
    }

    // 역할별 활성화된 사용자 정보 조회
    @GetMapping("/role/{role}")
    public ResponseEntity<List<UserInfo>> getActiveUsersByRole(@PathVariable String role) {
        List<UserInfo> users = userInfoService.getActiveUsersByRole(role);
        return ResponseEntity.ok(users);
    }

    // 사용자 정보 재활성화
    @PutMapping("/{userId}/reactivate")
    public ResponseEntity<MessageApiResponse> reactivateUser(@PathVariable UUID userId) {
        boolean reactivated = userInfoService.reactivateUser(userId);
        if (reactivated) {
            return ResponseEntity.ok(MessageApiResponse.success("사용자가 성공적으로 재활성화되었습니다."));
        } else {
            return ResponseEntity.notFound().build();
        }
    }

    // 그룹의 모든 사용자 비활성화
    @DeleteMapping("/group/{groupId}")
    public ResponseEntity<MessageApiResponse> deleteUsersByGroup(@PathVariable UUID groupId) {
        boolean deleted = userInfoService.softDeleteUsersByGroupId(groupId);
        if (deleted) {
            return ResponseEntity.ok(MessageApiResponse.success("그룹의 모든 사용자가 성공적으로 삭제되었습니다."));
        } else {
            return ResponseEntity.notFound().build();
        }
    }

    // 활성화된 사용자 수 조회
    @GetMapping("/count")
    public ResponseEntity<MessageApiResponse> getActiveUserCount() {
        long count = userInfoService.getActiveUserCount();
        return ResponseEntity.ok(MessageApiResponse.success("활성화된 사용자 수를 조회했습니다.", count));
    }

    // 사용자 생성 - CreateUserInfoRequest 사용
    @PostMapping
    public ResponseEntity<UserInfo> createUser(@Valid @RequestBody CreateUserInfoRequest request) {

        UserInfo userInfo = userInfoService.createUser(request.getGroupId(), request.getRole(),request.getUserId());
        return ResponseEntity.ok(userInfo);
    }

    // 사용자 정보 수정 - UpdateUserInfoRequest 사용
    @PutMapping("/{userId}")
    public ResponseEntity<UserInfo> updateUser(@PathVariable UUID userId,
                                               @Valid @RequestBody UpdateUserInfoRequest request) {


        Optional<UserInfo> updatedUser = userInfoService.updateUser(userId, request.getGroupId(), request.getRole());
        return updatedUser.map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

}