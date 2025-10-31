package OrangeCloud.UserRepo.dto.team;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotEmpty;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;
import java.util.UUID;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class CreateTeamRequest {

    @NotBlank(message = "팀 이름은 필수입니다.")
    @Size(min = 2, max = 100, message = "팀 이름은 2자 이상 100자 이하여야 합니다.")
    private String teamName;

    @NotBlank(message = "회사명은 필수입니다.")
    private String companyName;

    private String groupName; // 새 그룹 생성시 사용할 그룹명

    @NotNull(message = "팀장 ID는 필수입니다.")
    private UUID leaderId;  // User 테이블의 실제 사용자 ID

    private String description;
}