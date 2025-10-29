// src/main/java/OrangeCloud/UserRepo/dto/userinfo/UpdateUserInfoRequest.java
package OrangeCloud.UserRepo.dto.userinfo;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.UUID;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class UpdateUserInfoRequest {

    @NotNull(message = "그룹 ID는 필수입니다.")
    private UUID groupId;

    @NotBlank(message = "역할은 필수입니다.")
    @Size(min = 1, max = 50, message = "역할은 1자 이상 50자 이하여야 합니다.")
    private String role;
}