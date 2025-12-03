package OrangeCloud.UserRepo;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.autoconfigure.jdbc.DataSourceAutoConfiguration;
import org.springframework.data.web.config.EnableSpringDataWebSupport;
import org.springframework.scheduling.annotation.EnableScheduling;
import org.springframework.boot.autoconfigure.orm.jpa.HibernateJpaAutoConfiguration;
import org.springframework.data.jpa.repository.config.EnableJpaRepositories;

import static org.springframework.data.web.config.EnableSpringDataWebSupport.PageSerializationMode.VIA_DTO;

@SpringBootApplication(exclude = {
    DataSourceAutoConfiguration.class,  // 조건부로 DB 설정하기 위해 제외
    HibernateJpaAutoConfiguration.class
})
// @EnableJpaRepositories(basePackages = "OrangeCloud.UserRepo.repository")
@EnableSpringDataWebSupport(pageSerializationMode = VIA_DTO)
@EnableScheduling
public class UserRepoApplication {
	public static void main(String[] args) {
		 SpringApplication app = new SpringApplication(UserRepoApplication.class);
        
		app.setLazyInitialization(true);
		app.run(args);
	}
}