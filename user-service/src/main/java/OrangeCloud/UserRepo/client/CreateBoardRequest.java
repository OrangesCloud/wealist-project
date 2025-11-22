package OrangeCloud.UserRepo.client;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.Map;
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
    
    @JsonProperty("author_id")
    private UUID authorId;
    
    @JsonProperty("assignee_id")
    private UUID assigneeId;
    
    @JsonProperty("title")
    private String title;
    
    @JsonProperty("content")
    private String content;
    
    @JsonProperty("custom_fields")
    private Map<String, Object> customFields;
    
    @JsonProperty("start_date")
    private LocalDateTime startDate;
    
    @JsonProperty("due_date")
    private LocalDateTime dueDate;
}
