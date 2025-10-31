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
public class UpdateTeamRequest {

    @NotNull(message = "사용자 ID는 필수입니다.")
    private UUID userId;

    @NotNull(message = "그룹 ID는 필수입니다.")
    private UUID groupId;
}