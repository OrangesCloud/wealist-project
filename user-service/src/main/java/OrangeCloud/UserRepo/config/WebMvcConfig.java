package OrangeCloud.UserRepo.config;

import OrangeCloud.UserRepo.interceptor.MetricInterceptor;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

@Configuration
public class WebMvcConfig implements WebMvcConfigurer {

    private final MetricInterceptor metricInterceptor;

    public WebMvcConfig(MetricInterceptor metricInterceptor) {
        this.metricInterceptor = metricInterceptor;
    }

    @Override
    public void addInterceptors(InterceptorRegistry registry) {
        registry.addInterceptor(metricInterceptor);
    }
}
