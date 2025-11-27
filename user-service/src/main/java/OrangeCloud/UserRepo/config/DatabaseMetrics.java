package OrangeCloud.UserRepo.config;

import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;
import org.springframework.stereotype.Component;
import java.time.Duration;

@Component
public class DatabaseMetrics {
    
    private final MeterRegistry meterRegistry;
    
    public DatabaseMetrics(MeterRegistry meterRegistry) {
        this.meterRegistry = meterRegistry;
    }
    
    // 파라미터 추가: outcome (SUCCESS/FAILURE), exception (예외이름)
    public void recordQuery(String operation, String table, Duration duration, String outcome, String exception) {
        Timer.builder("custom_db_query_duration_seconds") // 메트릭 이름
            .tag("operation", operation)
            .tag("table", table)
            .tag("outcome", outcome)       // 성공/실패 여부 태그 추가
            .tag("exception", exception)   // 발생한 예외 종류 태그 추가
            .tag("application", "wealist-api-dev")
            .description("Duration of database queries")
            .publishPercentileHistogram()
            .register(meterRegistry)
            .record(duration);
    }
}