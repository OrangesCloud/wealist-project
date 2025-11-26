package OrangeCloud.UserRepo.dto.user.projection;

import java.util.UUID;
public interface CreatorAndProfileProjection {
    UUID getUserId();
    String getUsername();
    String getEmail(); // 예시: 응답에 이메일이 필요하다고 가정
    String getNickName();
}
