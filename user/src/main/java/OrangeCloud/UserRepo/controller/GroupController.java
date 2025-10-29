package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.MessageApiResponse;
import OrangeCloud.UserRepo.dto.group.CreateGroupRequest;
import OrangeCloud.UserRepo.dto.group.UpdateGroupRequest;
import OrangeCloud.UserRepo.entity.Group;
import OrangeCloud.UserRepo.service.GroupService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@RestController
@RequestMapping("/api/groups")
@RequiredArgsConstructor
@Slf4j
public class GroupController {

    private final GroupService groupService;

    // 그룹 생성 (회사명 중복 시 기존 그룹 반환)
    @PostMapping
    public ResponseEntity<MessageApiResponse> createGroup(@Valid @RequestBody CreateGroupRequest request) {
        log.info("Request to create group: {}", request);

        try {
            // 같은 회사명의 기존 그룹들이 있는지 확인
            boolean existsAlready = groupService.existsByCompanyName(request.getCompanyName());
            long existingGroupCount = groupService.countGroupsByCompanyName(request.getCompanyName());

            // 항상 새로운 UUID로 그룹 생성
            Group group = groupService.createGroup(request.getName(), request.getCompanyName());

            // 기존 그룹이 있었는지에 따라 다른 메시지 반환
            String message;
            if (existsAlready) {
                message = String.format("'%s' 회사의 %d번째 그룹 '%s'이(가) 생성되었습니다. (그룹 ID: %s)",
                        request.getCompanyName(), existingGroupCount + 1, request.getName(), group.getGroupId());
            } else {
                message = String.format("'%s' 회사의 첫 번째 그룹 '%s'이(가) 성공적으로 생성되었습니다. (그룹 ID: %s)",
                        request.getCompanyName(), request.getName(), group.getGroupId());
            }

            return ResponseEntity.ok(MessageApiResponse.success(message, group));
        } catch (Exception e) {
            return ResponseEntity.badRequest()
                    .body(MessageApiResponse.failure("그룹 생성에 실패했습니다: " + e.getMessage()));
        }
    }
    // 회사별 그룹 목록 조회 API 추가
    @GetMapping("/company/{companyName}/all")
    public ResponseEntity<MessageApiResponse> getAllGroupsByCompany(@PathVariable String companyName) {
        log.info("Request to get all groups by company name: {}", companyName);

        List<Group> groups = groupService.getAllGroupsByCompanyName(companyName);
        long groupCount = groups.size();

        String message = String.format("'%s' 회사의 그룹 %d개를 성공적으로 조회했습니다.", companyName, groupCount);
        return ResponseEntity.ok(MessageApiResponse.success(message, groups));
    }

    // 강제로 새 그룹 생성 (회사명이 같아도 새로 생성)
    @PostMapping("/force-new")
    public ResponseEntity<MessageApiResponse> createNewGroupForce(@Valid @RequestBody CreateGroupRequest request) {
        log.info("Request to force create new group: {}", request);

        try {
            Group group = groupService.createNewGroupForceNew(request.getName(), request.getCompanyName());
            return ResponseEntity.ok(MessageApiResponse.success(
                    String.format("'%s' 회사의 새 그룹이 강제로 생성되었습니다.", request.getCompanyName()),
                    group));
        } catch (Exception e) {
            return ResponseEntity.badRequest()
                    .body(MessageApiResponse.failure("그룹 생성에 실패했습니다: " + e.getMessage()));
        }
    }

    // 회사명으로 기존 그룹 조회
    @GetMapping("/company/{companyName}")
    public ResponseEntity<MessageApiResponse> getGroupByCompanyName(@PathVariable String companyName) {
        log.info("Request to get group by company name: {}", companyName);

        Optional<Group> group = groupService.findGroupByCompanyName(companyName);
        if (group.isPresent()) {
            return ResponseEntity.ok(MessageApiResponse.success(
                    String.format("'%s' 회사의 그룹 정보를 성공적으로 조회했습니다.", companyName),
                    group.get()));
        } else {
            return ResponseEntity.ok(MessageApiResponse.success(
                    String.format("'%s' 회사의 그룹이 존재하지 않습니다.", companyName),
                    null));
        }
    }

    // 회사명 중복 체크
    @GetMapping("/check-company/{companyName}")
    public ResponseEntity<MessageApiResponse> checkCompanyExists(@PathVariable String companyName) {
        boolean exists = groupService.existsByCompanyName(companyName);
        String message = exists ?
                String.format("'%s' 회사명의 그룹이 이미 존재합니다.", companyName) :
                String.format("'%s' 회사명의 그룹이 존재하지 않습니다.", companyName);
        return ResponseEntity.ok(MessageApiResponse.success(message, exists));
    }

