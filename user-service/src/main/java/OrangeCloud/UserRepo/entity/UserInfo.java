package OrangeCloud.UserRepo.entity;

import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(name = "userinfo")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
@ToString
@EqualsAndHashCode(of = "userId")
public class UserInfo {
    @Id
//    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(name = "user_id", updatable = false, nullable = false, columnDefinition = "UUID")
    private UUID userId;

    @Column(name = "group_id", nullable = false, columnDefinition = "UUID")
    private UUID groupId;




    @Column(name = "role")  // 사용자의 역할 추가
    private String role;


    // 활성화 상태 필드
    @Column(name = "is_active", nullable = false)
    @Builder.Default
    private Boolean isActive = true;

    // 삭제 시간 필드
    @Column(name = "deleted_at")
    private LocalDateTime deletedAt;
}