package OrangeCloud.UserRepo.entity;

import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(name = "users")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
@ToString
@EqualsAndHashCode(of = "userId")
public class User {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(name = "user_id", updatable = false, nullable = false, columnDefinition = "UUID")
    private UUID userId;

    @Column(nullable = false)
    private String name;

    @Column(nullable = false, unique = true)
    private String email;

    // OAuth2 로그인 시 비밀번호가 없을 수 있으므로 nullable
    @Column(name = "password_hash",nullable=true)
    private String passwordHash;

    // 로그인 제공자 (google, local 등)
    @Column(name = "provider")
    @Builder.Default
    private String provider = "local";

    // 사용자 역할
    @Column(name = "role")
    @Builder.Default
    private String role = "ROLE_USER";

    @CreationTimestamp
    @Column(name = "created_at", updatable = false)
    private LocalDateTime createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at")
    private LocalDateTime updatedAt;

    // 소프트 삭제를 위한 필드들
    @Column(name = "is_active", nullable = false)
    @Builder.Default
    private Boolean isActive = true;

    @Column(name = "deleted_at")
    private LocalDateTime deletedAt;

    /**
     * OAuth2 로그인 시 사용자 정보 업데이트
     */
    public User updateOAuth2Info(String name) {
        this.name = name;
        return this;
    }

    /**
     * 소프트 삭제 처리
     */
    public void softDelete() {
        this.isActive = false;
        this.deletedAt = LocalDateTime.now();
    }

    /**
     * 소프트 삭제 복구
     */
    public void restore() {
        this.isActive = true;
        this.deletedAt = null;
    }
}