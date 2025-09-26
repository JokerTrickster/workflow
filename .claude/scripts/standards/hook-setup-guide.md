# Claude Hook 설정 가이드

## 작업 완료 후 자동 빌드 및 검증 시스템

### 🎯 개요
Claude Hook을 사용하여 작업 완료 후 자동으로 빌드, 테스트, 검증을 수행하는 시스템을 구축했습니다.

### ⚙️ 현재 설정
`.claude/settings.local.json`에 다음이 추가되었습니다:

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/scripts/standards/post-completion-hook.sh",
            "timeout": 60
          }
        ]
      }
    ]
  }
}
```

### 🚀 작동 방식

1. **트리거**: 사용자가 프롬프트를 제출할 때마다 실행
2. **자동 감지**: 유의미한 변경사항이 있을 때만 실행
3. **검증 단계**:
   - 프로젝트 표준 검증
   - 빌드 테스트 (Frontend/Backend)
   - 테스트 실행
   - 코드 품질 검사 (Linting)
   - Git 상태 확인

### 📋 Hook 스크립트 기능

#### 변경사항 감지
- Git 상태 확인 (uncommitted changes)
- 최근 5분 내 커밋 확인
- 유의미한 변경이 없으면 스킵

#### 자동 검증
- **표준 검증**: `validate-standards.sh --fix` 실행
- **빌드 테스트**: Frontend (npm), Backend (go build)
- **테스트 실행**: 사용 가능한 모든 테스트 실행
- **코드 품질**: ESLint, Go formatting 등

#### 결과 리포트
```bash
================================
  Post-Completion Hook
================================

✅ Standards validation completed
✅ Build verification completed
✅ Test execution completed
✅ Code quality checks completed
✅ Git status reviewed

Next suggested actions:
1. Review any warnings above
2. Commit changes if satisfied
3. Create PR when feature is complete
```

### 🛠️ 추가 Hook 옵션

#### 1. 특정 도구 사용 후에만 실행
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/scripts/standards/post-completion-hook.sh"
          }
        ]
      }
    ]
  }
}
```

#### 2. 세션 종료 시 실행
```json
{
  "hooks": {
    "SessionEnd": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/scripts/standards/session-cleanup.sh"
          }
        ]
      }
    ]
  }
}
```

#### 3. 조건부 실행 (Git 상태 기반)
```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "matcher": "commit|push|완료|끝|done|finish",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/scripts/standards/post-completion-hook.sh"
          }
        ]
      }
    ]
  }
}
```

### 🔧 커스터마이징

#### Hook 스크립트 수정
`/Users/luxrobo/project/workflow/.claude/scripts/standards/post-completion-hook.sh`를 수정하여:

- 추가 빌드 단계
- 커스텀 테스트 실행
- 특정 프로젝트 요구사항
- 알림 시스템 연동

#### 환경별 설정
```bash
# 개발 환경에서만 실행
if [ "$NODE_ENV" = "development" ]; then
    run_dev_specific_checks
fi

# CI/CD 환경에서 스킵
if [ "$CI" = "true" ]; then
    exit 0
fi
```

### 🎛️ Hook 제어

#### 일시적 비활성화
```json
{
  "disableAllHooks": true
}
```

#### 특정 Hook만 비활성화
Hook 배열에서 해당 항목 제거

#### 타임아웃 조정
```json
{
  "timeout": 120  // 2분으로 연장
}
```

### 📊 성능 고려사항

- **감지 최적화**: 변경사항이 없으면 즉시 종료
- **병렬 실행**: 독립적인 검증을 병렬로 실행
- **캐싱**: 빌드 캐시 활용으로 속도 향상
- **선택적 실행**: 변경된 부분만 검증

### 🔍 디버깅

Hook 실행 로그 확인:
```bash
# Hook 실행 여부 확인
echo "Hook executed at $(date)" >> /tmp/hook-log.txt

# 자세한 로그
./.claude/scripts/standards/post-completion-hook.sh 2>&1 | tee /tmp/hook-debug.log
```

### ✅ 사용 준비 완료

현재 설정으로 다음 상황에서 자동으로 검증이 실행됩니다:
- 코드 작성/수정 후 프롬프트 제출
- 작업 완료 메시지 전송
- 유의미한 Git 변경사항 감지 시

이제 작업할 때마다 자동으로 품질 검증이 이루어집니다! 🎉