    // 활성화된 그룹 목록 조회
    @GetMapping
    public ResponseEntity<MessageApiResponse> getAllActiveGroups() {
        log.info("Request to get all active groups");
        List<Group> groups = groupService.getAllActiveGroups();
        return ResponseEntity.ok(MessageApiResponse.success("활성화된 그룹 목록을 성공적으로 조회했습니다.", groups));
    }

    // 특정 그룹 조회
    @GetMapping("/{groupId}")
    public ResponseEntity<MessageApiResponse> getActiveGroup(@PathVariable UUID groupId) {
        log.info("Request to get group: {}", groupId);

        Optional<Group> group = groupService.getActiveGroupById(groupId);
        if (group.isPresent()) {
            return ResponseEntity.ok(MessageApiResponse.success("그룹 정보를 성공적으로 조회했습니다.", group.get()));
        } else {
            return ResponseEntity.notFound().build();
        }
    }

    // 회사명으로 그룹 검색
    @GetMapping("/search")
    public ResponseEntity<MessageApiResponse> searchGroupsByCompany(@RequestParam String companyName) {
        log.info("Request to search groups by company: {}", companyName);
        List<Group> groups = groupService.getActiveGroupsByCompanyName(companyName);
        return ResponseEntity.ok(MessageApiResponse.success("회사별 그룹 목록을 성공적으로 조회했습니다.", groups));
    }

    // 그룹명으로 그룹 검색
    @GetMapping("/search/name")
    public ResponseEntity<MessageApiResponse> searchGroupsByName(@RequestParam String name) {
        log.info("Request to search groups by name: {}", name);
        List<Group> groups = groupService.searchActiveGroupsByName(name);
        return ResponseEntity.ok(MessageApiResponse.success("그룹명 검색 결과를 성공적으로 조회했습니다.", groups));
    }

    // 그룹 정보 수정
    @PutMapping("/{groupId}")
    public ResponseEntity<MessageApiResponse> updateGroup(@PathVariable UUID groupId,
                                                          @Valid @RequestBody UpdateGroupRequest request) {
        log.info("Request to update group: {} with data: {}", groupId, request);

        try {
            Optional<Group> updatedGroup = groupService.updateGroup(groupId, request.getName(), request.getCompanyName());
            if (updatedGroup.isPresent()) {
                return ResponseEntity.ok(MessageApiResponse.success("그룹 정보가 성공적으로 수정되었습니다.", updatedGroup.get()));
            } else {
                return ResponseEntity.notFound().build();
            }
        } catch (Exception e) {
            return ResponseEntity.badRequest()
                    .body(MessageApiResponse.failure("그룹 수정에 실패했습니다: " + e.getMessage()));
        }
    }

    // 그룹 소프트 삭제
    @DeleteMapping("/{groupId}")
    public ResponseEntity<MessageApiResponse> deleteGroup(@PathVariable UUID groupId) {
        log.info("Request to delete group: {}", groupId);

        boolean deleted = groupService.softDeleteGroup(groupId);
        if (deleted) {
            return ResponseEntity.ok(MessageApiResponse.success("그룹이 성공적으로 삭제되었습니다."));
        } else {
            return ResponseEntity.notFound().build();
        }
    }

    // 그룹 재활성화
    @PutMapping("/{groupId}/reactivate")
    public ResponseEntity<MessageApiResponse> reactivateGroup(@PathVariable UUID groupId) {
        log.info("Request to reactivate group: {}", groupId);

        boolean reactivated = groupService.reactivateGroup(groupId);
        if (reactivated) {
            return ResponseEntity.ok(MessageApiResponse.success("그룹이 성공적으로 재활성화되었습니다."));
        } else {
            return ResponseEntity.notFound().build();
        }
    }

    // 활성화된 그룹 수 조회
    @GetMapping("/count")
    public ResponseEntity<MessageApiResponse> getActiveGroupCount() {
        long count = groupService.getActiveGroupCount();
        return ResponseEntity.ok(MessageApiResponse.success("활성화된 그룹 수를 성공적으로 조회했습니다.", count));
    }

    // 비활성화된 그룹 조회 (관리자용)
    @GetMapping("/inactive")
    public ResponseEntity<MessageApiResponse> getInactiveGroups() {
        log.info("Request to get inactive groups");
        List<Group> groups = groupService.getInactiveGroups();
        return ResponseEntity.ok(MessageApiResponse.success("비활성화된 그룹 목록을 성공적으로 조회했습니다.", groups));
    }
}