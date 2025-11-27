package OrangeCloud.UserRepo.config;
import javax.sql.DataSource;
import OrangeCloud.UserRepo.repository.UserRepository;
import OrangeCloud.UserRepo.repository.WorkspaceRepository;
import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.Gauge;
import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Lazy;

@Configuration
public class MetricsConfig {

    @Bean
    public Counter httpRequestTotalCounter(MeterRegistry meterRegistry) {
        return Counter.builder("user_service_http_requests_total")
                .description("Total number of HTTP requests.")
                .register(meterRegistry);
    }

    @Bean
    public Timer dbQueryDurationTimer(MeterRegistry meterRegistry) {
        return Timer.builder("user_service_db_query_duration_seconds")
                .description("Duration of database queries.")
                .publishPercentileHistogram()
                .register(meterRegistry);
    }

    @Bean
    public Counter dbQueryErrorCounter(MeterRegistry meterRegistry) {
        return Counter.builder("user_service_db_query_errors_total")
                .description("Total number of database query errors.")
                .register(meterRegistry);
    }

    @Bean
    public Counter externalApiErrorCounter(MeterRegistry meterRegistry) {
        return Counter.builder("user_service_external_api_errors_total")
                .description("Total number of external API errors.")
                .register(meterRegistry);
    }

    @Bean
    public Counter userSignupTotalCounter(MeterRegistry meterRegistry) {
        return Counter.builder("user_service_user_signup_total")
                .description("Total number of user signups.")
                .register(meterRegistry);
    }

    @Bean
    public Counter workspaceCreatedTotalCounter(MeterRegistry meterRegistry) {
        return Counter.builder("user_service_workspace_created_total")
                .description("Total number of workspace creations.")
                .register(meterRegistry);
    }
    @Lazy(false)
    @Bean
    public Gauge usersTotalGauge(MeterRegistry meterRegistry, UserRepository userRepository) {
        return Gauge.builder("user_service_users_total", userRepository, ur -> (double) ur.count())
                .description("Total number of users.")
                .register(meterRegistry);
    }
    @Lazy(false)
    @Bean
    public Gauge workspacesTotalGauge(MeterRegistry meterRegistry, WorkspaceRepository workspaceRepository) {
        return Gauge.builder("user_service_workspaces_total", workspaceRepository, wr -> (double) wr.count())
                .description("Total number of workspaces.")
                .register(meterRegistry);
    }
    @Bean
    public Gauge activeDbConnections(MeterRegistry meterRegistry, DataSource dataSource) {
        return Gauge.builder("user_service_db_connections_active", dataSource, ds -> {
            try {
                // HikariCP 전용
                if (ds instanceof com.zaxxer.hikari.HikariDataSource hikari) {
                    return hikari.getHikariPoolMXBean().getActiveConnections();
                }
            } catch (Exception e) {
                return 0;
            }
            return 0;
        }).description("Number of active DB connections").register(meterRegistry);
    }
    @Bean
    public DatabaseMetrics databaseMetrics(MeterRegistry meterRegistry) {
        return new DatabaseMetrics(meterRegistry);
    }
}
