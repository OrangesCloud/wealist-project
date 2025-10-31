package OrangeCloud.UserRepo.dto.team;

import jakarta.validation.constraints.NotNull;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.UUID;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class AddTeamMemberRequest {

    @NotNull(message = "사용자 ID는 필수입니다.")
    private UUID userId;  // User 테이블의 실제 사용자 ID

    private String teamRole; // 팀에서의 역할 (개발자, 디자이너 등)
}