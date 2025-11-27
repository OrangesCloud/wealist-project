package OrangeCloud.UserRepo.aspect;

import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.Timer;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.springframework.stereotype.Component;

import java.util.concurrent.TimeUnit;

@Aspect
@Component
public class RepositoryMetricsAspect {

    private final Timer dbQueryDurationTimer;
    private final Counter dbQueryErrorCounter;

    public RepositoryMetricsAspect(Timer dbQueryDurationTimer, Counter dbQueryErrorCounter) {
        this.dbQueryDurationTimer = dbQueryDurationTimer;
        this.dbQueryErrorCounter = dbQueryErrorCounter;
    }

    @Around("execution(* OrangeCloud.UserRepo.repository..*(..))")
    public Object measureQueryDuration(ProceedingJoinPoint pjp) throws Throwable {
        long start = System.nanoTime();
        try {
            return pjp.proceed();
        } catch (Throwable t) {
            dbQueryErrorCounter.increment();
            throw t;
        } finally {
            long end = System.nanoTime();
            dbQueryDurationTimer.record(end - start, TimeUnit.NANOSECONDS);
        }
    }
}
