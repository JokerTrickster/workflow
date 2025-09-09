#!/usr/bin/env node

/**
 * PM Korean Comment Integration Script
 * 
 * This script demonstrates how to integrate Korean comment creation 
 * with the PM system's issue-start workflow.
 * 
 * Usage: node pmKoreanCommentIntegration.js <issue_number> [repo_id] [task_title]
 */

import { PMGitHubIntegrationService } from './pmGitHubIntegration.js';

async function createIssueStartComment(issueNumber, repoId = 'JokerTrickster/workflow', taskTitle = null) {
  try {
    console.log(`🚀 Starting Korean comment integration for issue #${issueNumber}...`);
    
    await PMGitHubIntegrationService.createIssueStartComment(
      parseInt(issueNumber),
      repoId,
      taskTitle
    );
    
    console.log(`✅ Successfully created Korean start comment for issue #${issueNumber}`);
    return true;
  } catch (error) {
    console.error(`❌ Failed to create Korean comment:`, error.message);
    return false;
  }
}

// Example usage for demonstration
async function demonstrateIntegration() {
  console.log('🧪 Demonstrating PM-Korean Comment Integration\n');
  
  const examples = [
    {
      issueNumber: 39,
      repoId: 'JokerTrickster/workflow',
      taskTitle: 'GitHub Issues에 한글 댓글 시스템 구현',
      description: 'Issue #39 start comment with task title'
    },
    {
      issueNumber: 40,
      repoId: 'JokerTrickster/workflow', 
      taskTitle: null,
      description: 'Issue #40 start comment without task title'
    }
  ];
  
  for (const example of examples) {
    console.log(`\n📝 ${example.description}:`);
    console.log(`   Issue: #${example.issueNumber}`);
    console.log(`   Repo: ${example.repoId}`);
    console.log(`   Title: ${example.taskTitle || 'N/A'}`);
    
    const success = await createIssueStartComment(
      example.issueNumber,
      example.repoId,
      example.taskTitle
    );
    
    if (success) {
      console.log(`   ✅ Success`);
    } else {
      console.log(`   ❌ Failed`);
    }
  }
}

// Command line interface
if (import.meta.url === `file://${process.argv[1]}`) {
  const args = process.argv.slice(2);
  
  if (args.length === 0) {
    console.log('🎯 PM Korean Comment Integration');
    console.log('');
    console.log('Usage:');
    console.log('  node pmKoreanCommentIntegration.js <issue_number> [repo_id] [task_title]');
    console.log('');
    console.log('Examples:');
    console.log('  node pmKoreanCommentIntegration.js 39');
    console.log('  node pmKoreanCommentIntegration.js 39 "owner/repo"');
    console.log('  node pmKoreanCommentIntegration.js 39 "owner/repo" "Task Title"');
    console.log('');
    console.log('Demo:');
    console.log('  node pmKoreanCommentIntegration.js --demo');
    process.exit(0);
  }
  
  if (args[0] === '--demo') {
    demonstrateIntegration().then(() => {
      console.log('\n🎉 Demo completed!');
    }).catch(error => {
      console.error('\n💥 Demo failed:', error);
      process.exit(1);
    });
  } else {
    const [issueNumber, repoId, taskTitle] = args;
    createIssueStartComment(issueNumber, repoId, taskTitle).then(success => {
      process.exit(success ? 0 : 1);
    });
  }
}

export { createIssueStartComment, demonstrateIntegration };