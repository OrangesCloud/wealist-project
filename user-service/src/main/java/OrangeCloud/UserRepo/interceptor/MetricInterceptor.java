package OrangeCloud.UserRepo.interceptor;

import io.micrometer.core.instrument.Counter;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.HandlerInterceptor;

@Component
public class MetricInterceptor implements HandlerInterceptor {

    private final Counter httpRequestTotalCounter;

    public MetricInterceptor(Counter httpRequestTotalCounter) {
        this.httpRequestTotalCounter = httpRequestTotalCounter;
    }

    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) {
        httpRequestTotalCounter.increment();
        return true;
    }
}
