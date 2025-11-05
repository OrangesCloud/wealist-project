package OrangeCloud.UserRepo.service;

import OrangeCloud.UserRepo.entity.Image;
import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.repository.ImageRepository;
import OrangeCloud.UserRepo.repository.UserRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Service
public class ImageService {

    private final UserRepository userRepository;
    private final ImageRepository imageRepository;

    public ImageService(UserRepository userRepository, ImageRepository imageRepository) {
        this.userRepository = userRepository;
        this.imageRepository = imageRepository;
    }

    @Transactional
    public String saveImageUrl(UUID userId, String imageUrl) {
        User user = userRepository.findById(userId)
                .orElseThrow(() -> new RuntimeException("User not found with id: " + userId));

        // userId로 기존 이미지 조회 (Primary Key 사용)
        Image existingImage = imageRepository.findById(userId).orElse(null);

        if (existingImage != null) {
            // 기존 이미지 URL 업데이트
            existingImage.setImageUrl(imageUrl);
            imageRepository.save(existingImage);
        } else {
            // 새 이미지 생성
            Image image = new Image();
            image.setUserId(userId);  // Primary Key 설정
            image.setImageUrl(imageUrl);
            imageRepository.save(image);
        }

        return imageUrl;
    }

    @Transactional
    public void deleteImageUrl(UUID userId) {
        User user = userRepository.findById(userId)
                .orElseThrow(() -> new RuntimeException("User not found with id: " + userId));

        // Primary Key로 직접 삭제
        if (imageRepository.existsById(userId)) {
            imageRepository.deleteById(userId);
        }
    }

    public String getImageUrl(UUID userId) {
        User user = userRepository.findById(userId)
                .orElseThrow(() -> new RuntimeException("User not found with id: " + userId));

        // Primary Key로 직접 조회
        Image image = imageRepository.findById(userId).orElse(null);
        if (image != null) {
            return image.getImageUrl();
        }

        // 기본 이미지 URL
        return "/images/default-profile.png";
    }
}