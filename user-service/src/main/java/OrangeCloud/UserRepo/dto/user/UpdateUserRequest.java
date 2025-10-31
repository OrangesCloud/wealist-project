package OrangeCloud.UserRepo.dto.user;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.Size;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class UpdateUserRequest {

    @Size(min = 2, max = 50, message = "이름은 2자 이상 50자 이하여야 합니다.")
    private String name;

    @Email(message = "올바른 이메일 형식이어야 합니다.")
    private String email;

    @Override
    public String toString() {
        return "UpdateUserRequest{" +
                "name='" + name + '\'' +
                ", email='" + email + '\'' +
                '}';
    }
}