package OrangeCloud.UserRepo.entity;

import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Entity
@Table(name = "teams")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class Team {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID teamId;

    private String teamName;
    private UUID groupId;
    private UUID leaderId;
    private String description;

    @Builder.Default
    private Boolean isActive = true;

    // 소프트 삭제를 위한 필드 추가
    @Column(name = "deleted_at")
    private LocalDateTime deletedAt;

    // 팀 멤버들의 ID 목록을 JSON으로 저장
    @Column(name = "member_ids", columnDefinition = "TEXT")
    private String memberIds;

    @CreationTimestamp
    private LocalDateTime createdAt;

    @UpdateTimestamp
    private LocalDateTime updatedAt;

    // 기존 헬퍼 메서드들...
    public List<UUID> getMemberIdsList() {
        if (memberIds == null || memberIds.isEmpty()) {
            return new ArrayList<>();
        }

        try {
            ObjectMapper mapper = new ObjectMapper();
            String[] ids = mapper.readValue(memberIds, String[].class);
            return Arrays.stream(ids)
                    .map(UUID::fromString)
                    .collect(Collectors.toList());
        } catch (Exception e) {
            return new ArrayList<>();
        }
    }

    public void setMemberIdsList(List<UUID> memberList) {
        try {
            ObjectMapper mapper = new ObjectMapper();
            this.memberIds = mapper.writeValueAsString(memberList);
        } catch (Exception e) {
            this.memberIds = "[]";
        }
    }

    public void addMember(UUID userId) {
        List<UUID> members = getMemberIdsList();
        if (!members.contains(userId)) {
            members.add(userId);
            setMemberIdsList(members);
        }
    }

    public void removeMember(UUID userId) {
        List<UUID> members = getMemberIdsList();
        members.remove(userId);
        setMemberIdsList(members);
    }

    public boolean hasMember(UUID userId) {
        return getMemberIdsList().contains(userId);
    }

    // 소프트 삭제 메서드 추가
    public void softDelete() {
        this.isActive = false;
        this.deletedAt = LocalDateTime.now();
    }

    public void reactivate() {
        this.isActive = true;
        this.deletedAt = null;
    }
}