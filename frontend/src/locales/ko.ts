import { TranslationMessages } from '@/types/i18n';

/**
 * Korean (한국어) translation messages
 */
export const koMessages: TranslationMessages = {
  // Authentication & User
  auth: {
    signIn: '로그인',
    signOut: '로그아웃',
    signInWithGitHub: 'GitHub로 로그인',
    signOutConfirmation: '정말 로그아웃 하시겠습니까?',
    welcome: '환영합니다',
    welcomeBack: '다시 오신 것을 환영합니다',
    loading: '인증 중...',
    error: '인증 오류가 발생했습니다',
    unauthorized: '로그인이 필요합니다',
    sessionExpired: '세션이 만료되었습니다. 다시 로그인해주세요',
  },
  
  // Dashboard & Navigation
  dashboard: {
    title: '대시보드',
    repositories: '저장소',
    search: '검색',
    searchPlaceholder: '저장소를 검색하세요...',
    noRepositories: '저장소가 없습니다',
    loading: '로딩 중...',
    error: '오류가 발생했습니다',
    refresh: '새로고침',
    filter: {
      all: '전체',
      connected: '연결됨',
      notConnected: '연결되지 않음',
      byLanguage: '언어별 필터',
    },
  },
  
  // Repository management
  repository: {
    connect: '연결',
    disconnect: '연결 해제',
    connecting: '연결 중...',
    connected: '연결됨',
    notConnected: '연결되지 않음',
    lastUpdated: '마지막 업데이트',
    createdAt: '생성일',
    language: '언어',
    stars: '스타',
    forks: '포크',
    private: '비공개',
    public: '공개',
    selectRepository: '저장소 선택',
    connectionSuccess: '저장소가 성공적으로 연결되었습니다',
    connectionError: '저장소 연결에 실패했습니다',
  },
  
  // Activity & Logging
  activity: {
    title: '활동 내역',
    recent: '최근 활동',
    noActivity: '활동 내역이 없습니다',
    loading: '활동 내역을 불러오는 중...',
    refresh: '활동 내역 새로고침',
    viewAll: '전체 보기',
    githubSync: 'GitHub 동기화',
    apiCall: 'API 호출',
    repositoryConnection: '저장소 연결',
    userAction: '사용자 동작',
    systemEvent: '시스템 이벤트',
  },
  
  // GitHub Integration
  github: {
    connecting: 'GitHub에 연결하는 중...',
    syncingRepositories: '저장소 동기화 중...',
    fetchingIssues: '이슈를 가져오는 중...',
    fetchingPullRequests: '풀 리퀘스트를 가져오는 중...',
    rateLimit: 'GitHub API 속도 제한',
    rateLimitWarning: 'API 속도 제한이 곧 초과됩니다',
    apiError: 'GitHub API 오류가 발생했습니다',
    unauthorized: 'GitHub 인증이 필요합니다',
    repositoryNotFound: '저장소를 찾을 수 없습니다',
  },
  
  // Theme & Settings
  theme: {
    light: '라이트 모드',
    dark: '다크 모드',
    system: '시스템 설정',
    toggleTheme: '테마 전환',
    themeChanged: '테마가 변경되었습니다',
  },
  
  // Common UI elements
  common: {
    save: '저장',
    cancel: '취소',
    delete: '삭제',
    edit: '편집',
    view: '보기',
    close: '닫기',
    back: '뒤로',
    next: '다음',
    previous: '이전',
    loading: '로딩 중...',
    error: '오류',
    success: '성공',
    warning: '경고',
    info: '정보',
    retry: '다시 시도',
    confirm: '확인',
    yes: '예',
    no: '아니오',
    settings: '설정',
    language: '언어',
    filter: '필터',
  },
  
  // Error messages
  errors: {
    generic: '알 수 없는 오류가 발생했습니다',
    network: '네트워크 연결을 확인해주세요',
    unauthorized: '권한이 없습니다',
    forbidden: '접근이 거부되었습니다',
    notFound: '요청한 리소스를 찾을 수 없습니다',
    serverError: '서버 오류가 발생했습니다',
    validationError: '입력값을 확인해주세요',
    requiredField: '필수 입력 항목입니다',
    invalidFormat: '올바른 형식이 아닙니다',
  },
  
  // Success messages  
  success: {
    saved: '저장되었습니다',
    updated: '업데이트되었습니다',
    deleted: '삭제되었습니다',
    connected: '연결되었습니다',
    disconnected: '연결이 해제되었습니다',
    refreshed: '새로고침되었습니다',
  },

  // Activity logging messages
  logs: {
    title: '활동 로그',
    noLogs: '활동 로그가 없습니다',
    noLogsMessage: '필터를 조정하시거나 작업과 저장소 상호작용이 발생하면 활동 로그가 나타납니다',
    showingLogs: '{{count}}개의 로그를 표시하고 있습니다',
    exportTitle: '내보내기',
    clearTitle: '지우기',
    clearConfirmation: '모든 활동 로그를 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다.',
    search: '활동 로그 검색...',
    filterByType: '유형별 필터',
    selectTimeRange: '시간 범위 선택',
    
    // Time ranges
    timeRanges: {
      lastHour: '지난 1시간',
      last24Hours: '지난 24시간',
      last7Days: '지난 7일',
      last30Days: '지난 30일',
      allTime: '전체 기간'
    },

    // Activity types
    types: {
      all: '모든 활동',
      connection: '연결',
      task: '작업',
      github: 'GitHub',
      navigation: '탐색'
    },

    // Repository connection messages
    repository_connected: '{{repositoryName}} 저장소에 성공적으로 연결되었습니다',
    repository_disconnected: '{{repositoryName}} 저장소 연결이 해제되었습니다',
    repository_connection_failed: '{{repositoryName}} 저장소 연결에 실패했습니다: {{error}}',

    // Task messages
    task_created: '{{repositoryName}}에 새 작업 "{{taskTitle}}"이 생성되었습니다',
    task_started: '"{{taskTitle}}" 작업 실행이 시작되었습니다',
    task_completed: '"{{taskTitle}}" 작업이 성공적으로 완료되었습니다',
    task_failed: '"{{taskTitle}}" 작업이 실패했습니다: {{error}}',
    task_cancelled: '"{{taskTitle}}" 작업이 취소되었습니다',

    // GitHub messages
    github_sync_started: '{{repositoryName}}에 대한 GitHub 동기화가 시작되었습니다',
    github_sync_completed: '{{repositoryName}}에 대한 GitHub 동기화가 완료되었습니다',
    github_sync_failed: '{{repositoryName}}에 대한 GitHub 동기화가 실패했습니다',
    github_api_call: 'GitHub API 호출: {{method}} {{endpoint}}',
    github_rate_limit: 'GitHub API 속도 제한: {{remaining}}개 요청 남음. {{resetTime}}에 재설정',
    github_rate_limit_warning: 'GitHub API 속도 제한이 곧 초과됩니다',
    github_rate_limit_exceeded: 'GitHub API 속도 제한이 초과되었습니다',

    // Navigation messages
    tab_switched: '{{repositoryName}}에서 {{previousTab}} 탭에서 {{currentTab}} 탭으로 전환',
    workspace_opened: '{{repositoryName}} 작업공간에 접근했습니다',
    workspace_closed: '{{repositoryName}} 작업공간을 닫았습니다',

    // User actions
    user_actions: {
      repository_connected: '저장소 연결',
      repository_disconnected: '저장소 연결 해제',
      repository_connection_failed: '저장소 연결 실패',
      task_created: '작업 생성',
      task_started: '작업 시작',
      task_completed: '작업 완료',
      task_failed: '작업 실패',
      task_cancelled: '작업 취소',
      github_sync_started: 'GitHub 동기화 시작',
      github_sync_completed: 'GitHub 동기화 완료',
      github_sync_failed: 'GitHub 동기화 실패',
      github_api_call: 'GitHub API 호출',
      tab_switched: '탭 전환',
      workspace_opened: '작업공간 열기',
      workspace_closed: '작업공간 닫기',
    }
  },
};