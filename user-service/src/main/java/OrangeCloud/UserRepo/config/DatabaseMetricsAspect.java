package OrangeCloud.UserRepo.config;

import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.aspectj.lang.annotation.Pointcut;
import org.springframework.stereotype.Component;

import java.time.Duration;

@Aspect
@Component
public class DatabaseMetricsAspect {

    private final DatabaseMetrics databaseMetrics;

    // 생성자 주입 (권장)
    public DatabaseMetricsAspect(DatabaseMetrics databaseMetrics) {
        this.databaseMetrics = databaseMetrics;
    }

    // Pointcut 분리: 가독성 및 재사용성 향상
    // 'OrangeCloud.UserRepo' 패키지 하위의 모든 'Repository'로 끝나는 클래스/인터페이스
    @Pointcut("execution(* OrangeCloud.UserRepo..*Repository.*(..))")
    public void repositoryMethods() {}

    @Around("repositoryMethods()")
    public Object measureDatabaseQuery(ProceedingJoinPoint joinPoint) throws Throwable {
        long startTime = System.nanoTime();
        String outcome = "SUCCESS"; // 기본 상태
        Throwable exception = null;

        try {
            return joinPoint.proceed();
        } catch (Throwable ex) {
            outcome = "FAILURE";
            exception = ex;
            throw ex; // 예외는 반드시 다시 던져야 비즈니스 로직이 정상 작동함
        } finally {
            // 메트릭 기록 로직을 별도로 분리하여 finally 블록을 깔끔하게 유지
            recordMetrics(joinPoint, startTime, outcome, exception);
        }
    }

    private void recordMetrics(ProceedingJoinPoint joinPoint, long startTime, String outcome, Throwable exception) {
        try {
            Duration duration = Duration.ofNanos(System.nanoTime() - startTime);

            // 중요: Proxy 객체가 아닌 선언된 인터페이스/클래스 타입을 가져옴
            String className = joinPoint.getSignature().getDeclaringType().getSimpleName();
            String methodName = joinPoint.getSignature().getName();

            String table = extractTableName(className);
            String operation = extractOperation(methodName);
            String exceptionName = (exception != null) ? exception.getClass().getSimpleName() : "None";

            // DatabaseMetrics 클래스에 outcome(성공여부)와 exception 파라미터도 넘겨주는 것을 추천
            databaseMetrics.recordQuery(operation, table, duration, outcome, exceptionName);
            
        } catch (Exception e) {
            // 메트릭 기록 중 에러가 발생해도 비즈니스 로직에는 영향을 주지 않도록 로깅만 함
            // log.warn("Failed to record database metrics", e);
        }
    }

    private String extractTableName(String repositoryName) {
        // "UserRepository" -> "user"
        return repositoryName.replace("Repository", "").toLowerCase();
    }

    private String extractOperation(String methodName) {
        if (methodName.startsWith("find") || methodName.startsWith("get")) return "SELECT";
        if (methodName.startsWith("save")) return "INSERT_UPDATE"; // JPA save는 insert/update 둘 다 가능
        if (methodName.startsWith("delete") || methodName.startsWith("remove")) return "DELETE";
        if (methodName.startsWith("count")) return "COUNT";
        if (methodName.startsWith("exists")) return "EXISTS";
        return "OTHER";
    }
}