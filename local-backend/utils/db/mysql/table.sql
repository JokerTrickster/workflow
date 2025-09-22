CREATE TABLE tokens (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    refresh_expired_at INT,
    user_id INT,
    access_token VARCHAR(255),
    refresh_token VARCHAR(255)
);

-- Workflow History Table (단순화된 버전)
-- 프론트엔드에서 보여주는 정보만 저장

CREATE TABLE IF NOT EXISTS workflow_histories (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    request_id VARCHAR(255) NOT NULL UNIQUE COMMENT '요청 ID (프론트엔드 표시용)',
    status ENUM('pending', 'processing', 'completed', 'failed', 'cancelled') NOT NULL DEFAULT 'pending' COMMENT '작업 상태',
    
    -- 요청 정보 (프론트엔드 입력값)
    tasks TEXT NOT NULL COMMENT '작업 내용',
    repository_name VARCHAR(255) NOT NULL COMMENT '저장소 이름',
    working_dir VARCHAR(500) COMMENT '작업 디렉토리 경로',
    claude_cmd VARCHAR(1000) COMMENT 'Claude 명령어',
    interactive BOOLEAN DEFAULT FALSE COMMENT '인터랙티브 모드 여부',
    continue_task BOOLEAN DEFAULT FALSE COMMENT '작업 계속 여부',
    
    -- 실행 시간 정보
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '생성 시간',
    completed_at TIMESTAMP NULL COMMENT '완료 시간',
    processing_time_ms BIGINT COMMENT '처리 시간(밀리초)',
    
    -- 결과 정보
    result TEXT COMMENT '작업 결과 (JSON 형태)',
    error TEXT COMMENT '오류 메시지',
    
    -- 인덱스
    INDEX idx_request_id (request_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_repository_name (repository_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='워크플로우 작업 히스토리 (단순화)';
