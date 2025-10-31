// src/main/java/OrangeCloud/UserRepo/dto/group/CreateGroupRequest.java
package OrangeCloud.UserRepo.dto.group;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class CreateGroupRequest {

    @NotBlank(message = "그룹명은 필수입니다.")
    @Size(min = 2, max = 100, message = "그룹명은 2자 이상 100자 이하여야 합니다.")
    private String name;

    @Size(max = 100, message = "회사명은 100자 이하여야 합니다.")
    private String companyName;
}