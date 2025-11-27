package OrangeCloud.UserRepo.dto.user.projection;

import java.util.UUID;

public interface UserAndMembershipProjection {
    UUID getUserId();
    String getEmail();

    // 멤버십 유무 (컬럼이 존재하면 true, 아니면 null/false)
    Boolean getIsMember();
}
