package OrangeCloud.UserRepo.aspect;

import io.micrometer.core.instrument.Counter;
import org.aspectj.lang.annotation.AfterThrowing;
import org.aspectj.lang.annotation.Aspect;
import org.springframework.stereotype.Component;

@Aspect
@Component
public class ClientMetricsAspect {

    private final Counter externalApiErrorCounter;

    public ClientMetricsAspect(Counter externalApiErrorCounter) {
        this.externalApiErrorCounter = externalApiErrorCounter;
    }

    @AfterThrowing(pointcut = "execution(* OrangeCloud.UserRepo.client..*(..))", throwing = "ex")
    public void countExternalApiErrors(Exception ex) {
        externalApiErrorCounter.increment();
    }
}
