/**
 * Korean comment templates for GitHub issues
 */

export interface KoreanCommentTemplate {
  emoji: string;
  title: string;
  content: string;
}

export type CommentType = 'start' | 'progress' | 'complete' | 'blocked' | 'review';

export interface KoreanCommentTemplates {
  [key: string]: KoreanCommentTemplate;
}

export const DEFAULT_KOREAN_TEMPLATES: KoreanCommentTemplates = {
  start: {
    emoji: '🚀',
    title: '작업을 시작합니다',
    content: `이 이슈 해결을 위한 작업을 시작하겠습니다.

**작업 계획:**
- 요구사항 분석 및 설계
- 구현 및 테스트
- 코드 리뷰 및 문서화

진행 상황을 지속적으로 업데이트하겠습니다.`
  },
  
  progress: {
    emoji: '⏳',
    title: '작업이 진행 중입니다',
    content: `현재 작업을 진행하고 있습니다.

**현재 상황:** {{status}}

**완료된 작업:**
{{completed_tasks}}

**다음 단계:**
{{next_steps}}`
  },
  
  complete: {
    emoji: '✅',
    title: '작업이 완료되었습니다',
    content: `모든 작업이 성공적으로 완료되었습니다.

**구현 내용:**
{{implementation_details}}

**테스트 결과:**
{{test_results}}

코드 리뷰를 부탁드립니다. 🙏`
  },
  
  blocked: {
    emoji: '🚧',
    title: '작업이 차단되었습니다',
    content: `작업 진행 중 다음과 같은 문제가 발생했습니다:

**차단 사유:**
{{blocking_reason}}

**해결 방안:**
{{solution_approach}}

지원이나 추가 정보가 필요합니다.`
  },
  
  review: {
    emoji: '👀',
    title: '리뷰 요청',
    content: `작업이 완료되어 리뷰를 요청드립니다.

**변경 사항:**
{{changes_summary}}

**테스트 완료:**
{{test_coverage}}

**확인 사항:**
{{review_points}}`
  }
};

/**
 * Template placeholder replacement
 */
export interface TemplateVariables {
  status?: string;
  completed_tasks?: string;
  next_steps?: string;
  implementation_details?: string;
  test_results?: string;
  blocking_reason?: string;
  solution_approach?: string;
  changes_summary?: string;
  test_coverage?: string;
  review_points?: string;
  [key: string]: string | undefined;
}

/**
 * Replace placeholders in template content with actual values
 */
export function replaceTemplateVariables(
  template: string, 
  variables: TemplateVariables
): string {
  let result = template;
  
  Object.entries(variables).forEach(([key, value]) => {
    if (value !== undefined) {
      const placeholder = `{{${key}}}`;
      result = result.replace(new RegExp(placeholder, 'g'), value);
    }
  });
  
  // Remove unused placeholders
  result = result.replace(/\{\{[^}]+\}\}/g, '');
  
  // Clean up empty lines and excessive whitespace
  result = result
    .split('\n')
    .map(line => line.trim())
    .filter(line => line !== '')
    .join('\n');
  
  return result;
}

/**
 * Generate Korean comment for GitHub issue
 */
export function generateKoreanComment(
  type: CommentType,
  variables?: TemplateVariables,
  customTemplates?: KoreanCommentTemplates
): string {
  const templates = customTemplates || DEFAULT_KOREAN_TEMPLATES;
  const template = templates[type];
  
  if (!template) {
    throw new Error(`Unknown comment type: ${type}`);
  }
  
  const title = `${template.emoji} **${template.title}**`;
  const content = variables 
    ? replaceTemplateVariables(template.content, variables)
    : template.content;
  
  return `${title}\n\n${content}`;
}