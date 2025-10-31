package OrangeCloud.UserRepo;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.data.web.config.EnableSpringDataWebSupport;

import org.springframework.data.jpa.repository.config.EnableJpaRepositories;

import static org.springframework.data.web.config.EnableSpringDataWebSupport.PageSerializationMode.VIA_DTO;

@SpringBootApplication
@EnableJpaRepositories(basePackages = "OrangeCloud.UserRepo.repository")
@EnableSpringDataWebSupport(pageSerializationMode = VIA_DTO)
public class UserRepoApplication {
	public static void main(String[] args) {
		SpringApplication.run(UserRepoApplication.class, args);
	}
}