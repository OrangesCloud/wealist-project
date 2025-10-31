// UserInfoResponse.java
package OrangeCloud.UserRepo.dto.userinfo;

import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.UUID;

@Data
@NoArgsConstructor
public class UserInfoResponse {
    private UUID id;
    private String username;
    private String email;
    private String name;
    private LocalDateTime createdAt;

    // 올바른 생성자 (5개 매개변수)
    public UserInfoResponse(UUID id, String username, String email, String name, LocalDateTime createdAt) {
        this.id = id;
        this.username = username;
        this.email = email;
        this.name = name;
        this.createdAt = createdAt;
    }
}