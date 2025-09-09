import { PMGitHubIntegrationService } from '@/services/pmGitHubIntegration';
import { GitHubApiService } from '@/services/githubApi';
import { ActivityLogger } from '@/services/ActivityLogger';

// Mock the dependencies
jest.mock('@/services/githubApi');
jest.mock('@/services/ActivityLogger');

const mockGitHubApiService = GitHubApiService as jest.Mocked<typeof GitHubApiService>;
const mockActivityLogger = {
  logActivity: jest.fn(),
  getInstance: jest.fn()
};

(ActivityLogger.getInstance as jest.Mock).mockReturnValue(mockActivityLogger);

describe('PMGitHubIntegrationService', () => {
  const testRepoId = 'test-owner/test-repo';
  const testIssueNumber = 123;

  beforeEach(() => {
    jest.clearAllMocks();
    // Mock console methods to avoid noise in tests
    jest.spyOn(console, 'log').mockImplementation();
    jest.spyOn(console, 'error').mockImplementation();
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  describe('createIssueStartComment', () => {
    it('should create Korean start comment successfully', async () => {
      mockGitHubApiService.createKoreanIssueComment.mockResolvedValue();

      await PMGitHubIntegrationService.createIssueStartComment(
        testIssueNumber,
        testRepoId,
        'Authentication System Implementation'
      );

      expect(mockGitHubApiService.createKoreanIssueComment).toHaveBeenCalledWith(
        testRepoId,
        testIssueNumber,
        'start',
        {
          status: 'Authentication System Implementation 작업 분석 및 설계'
        }
      );

      expect(mockActivityLogger.logActivity).toHaveBeenCalledWith(
        'PM Issue Start',
        `Started work on issue #${testIssueNumber} with Korean comment`,
        'success'
      );

      expect(console.log).toHaveBeenCalledWith(
        `✅ Created Korean start comment for issue #${testIssueNumber}`
      );
    });

    it('should create Korean start comment without task title', async () => {
      mockGitHubApiService.createKoreanIssueComment.mockResolvedValue();

      await PMGitHubIntegrationService.createIssueStartComment(
        testIssueNumber,
        testRepoId
      );

      expect(mockGitHubApiService.createKoreanIssueComment).toHaveBeenCalledWith(
        testRepoId,
        testIssueNumber,
        'start',
        {}
      );
    });

    it('should handle errors and rethrow', async () => {
      const error = new Error('GitHub API error');
      mockGitHubApiService.createKoreanIssueComment.mockRejectedValue(error);

      await expect(
        PMGitHubIntegrationService.createIssueStartComment(
          testIssueNumber,
          testRepoId
        )
      ).rejects.toThrow('GitHub API error');

      expect(mockActivityLogger.logActivity).toHaveBeenCalledWith(
        'PM Issue Start',
        `Failed to create start comment for issue #${testIssueNumber}: GitHub API error`,
        'error'
      );

      expect(console.error).toHaveBeenCalledWith(
        `❌ Failed to create Korean start comment for issue #${testIssueNumber}:`,
        error
      );
    });
  });

  describe('createIssueProgressComment', () => {
    it('should create Korean progress comment with all parameters', async () => {
      mockGitHubApiService.createKoreanIssueComment.mockResolvedValue();

      const completedTasks = ['API 설계 완료', '데이터베이스 스키마 구축'];
      const nextSteps = ['인증 미들웨어 구현', '테스트 케이스 작성'];

      await PMGitHubIntegrationService.createIssueProgressComment(
        testIssueNumber,
        testRepoId,
        '기능 구현 50% 완료',
        completedTasks,
        nextSteps
      );

      expect(mockGitHubApiService.createKoreanIssueComment).toHaveBeenCalledWith(
        testRepoId,
        testIssueNumber,
        'progress',
        {
          status: '기능 구현 50% 완료',
          completed_tasks: '- API 설계 완료\n- 데이터베이스 스키마 구축',
          next_steps: '- 인증 미들웨어 구현\n- 테스트 케이스 작성'
        }
      );
    });

    it('should create Korean progress comment with empty arrays', async () => {
      mockGitHubApiService.createKoreanIssueComment.mockResolvedValue();

      await PMGitHubIntegrationService.createIssueProgressComment(
        testIssueNumber,
        testRepoId,
        '분석 진행 중'
      );

      expect(mockGitHubApiService.createKoreanIssueComment).toHaveBeenCalledWith(
        testRepoId,
        testIssueNumber,
        'progress',
        {
          status: '분석 진행 중',
          completed_tasks: '',
          next_steps: ''
        }
      );
    });

    it('should handle errors in progress comment', async () => {
      const error = new Error('Network error');
      mockGitHubApiService.createKoreanIssueComment.mockRejectedValue(error);

      await expect(
        PMGitHubIntegrationService.createIssueProgressComment(
          testIssueNumber,
          testRepoId,
          'Status update'
        )
      ).rejects.toThrow('Network error');

      expect(console.error).toHaveBeenCalledWith(
        `❌ Failed to create Korean progress comment for issue #${testIssueNumber}:`,
        error
      );
    });
  });

  describe('createIssueCompleteComment', () => {
    it('should create Korean completion comment with details', async () => {
      mockGitHubApiService.createKoreanIssueComment.mockResolvedValue();

      const implementationDetails = [
        'JWT 인증 시스템 구현',
        'React Context 기반 상태 관리',
        '로그인/로그아웃 기능 완성'
      ];

      await PMGitHubIntegrationService.createIssueCompleteComment(
        testIssueNumber,
        testRepoId,
        implementationDetails,
        '100% 테스트 커버리지 달성'
      );

      expect(mockGitHubApiService.createKoreanIssueComment).toHaveBeenCalledWith(
        testRepoId,
        testIssueNumber,
        'complete',
        {
          implementation_details: '- JWT 인증 시스템 구현\n- React Context 기반 상태 관리\n- 로그인/로그아웃 기능 완성',
          test_results: '100% 테스트 커버리지 달성'
        }
      );
    });

    it('should use default test results', async () => {
      mockGitHubApiService.createKoreanIssueComment.mockResolvedValue();

      await PMGitHubIntegrationService.createIssueCompleteComment(
        testIssueNumber,
        testRepoId
      );

      expect(mockGitHubApiService.createKoreanIssueComment).toHaveBeenCalledWith(
        testRepoId,
        testIssueNumber,
        'complete',
        {
          implementation_details: '',
          test_results: '모든 테스트 통과'
        }
      );
    });

    it('should handle errors in completion comment', async () => {
      const error = new Error('API error');
      mockGitHubApiService.createKoreanIssueComment.mockRejectedValue(error);

      await expect(
        PMGitHubIntegrationService.createIssueCompleteComment(
          testIssueNumber,
          testRepoId
        )
      ).rejects.toThrow('API error');
    });
  });

  describe('createIssueBlockedComment', () => {
    it('should create Korean blocked comment with custom solution', async () => {
      mockGitHubApiService.createKoreanIssueComment.mockResolvedValue();

      await PMGitHubIntegrationService.createIssueBlockedComment(
        testIssueNumber,
        testRepoId,
        '외부 API 권한 부족',
        '관리자에게 API 키 발급 요청'
      );

      expect(mockGitHubApiService.createKoreanIssueComment).toHaveBeenCalledWith(
        testRepoId,
        testIssueNumber,
        'blocked',
        {
          blocking_reason: '외부 API 권한 부족',
          solution_approach: '관리자에게 API 키 발급 요청'
        }
      );
    });

    it('should use default solution approach', async () => {
      mockGitHubApiService.createKoreanIssueComment.mockResolvedValue();

      await PMGitHubIntegrationService.createIssueBlockedComment(
        testIssueNumber,
        testRepoId,
        '라이브러리 호환성 이슈'
      );

      expect(mockGitHubApiService.createKoreanIssueComment).toHaveBeenCalledWith(
        testRepoId,
        testIssueNumber,
        'blocked',
        {
          blocking_reason: '라이브러리 호환성 이슈',
          solution_approach: '추가 조사 및 해결 방안 모색 필요'
        }
      );
    });

    it('should handle errors in blocked comment', async () => {
      const error = new Error('Rate limit exceeded');
      mockGitHubApiService.createKoreanIssueComment.mockRejectedValue(error);

      await expect(
        PMGitHubIntegrationService.createIssueBlockedComment(
          testIssueNumber,
          testRepoId,
          'Blocking issue'
        )
      ).rejects.toThrow('Rate limit exceeded');
    });
  });

  describe('createIssueReviewComment', () => {
    it('should create Korean review comment with all details', async () => {
      mockGitHubApiService.createKoreanIssueComment.mockResolvedValue();

      const reviewPoints = [
        '보안 취약점 검토',
        '성능 최적화 확인',
        '코드 스타일 점검'
      ];

      await PMGitHubIntegrationService.createIssueReviewComment(
        testIssueNumber,
        testRepoId,
        '한글 댓글 시스템 완성',
        '98% 테스트 커버리지',
        reviewPoints
      );

      expect(mockGitHubApiService.createKoreanIssueComment).toHaveBeenCalledWith(
        testRepoId,
        testIssueNumber,
        'review',
        {
          changes_summary: '한글 댓글 시스템 완성',
          test_coverage: '98% 테스트 커버리지',
          review_points: '- 보안 취약점 검토\n- 성능 최적화 확인\n- 코드 스타일 점검'
        }
      );
    });

    it('should use default values for optional parameters', async () => {
      mockGitHubApiService.createKoreanIssueComment.mockResolvedValue();

      await PMGitHubIntegrationService.createIssueReviewComment(
        testIssueNumber,
        testRepoId,
        'Feature implementation completed'
      );

      expect(mockGitHubApiService.createKoreanIssueComment).toHaveBeenCalledWith(
        testRepoId,
        testIssueNumber,
        'review',
        {
          changes_summary: 'Feature implementation completed',
          test_coverage: '단위 테스트 및 통합 테스트 완료',
          review_points: ''
        }
      );
    });

    it('should handle errors in review comment', async () => {
      const error = new Error('Authentication failed');
      mockGitHubApiService.createKoreanIssueComment.mockRejectedValue(error);

      await expect(
        PMGitHubIntegrationService.createIssueReviewComment(
          testIssueNumber,
          testRepoId,
          'Changes ready'
        )
      ).rejects.toThrow('Authentication failed');
    });
  });

  describe('Error handling consistency', () => {
    it('should handle all comment types consistently', async () => {
      const error = new Error('Common error');
      mockGitHubApiService.createKoreanIssueComment.mockRejectedValue(error);

      const commentMethods = [
        () => PMGitHubIntegrationService.createIssueStartComment(testIssueNumber, testRepoId),
        () => PMGitHubIntegrationService.createIssueProgressComment(testIssueNumber, testRepoId, 'status'),
        () => PMGitHubIntegrationService.createIssueCompleteComment(testIssueNumber, testRepoId),
        () => PMGitHubIntegrationService.createIssueBlockedComment(testIssueNumber, testRepoId, 'reason'),
        () => PMGitHubIntegrationService.createIssueReviewComment(testIssueNumber, testRepoId, 'changes')
      ];

      for (const method of commentMethods) {
        await expect(method()).rejects.toThrow('Common error');
      }

      // All methods should have called the GitHub API
      expect(mockGitHubApiService.createKoreanIssueComment).toHaveBeenCalledTimes(5);
    });
  });

  describe('Template variable formatting', () => {
    it('should format arrays correctly for progress comments', async () => {
      mockGitHubApiService.createKoreanIssueComment.mockResolvedValue();

      const tasks = ['Task 1', 'Task 2', 'Task 3'];
      const steps = ['Step A', 'Step B'];

      await PMGitHubIntegrationService.createIssueProgressComment(
        testIssueNumber,
        testRepoId,
        'In progress',
        tasks,
        steps
      );

      expect(mockGitHubApiService.createKoreanIssueComment).toHaveBeenCalledWith(
        testRepoId,
        testIssueNumber,
        'progress',
        {
          status: 'In progress',
          completed_tasks: '- Task 1\n- Task 2\n- Task 3',
          next_steps: '- Step A\n- Step B'
        }
      );
    });

    it('should format arrays correctly for review comments', async () => {
      mockGitHubApiService.createKoreanIssueComment.mockResolvedValue();

      const reviewPoints = ['Point 1', 'Point 2'];

      await PMGitHubIntegrationService.createIssueReviewComment(
        testIssueNumber,
        testRepoId,
        'Summary',
        'Coverage',
        reviewPoints
      );

      expect(mockGitHubApiService.createKoreanIssueComment).toHaveBeenCalledWith(
        testRepoId,
        testIssueNumber,
        'review',
        {
          changes_summary: 'Summary',
          test_coverage: 'Coverage',
          review_points: '- Point 1\n- Point 2'
        }
      );
    });

    it('should format arrays correctly for completion comments', async () => {
      mockGitHubApiService.createKoreanIssueComment.mockResolvedValue();

      const details = ['Detail 1', 'Detail 2', 'Detail 3'];

      await PMGitHubIntegrationService.createIssueCompleteComment(
        testIssueNumber,
        testRepoId,
        details
      );

      expect(mockGitHubApiService.createKoreanIssueComment).toHaveBeenCalledWith(
        testRepoId,
        testIssueNumber,
        'complete',
        {
          implementation_details: '- Detail 1\n- Detail 2\n- Detail 3',
          test_results: '모든 테스트 통과'
        }
      );
    });
  });
});