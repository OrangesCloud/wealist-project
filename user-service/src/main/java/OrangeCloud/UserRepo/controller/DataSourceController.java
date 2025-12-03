package OrangeCloud.UserRepo.controller;

import com.zaxxer.hikari.HikariConfig;
import com.zaxxer.hikari.HikariDataSource;
import lombok.Data;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.autoconfigure.jdbc.DataSourceProperties;
import org.springframework.context.ApplicationContext;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.core.env.ConfigurableEnvironment;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.sql.DataSource;
import java.sql.Connection;
import java.sql.SQLException;
import java.util.HashMap;
import java.util.Map;

/**
 * 런타임에 DataSource 연결 정보를 동적으로 설정/변경하는 컨트롤러
 * 
 * 주의: 프로덕션 환경에서는 보안을 위해 인증/인가 필수!
 */
@Slf4j
@RestController
@RequestMapping("/admin/datasource")
public class DataSourceController {
    
    @Autowired(required = false)
    private DataSource dataSource;
    
    @Autowired
    private ApplicationContext applicationContext;
    
    /**
     * 현재 DataSource 연결 상태 확인
     */
    @GetMapping("/status")
    public ResponseEntity<Map<String, Object>> getStatus() {
        Map<String, Object> status = new HashMap<>();
        
        if (dataSource == null) {
            status.put("connected", false);
            status.put("message", "DataSource not configured (no-database profile)");
            return ResponseEntity.ok(status);
        }
        
        try (Connection conn = dataSource.getConnection()) {
            status.put("connected", true);
            status.put("url", conn.getMetaData().getURL());
            status.put("username", conn.getMetaData().getUserName());
            status.put("databaseProduct", conn.getMetaData().getDatabaseProductName());
            status.put("databaseVersion", conn.getMetaData().getDatabaseProductVersion());
            
            if (dataSource instanceof HikariDataSource) {
                HikariDataSource hikari = (HikariDataSource) dataSource;
                status.put("activeConnections", hikari.getHikariPoolMXBean().getActiveConnections());
                status.put("idleConnections", hikari.getHikariPoolMXBean().getIdleConnections());
                status.put("totalConnections", hikari.getHikariPoolMXBean().getTotalConnections());
                status.put("maxPoolSize", hikari.getMaximumPoolSize());
            }
            
            return ResponseEntity.ok(status);
        } catch (SQLException e) {
            log.error("Failed to get database connection", e);
            status.put("connected", false);
            status.put("error", e.getMessage());
            return ResponseEntity.status(503).body(status);
        }
    }
    
    /**
     * DataSource 연결 테스트
     */
    @GetMapping("/test")
    public ResponseEntity<Map<String, Object>> testConnection() {
        Map<String, Object> result = new HashMap<>();
        
        if (dataSource == null) {
            result.put("success", false);
            result.put("message", "DataSource not configured");
            return ResponseEntity.ok(result);
        }
        
        long startTime = System.currentTimeMillis();
        try (Connection conn = dataSource.getConnection()) {
            long duration = System.currentTimeMillis() - startTime;
            
            result.put("success", true);
            result.put("connectionTime", duration + "ms");
            result.put("message", "Database connection successful");
            
            return ResponseEntity.ok(result);
        } catch (SQLException e) {
            long duration = System.currentTimeMillis() - startTime;
            
            log.error("Database connection test failed", e);
            result.put("success", false);
            result.put("connectionTime", duration + "ms");
            result.put("error", e.getMessage());
            
            return ResponseEntity.status(503).body(result);
        }
    }
    
    /**
     * DataSource 설정 정보 업데이트
     * 
     * 주의: 이 방법은 애플리케이션 재시작이 필요합니다.
     * 실제 프로덕션에서는 ConfigMap/Secret 업데이트 + Pod 재시작 권장
     */
    @PostMapping("/update")
    public ResponseEntity<Map<String, Object>> updateDataSource(
            @RequestBody DataSourceUpdateRequest request) {
        
        Map<String, Object> result = new HashMap<>();
        
        try {
            // 환경변수 업데이트 (재시작 시 반영됨)
            if (applicationContext instanceof ConfigurableApplicationContext) {
                ConfigurableEnvironment env = 
                    ((ConfigurableApplicationContext) applicationContext).getEnvironment();
                
                System.setProperty("spring.datasource.url", request.getUrl());
                System.setProperty("spring.datasource.username", request.getUsername());
                System.setProperty("spring.datasource.password", request.getPassword());
                
                result.put("success", true);
                result.put("message", "DataSource configuration updated. Please restart the application.");
                result.put("requiresRestart", true);
                
                log.info("DataSource configuration updated: url={}, username={}", 
                         request.getUrl(), request.getUsername());
                
                return ResponseEntity.ok(result);
            } else {
                result.put("success", false);
                result.put("message", "Unable to update configuration");
                return ResponseEntity.status(500).body(result);
            }
            
        } catch (Exception e) {
            log.error("Failed to update DataSource configuration", e);
            result.put("success", false);
            result.put("error", e.getMessage());
            return ResponseEntity.status(500).body(result);
        }
    }
    
    /**
     * 현재 활성화된 Spring Profile 확인
     */
    @GetMapping("/profile")
    public ResponseEntity<Map<String, Object>> getActiveProfile() {
        Map<String, Object> result = new HashMap<>();
        
        if (applicationContext instanceof ConfigurableApplicationContext) {
            ConfigurableEnvironment env = 
                ((ConfigurableApplicationContext) applicationContext).getEnvironment();
            
            result.put("activeProfiles", env.getActiveProfiles());
            result.put("defaultProfiles", env.getDefaultProfiles());
            result.put("databaseConfigured", dataSource != null);
        }
        
        return ResponseEntity.ok(result);
    }
    
    /**
     * HikariCP 풀 통계 조회
     */
    @GetMapping("/pool/stats")
    public ResponseEntity<Map<String, Object>> getPoolStats() {
        Map<String, Object> stats = new HashMap<>();
        
        if (dataSource == null) {
            stats.put("error", "DataSource not configured");
            return ResponseEntity.ok(stats);
        }
        
        if (dataSource instanceof HikariDataSource) {
            HikariDataSource hikari = (HikariDataSource) dataSource;
            
            stats.put("activeConnections", hikari.getHikariPoolMXBean().getActiveConnections());
            stats.put("idleConnections", hikari.getHikariPoolMXBean().getIdleConnections());
            stats.put("totalConnections", hikari.getHikariPoolMXBean().getTotalConnections());
            stats.put("threadsAwaitingConnection", hikari.getHikariPoolMXBean().getThreadsAwaitingConnection());
            stats.put("maxPoolSize", hikari.getMaximumPoolSize());
            stats.put("minIdle", hikari.getMinimumIdle());
            
            return ResponseEntity.ok(stats);
        } else {
            stats.put("error", "Not a HikariCP DataSource");
            return ResponseEntity.ok(stats);
        }
    }
    
    /**
     * DataSource 업데이트 요청 DTO
     */
    @Data
    public static class DataSourceUpdateRequest {
        private String url;
        private String username;
        private String password;
    }
}