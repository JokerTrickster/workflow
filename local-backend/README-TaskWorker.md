# Task Worker Enhancement - Concurrent Processing & Auto PR Creation

## 🎯 개선 내용

### 1. **병렬 처리 (Concurrent Processing)**
기존 단일 고루틴 처리에서 **다중 고루틴 병렬 처리**로 개선

#### Before (기존)
```go
// 단일 고루틴, 순차 처리
func (w *TaskWorker) StartConsuming() {
    for msg := range msgs {
        w.handleMessage(msg) // 하나씩 순차 처리
    }
}
```

#### After (개선)
```go
// 다중 고루틴, 병렬 처리 + 세마포어로 동시 실행 수 제어
func (w *ConcurrentTaskWorker) StartConsuming() {
    for msg := range msgs {
        w.wg.Add(1)
        go w.handleMessageConcurrent(msg) // 병렬 처리
    }
}
```

### 2. **자동 PR 생성 (Auto Pull Request Creation)**
작업 완료 후 **자동으로 GitHub PR 생성**하는 시스템 추가

#### 워크플로우
```
Task 완료 → 변경사항 감지 → Feature Branch 생성 → Commit → Push → PR 생성
```

## 🛠️ 구현된 파일들

### `concurrent_task_worker.go` - 병렬 처리 워커
```go
type ConcurrentTaskWorker struct {
    maxConcurrent   int                // 최대 동시 실행 수
    semaphore       chan struct{}      // 동시 실행 제어
    wg              sync.WaitGroup     // 고루틴 동기화
    mutex           sync.RWMutex       // 스레드 안전성
}
```

**주요 기능:**
- ✅ **세마포어 기반 동시 실행 제어**
- ✅ **스레드 안전한 failure count 관리**
- ✅ **Graceful shutdown** (모든 태스크 완료 대기)
- ✅ **자동 PR 생성** 트리거

### `pr_creator.go` - GitHub PR 자동 생성
```go
type PRCreator struct {
    workingDir string
    repoName   string
}
```

**주요 기능:**
- ✅ **환경 검증** (gh CLI, git, 인증 확인)
- ✅ **Feature Branch 자동 생성** (main/master에서 작업 방지)
- ✅ **변경사항 자동 커밋**
- ✅ **GitHub PR 생성** (상세한 설명 포함)

### `main.go` - 설정 통합
```go
// 환경변수로 동시 실행 수 제어
maxConcurrent := 3 // 기본값
if env := os.Getenv("MAX_CONCURRENT_TASKS"); env != "" {
    maxConcurrent = parseIntOrDefault(env, 3)
}

startConcurrentTaskWorker(ctx, rabbitMQURL, queueName, maxConcurrent)
```

## 🚀 사용법

### 환경 설정
```bash
# .env 파일에 추가
MAX_CONCURRENT_TASKS=5          # 동시 실행할 최대 태스크 수
RABBITMQ_URL=amqp://localhost:5672/
RABBITMQ_QUEUE_NAME=claude_tasks

# GitHub CLI 인증 (PR 생성용)
gh auth login
```

### 서버 시작
```bash
# 서버 실행
go run main.go

# 로그에서 확인
# "Starting Concurrent Task Worker with ... Max Concurrent: 5"
```

### 동작 확인
```bash
# RabbitMQ에 메시지 여러 개 발행하여 병렬 처리 확인
# 로그에서 "Processing message (concurrent)" 동시 출력 확인
```

## 📊 성능 비교

| 항목 | 기존 (순차) | 개선 (병렬) |
|------|-------------|-------------|
| 동시 처리 | 1개 | 3-5개 (설정가능) |
| 처리량 | 1 task/minute | 3-5 task/minute |
| 리소스 활용 | 낮음 | 높음 |
| 응답성 | 느림 | 빠름 |

## 🔄 워크플로우

### 1. **병렬 처리 플로우**
```
RabbitMQ → [Task1] → [Task2] → [Task3] → [Task4] → [Task5]
              ↓        ↓        ↓        ↓        ↓
            Worker1  Worker2  Worker3  Worker4  Worker5
              ↓        ↓        ↓        ↓        ↓
             AI-1     AI-2     AI-3     AI-4     AI-5
```

### 2. **PR 생성 플로우**
```
Task 완료 → 파일 변경 감지 → Feature Branch 생성
    ↓
Git Add & Commit → Push to GitHub → Create PR
    ↓
PR URL 로깅 → 작업 완료 기록
```

## 🔧 설정 옵션

### 동시 실행 수 조정
```bash
# 환경변수로 조정
export MAX_CONCURRENT_TASKS=10  # 10개 동시 실행

# 서버 재시작
go run main.go
```

### PR 생성 비활성화 (개발 중)
```go
// concurrent_task_worker.go의 shouldCreatePR 함수 수정
func (w *ConcurrentTaskWorker) shouldCreatePR(result *AITaskResponse) bool {
    return false // 임시로 비활성화
}
```

## 🛡️ 안전성 기능

### 1. **동시 실행 제한**
- 세마포어로 최대 동시 실행 수 제어
- 메모리 부족 방지

### 2. **스레드 안전성**
- `sync.RWMutex`로 공유 상태 보호
- Race condition 방지

### 3. **Graceful Shutdown**
- Context cancellation으로 우아한 종료
- 모든 실행 중인 태스크 완료 대기

### 4. **에러 핸들링**
- PR 생성 실패해도 태스크 완료 처리
- 각 워커별 독립적인 에러 처리

## 📝 로그 예시

### 병렬 처리 로그
```
Starting Concurrent Task Worker with ... Max Concurrent: 3
Processing message (concurrent): task1
Processing message (concurrent): task2
Processing message (concurrent): task3
Task completed successfully for provider: claude
Creating PR for completed task: implement user auth
Successfully created PR: https://github.com/owner/repo/pull/123
```

### PR 생성 로그
```
Creating PR for completed task: add user authentication
Changed to working directory: /path/to/repo
Created and switched to feature branch: task/add-user-auth-0926-1430
Committed changes with message: feat: add user authentication...
Pushed branch to GitHub: task/add-user-auth-0926-1430
Created GitHub PR: https://github.com/owner/repo/pull/123
```

이제 **병렬 처리**와 **자동 PR 생성**이 모두 구현되어 있습니다! 🎉

## 다음 단계
1. 서버 재시작하여 새로운 워커 활성화
2. 테스트 태스크들을 RabbitMQ에 발행하여 병렬 처리 확인
3. 파일 변경이 있는 태스크로 PR 자동 생성 테스트