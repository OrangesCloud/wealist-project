package OrangeCloud.UserRepo.service;

import OrangeCloud.UserRepo.entity.Group;
import OrangeCloud.UserRepo.repository.GroupRepository;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
@Transactional
@Slf4j
public class GroupService {

    @Autowired
    private GroupRepository groupRepository;

    // 그룹 비활성화 (소프트 삭제)
    public boolean deactivateGroup(UUID groupId) {
        int updatedRows = groupRepository.deactivateById(groupId);
        return updatedRows > 0;
    }
    // 그룹 생성
    // 그룹 생성 (동일한 회사명이 있으면 기존 그룹 반환)
    public Group createGroup(String name, String companyName) {
        // 항상 새로운 UUID로 그룹 생성
        Group group = Group.builder()
                .name(name)
                .companyName(companyName)
                .isActive(true)
                .build();

        return groupRepository.save(group);
    }

    // 삭제 시간과 함께 비활성화
    public boolean deactivateGroupWithTimestamp(UUID groupId) {
        int updatedRows = groupRepository.deactivateByIdWithTimestamp(groupId, LocalDateTime.now());
        return updatedRows > 0;
    }

    // 소프트 삭제 (현재 시간 자동 설정)
    public boolean softDeleteGroup(UUID groupId) {
        int updatedRows = groupRepository.softDeleteById(groupId);
        return updatedRows > 0;
    }

    // 활성화된 모든 그룹 조회
    public List<Group> getAllActiveGroups() {
        return groupRepository.findAllActiveGroups();
    }

    // ID로 활성화된 그룹 조회
    public Optional<Group> getActiveGroupById(UUID groupId) {
        return groupRepository.findActiveById(groupId);
    }

    // 회사명으로 활성화된 그룹 조회
    public List<Group> getActiveGroupsByCompanyName(String companyName) {
        return groupRepository.findActiveByCompanyName(companyName);
    }

    // 그룹명으로 검색
    public List<Group> searchActiveGroupsByName(String name) {
        return groupRepository.findActiveByNameContaining(name);
    }

    // 그룹 재활성화
    public boolean reactivateGroup(UUID groupId) {
        int updatedRows = groupRepository.reactivateById(groupId);
        return updatedRows > 0;
    }

    // 활성화된 그룹 수 조회
    public long getActiveGroupCount() {
        return groupRepository.countActiveGroups();
    }

    // 비활성화된 그룹 조회 (관리자용)
    public List<Group> getInactiveGroups() {
        return groupRepository.findInactiveGroups();
    }
    public Group findOrCreateGroupByCompanyName(String companyName, String groupName) {
        return createGroup(groupName, companyName);
    }



    // 그룹 정보 수정
    public Optional<Group> updateGroup(UUID groupId, String name, String companyName) {
        Optional<Group> groupOpt = groupRepository.findActiveById(groupId);
        if (groupOpt.isPresent()) {
            Group group = groupOpt.get();
            group.setName(name);
            group.setCompanyName(companyName);
            return Optional.of(groupRepository.save(group));
        }
        return Optional.empty();
    }
    public Optional<Group> findGroupByCompanyName(String companyName) {


        List<Group> existingGroups = groupRepository.findActiveByCompanyName(companyName);

        if (!existingGroups.isEmpty()) {
            Group existingGroup = existingGroups.get(0);

            return Optional.of(existingGroup);
        }

        log.info("No existing group found for company: {}", companyName);
        return Optional.empty();
    }
    public Group createNewGroupForceNew(String name, String companyName) {
        // 강제로 새 그룹 생성 (회사명이 같아도 새로 만들기)
        Group group = Group.builder()
                .name(name)
                .companyName(companyName)
                .isActive(true)
                .build();

        return groupRepository.save(group);
    }
    public boolean existsByCompanyName(String companyName) {
        return groupRepository.existsActiveByCompanyName(companyName);
    }
    public Optional<UUID> getExistingGroupIdByCompanyName(String companyName) {
        List<Group> existingGroups = groupRepository.findActiveByCompanyName(companyName);
        if (!existingGroups.isEmpty()) {
            return Optional.of(existingGroups.get(0).getGroupId());
        }
        return Optional.empty();
    }
    public List<Group> getAllGroupsByCompanyName(String companyName) {
        return groupRepository.findActiveByCompanyName(companyName);
    }
    public long countGroupsByCompanyName(String companyName) {
        return groupRepository.countActiveByCompanyName(companyName);
    }
}