import {
  generateKoreanComment,
  replaceTemplateVariables,
  DEFAULT_KOREAN_TEMPLATES,
  TemplateVariables,
  CommentType
} from '@/types/korean-templates';

describe('Korean Templates', () => {
  describe('replaceTemplateVariables', () => {
    it('should replace simple placeholders', () => {
      const template = 'Hello {{name}}, welcome to {{project}}!';
      const variables = { name: 'Claude', project: 'AI Workbench' };
      
      const result = replaceTemplateVariables(template, variables);
      
      expect(result).toBe('Hello Claude, welcome to AI Workbench!');
    });

    it('should handle multiline templates', () => {
      const template = `**Status:** {{status}}

**Tasks:**
{{tasks}}

**Notes:**
{{notes}}`;
      
      const variables = {
        status: 'In Progress',
        tasks: '- Implement feature\n- Write tests',
        notes: 'All going well'
      };
      
      const result = replaceTemplateVariables(template, variables);
      
      expect(result).toContain('**Status:** In Progress');
      expect(result).toContain('- Implement feature');
    });

    it('should remove unused placeholders', () => {
      const template = 'Hello {{name}}! {{unused_placeholder}} How are you?';
      const variables = { name: 'Claude' };
      
      const result = replaceTemplateVariables(template, variables);
      
      expect(result).toBe('Hello Claude! How are you?');
      expect(result).not.toContain('{{unused_placeholder}}');
    });

    it('should handle empty variables', () => {
      const template = 'Hello {{name}}!';
      const variables = {};
      
      const result = replaceTemplateVariables(template, variables);
      
      expect(result).toBe('Hello !');
    });
  });

  describe('generateKoreanComment', () => {
    it('should generate start comment', () => {
      const result = generateKoreanComment('start');
      
      expect(result).toContain('🚀 **작업을 시작합니다**');
      expect(result).toContain('이 이슈 해결을 위한 작업을 시작하겠습니다');
      expect(result).toContain('**작업 계획:**');
    });

    it('should generate progress comment with variables', () => {
      const variables: TemplateVariables = {
        status: '기능 구현 중',
        completed_tasks: '- API 설계 완료\n- 테스트 케이스 작성',
        next_steps: '- 구현 완료\n- 통합 테스트'
      };
      
      const result = generateKoreanComment('progress', variables);
      
      expect(result).toContain('⏳ **작업이 진행 중입니다**');
      expect(result).toContain('기능 구현 중');
      expect(result).toContain('API 설계 완료');
      expect(result).toContain('구현 완료');
    });

    it('should generate complete comment', () => {
      const variables: TemplateVariables = {
        implementation_details: '- GitHub API 연동 완료\n- 한글 템플릿 시스템 구축',
        test_results: '모든 테스트 통과 (100% 커버리지)'
      };
      
      const result = generateKoreanComment('complete', variables);
      
      expect(result).toContain('✅ **작업이 완료되었습니다**');
      expect(result).toContain('GitHub API 연동 완료');
      expect(result).toContain('모든 테스트 통과');
      expect(result).toContain('코드 리뷰를 부탁드립니다');
    });

    it('should generate blocked comment', () => {
      const variables: TemplateVariables = {
        blocking_reason: 'GitHub API 권한 부족',
        solution_approach: '관리자에게 권한 요청 필요'
      };
      
      const result = generateKoreanComment('blocked', variables);
      
      expect(result).toContain('🚧 **작업이 차단되었습니다**');
      expect(result).toContain('GitHub API 권한 부족');
      expect(result).toContain('관리자에게 권한 요청 필요');
    });

    it('should generate review comment', () => {
      const variables: TemplateVariables = {
        changes_summary: 'Korean comment system implemented',
        test_coverage: '98% test coverage achieved',
        review_points: '- Code quality check\n- Performance validation'
      };
      
      const result = generateKoreanComment('review', variables);
      
      expect(result).toContain('👀 **리뷰 요청**');
      expect(result).toContain('Korean comment system implemented');
      expect(result).toContain('98% test coverage achieved');
    });

    it('should use custom templates', () => {
      const customTemplates = {
        custom: {
          emoji: '🎯',
          title: '커스텀 메시지',
          content: '이것은 {{custom_var}} 메시지입니다.'
        }
      };
      
      const result = generateKoreanComment(
        'custom' as CommentType,
        { custom_var: '테스트' },
        customTemplates
      );
      
      expect(result).toContain('🎯 **커스텀 메시지**');
      expect(result).toContain('이것은 테스트 메시지입니다');
    });

    it('should throw error for unknown comment type', () => {
      expect(() => {
        generateKoreanComment('unknown' as CommentType);
      }).toThrow('Unknown comment type: unknown');
    });
  });

  describe('DEFAULT_KOREAN_TEMPLATES', () => {
    it('should have all required comment types', () => {
      const requiredTypes: CommentType[] = ['start', 'progress', 'complete', 'blocked', 'review'];
      
      requiredTypes.forEach(type => {
        expect(DEFAULT_KOREAN_TEMPLATES[type]).toBeDefined();
        expect(DEFAULT_KOREAN_TEMPLATES[type].emoji).toBeTruthy();
        expect(DEFAULT_KOREAN_TEMPLATES[type].title).toBeTruthy();
        expect(DEFAULT_KOREAN_TEMPLATES[type].content).toBeTruthy();
      });
    });

    it('should have proper emoji for each type', () => {
      expect(DEFAULT_KOREAN_TEMPLATES.start.emoji).toBe('🚀');
      expect(DEFAULT_KOREAN_TEMPLATES.progress.emoji).toBe('⏳');
      expect(DEFAULT_KOREAN_TEMPLATES.complete.emoji).toBe('✅');
      expect(DEFAULT_KOREAN_TEMPLATES.blocked.emoji).toBe('🚧');
      expect(DEFAULT_KOREAN_TEMPLATES.review.emoji).toBe('👀');
    });
  });
});