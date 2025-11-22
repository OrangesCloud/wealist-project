package OrangeCloud.UserRepo.client;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;
import java.util.UUID;

/**
 * Request DTO for creating a board in Board-Service.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CreateBoardRequest {
    
    @JsonProperty("project_id")
    private UUID projectId;
    
    @JsonProperty("title")
    private String title;
    
    @JsonProperty("content")
    private String content;
    
    @JsonProperty("role_ids")
    private List<UUID> roleIds;
    
    @JsonProperty("stage_id")
    private UUID stageId;
    
    @JsonProperty("importance_id")
    private UUID importanceId;
    
    @JsonProperty("assignee_id")
    private UUID assigneeId;
    
    @JsonProperty("dueDate")
    private String dueDate;
}
