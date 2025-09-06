export interface Task {
  id: string;
  title: string;
  description: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  repository_id: number;
  created_at: string;
  updated_at: string;
  started_at?: string;
  completed_at?: string;
  branch_name?: string;
  pr_url?: string;
  build_status?: 'pending' | 'success' | 'failed';
  lint_status?: 'pending' | 'success' | 'failed';
  ai_tokens_used?: number;
  error_message?: string;
}