-- 기본 권한 데이터 INSERT
INSERT INTO roles (name, level, description) VALUES
('OWNER', 100, '조직장 (최고 권한)'),
('ADMIN', 50, '운영자 (관리 권한)'),
('MEMBER', 10, '일반 멤버')
ON CONFLICT (name) DO NOTHING;
