package OrangeCloud.UserRepo.dto.team;

import OrangeCloud.UserRepo.entity.Group;
import OrangeCloud.UserRepo.entity.Team;
import OrangeCloud.UserRepo.entity.UserInfo;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class TeamWithMembersResponse {
    private Team team;
    private List<UserInfo> members;
    private Group group;
    private long memberCount;
}