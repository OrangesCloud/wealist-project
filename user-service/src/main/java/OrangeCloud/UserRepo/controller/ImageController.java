package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.service.ImageService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.UUID;

@Slf4j
@RestController
@RequestMapping("/api/users/{userId}/image")
public class ImageController {

    private final ImageService imageService;

    public ImageController(ImageService imageService) {
        this.imageService = imageService;
    }

    @PostMapping
    public ResponseEntity<String> saveImageUrl(@PathVariable UUID userId, @RequestParam("imageUrl") String imageUrl) {
        System.out.println("===== 이미지 URL 저장 요청 들어옴 =====");
        System.out.println("imageUrl: " + imageUrl);
        System.out.println("userId: " + userId);

        String savedImageUrl = imageService.saveImageUrl(userId, imageUrl);
        return ResponseEntity.ok().body("Image URL saved successfully: " + savedImageUrl);
    }

    @DeleteMapping
    public ResponseEntity<Void> deleteImage(@PathVariable UUID userId) {
        imageService.deleteImageUrl(userId);
        return ResponseEntity.noContent().build();
    }

    @GetMapping
    public ResponseEntity<String> getImageUrl(@PathVariable UUID userId) {
        String imageUrl = imageService.getImageUrl(userId);
        return ResponseEntity.ok().body(imageUrl);
    }
}