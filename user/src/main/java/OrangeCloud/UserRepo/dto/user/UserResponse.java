// UserResponse.java
package OrangeCloud.UserRepo.dto.user;

import java.time.LocalDateTime;
import java.util.UUID;

public class UserResponse {
    private UUID userId;
    private String name;
    private String email;
    private LocalDateTime createdAt;

    // 기본 생성자
    public UserResponse() {}

    // 생성자
    public UserResponse(UUID userId, String name, String email, LocalDateTime createdAt) {
        this.userId = userId;
        this.name = name;
        this.email = email;
        this.createdAt = createdAt;
    }

    // Getter 메서드들
    public UUID getUserId() { return userId; }
    public String getName() { return name; }
    public String getEmail() { return email; }
    public LocalDateTime getCreatedAt() { return createdAt; }

    // Setter 메서드들
    public void setUserId(UUID userId) { this.userId = userId; }
    public void setName(String name) { this.name = name; }
    public void setEmail(String email) { this.email = email; }
    public void setCreatedAt(LocalDateTime createdAt) { this.createdAt = createdAt; }
}