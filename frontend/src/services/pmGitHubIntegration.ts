import { GitHubApiService } from './githubApi';
import { CommentType, TemplateVariables } from '@/types/korean-templates';
import { ActivityLogger } from './ActivityLogger';

/**
 * PM-GitHub Integration Service
 * Handles Korean comment creation for PM system workflows
 */
export class PMGitHubIntegrationService {
  /**
   * Create a Korean start comment when beginning work on an issue
   * @param issueNumber - GitHub issue number
   * @param repoId - Repository ID (format: "owner/repo")
   * @param taskTitle - Optional task title for context
   */
  static async createIssueStartComment(
    issueNumber: number,
    repoId: string,
    taskTitle?: string
  ): Promise<void> {
    const activityLogger = ActivityLogger.getInstance();
    
    try {
      // Prepare template variables for start comment
      const variables: TemplateVariables = {};
      
      if (taskTitle) {
        variables.status = `${taskTitle} 작업 분석 및 설계`;
      }
      
      // Create Korean start comment
      await GitHubApiService.createKoreanIssueComment(
        repoId,
        issueNumber,
        'start',
        variables
      );
      
      activityLogger.logActivity(
        'PM Issue Start',
        `Started work on issue #${issueNumber} with Korean comment`,
        'success'
      );
      
      console.log(`✅ Created Korean start comment for issue #${issueNumber}`);
    } catch (error) {
      activityLogger.logActivity(
        'PM Issue Start',
        `Failed to create start comment for issue #${issueNumber}: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'error'
      );
      
      console.error(`❌ Failed to create Korean start comment for issue #${issueNumber}:`, error);
      throw error;
    }
  }
  
  /**
   * Create a Korean progress comment during issue work
   * @param issueNumber - GitHub issue number  
   * @param repoId - Repository ID (format: "owner/repo")
   * @param status - Current status
   * @param completedTasks - List of completed tasks
   * @param nextSteps - Upcoming work items
   */
  static async createIssueProgressComment(
    issueNumber: number,
    repoId: string,
    status: string,
    completedTasks: string[] = [],
    nextSteps: string[] = []
  ): Promise<void> {
    const variables: TemplateVariables = {
      status,
      completed_tasks: completedTasks.map(task => `- ${task}`).join('\n'),
      next_steps: nextSteps.map(step => `- ${step}`).join('\n')
    };
    
    try {
      await GitHubApiService.createKoreanIssueComment(
        repoId,
        issueNumber,
        'progress',
        variables
      );
      
      console.log(`✅ Created Korean progress comment for issue #${issueNumber}`);
    } catch (error) {
      console.error(`❌ Failed to create Korean progress comment for issue #${issueNumber}:`, error);
      throw error;
    }
  }
  
  /**
   * Create a Korean completion comment when finishing issue work
   * @param issueNumber - GitHub issue number
   * @param repoId - Repository ID (format: "owner/repo") 
   * @param implementationDetails - What was implemented
   * @param testResults - Test results summary
   */
  static async createIssueCompleteComment(
    issueNumber: number,
    repoId: string,
    implementationDetails: string[] = [],
    testResults: string = '모든 테스트 통과'
  ): Promise<void> {
    const variables: TemplateVariables = {
      implementation_details: implementationDetails.map(detail => `- ${detail}`).join('\n'),
      test_results: testResults
    };
    
    try {
      await GitHubApiService.createKoreanIssueComment(
        repoId,
        issueNumber,
        'complete',
        variables
      );
      
      console.log(`✅ Created Korean completion comment for issue #${issueNumber}`);
    } catch (error) {
      console.error(`❌ Failed to create Korean completion comment for issue #${issueNumber}:`, error);
      throw error;
    }
  }
  
  /**
   * Create a Korean blocked comment when work is blocked
   * @param issueNumber - GitHub issue number
   * @param repoId - Repository ID (format: "owner/repo")
   * @param blockingReason - Why work is blocked
   * @param solutionApproach - Proposed solution
   */
  static async createIssueBlockedComment(
    issueNumber: number,
    repoId: string,
    blockingReason: string,
    solutionApproach: string = '추가 조사 및 해결 방안 모색 필요'
  ): Promise<void> {
    const variables: TemplateVariables = {
      blocking_reason: blockingReason,
      solution_approach: solutionApproach
    };
    
    try {
      await GitHubApiService.createKoreanIssueComment(
        repoId,
        issueNumber,
        'blocked',
        variables
      );
      
      console.log(`✅ Created Korean blocked comment for issue #${issueNumber}`);
    } catch (error) {
      console.error(`❌ Failed to create Korean blocked comment for issue #${issueNumber}:`, error);
      throw error;
    }
  }
  
  /**
   * Create a Korean review comment when requesting review
   * @param issueNumber - GitHub issue number
   * @param repoId - Repository ID (format: "owner/repo") 
   * @param changesSummary - Summary of changes made
   * @param testCoverage - Test coverage information
   * @param reviewPoints - Specific review points
   */
  static async createIssueReviewComment(
    issueNumber: number,
    repoId: string,
    changesSummary: string,
    testCoverage: string = '단위 테스트 및 통합 테스트 완료',
    reviewPoints: string[] = []
  ): Promise<void> {
    const variables: TemplateVariables = {
      changes_summary: changesSummary,
      test_coverage: testCoverage,
      review_points: reviewPoints.map(point => `- ${point}`).join('\n')
    };
    
    try {
      await GitHubApiService.createKoreanIssueComment(
        repoId,
        issueNumber,
        'review',
        variables
      );
      
      console.log(`✅ Created Korean review comment for issue #${issueNumber}`);
    } catch (error) {
      console.error(`❌ Failed to create Korean review comment for issue #${issueNumber}:`, error);
      throw error;
    }
  }
}