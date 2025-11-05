package OrangeCloud.UserRepo.entity;

import jakarta.persistence.*;
import lombok.Data;

import java.util.UUID;

@Entity
@Data
@Table(name = "images")
public class Image {

    @Id
    @Column(name = "user_id")
    private UUID userId;  // userId를 Primary Key로 사용

    @Column(name = "image_url")
    private String imageUrl;
}
