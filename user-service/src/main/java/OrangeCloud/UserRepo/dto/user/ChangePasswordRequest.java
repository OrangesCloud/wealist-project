//package OrangeCloud.UserRepo.dto.user;
//
//import jakarta.validation.constraints.NotBlank;
//import jakarta.validation.constraints.Size;
//import lombok.AllArgsConstructor;
//import lombok.Builder;
//import lombok.Data;
//import lombok.NoArgsConstructor;
//
//@Data
//@NoArgsConstructor
//@AllArgsConstructor
//@Builder
//public class ChangePasswordRequest {
//
//    @NotBlank(message = "현재 비밀번호는 필수입니다.")
//    private String currentPassword;
//
//    @NotBlank(message = "새 비밀번호는 필수입니다.")
//    @Size(min = 8, message = "비밀번호는 8자 이상이어야 합니다.")
//    private String newPassword;
//
//    @Override
//    public String toString() {
//        return "ChangePasswordRequest{" +
//                "currentPassword='[PROTECTED]'" +
//                ", newPassword='[PROTECTED]'" +
//                '}';
//    }
//}