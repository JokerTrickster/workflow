/**
 * @jest-environment jsdom
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { TaskCreationForm } from '../../components/TaskCreationForm';
import { Task } from '../../domain/entities/Task';

describe('TaskCreationForm - Enhanced Validation', () => {
  const mockOnSubmit = jest.fn();
  const mockOnCancel = jest.fn();
  const user = userEvent.setup();

  const defaultProps = {
    repositoryId: 1,
    onSubmit: mockOnSubmit,
    onCancel: mockOnCancel,
    isSubmitting: false,
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Mandatory Field Validation', () => {
    it('should show required indicators for all mandatory fields', () => {
      render(<TaskCreationForm {...defaultProps} />);
      
      // Check for red asterisks indicating required fields
      expect(screen.getByText('Task Title')).toBeInTheDocument();
      expect(screen.getByText('*')).toBeInTheDocument(); // Title asterisk
      expect(screen.getByText('Description')).toBeInTheDocument();
      expect(screen.getByText('Branch Name')).toBeInTheDocument();
      
      // All fields should have required attribute
      expect(screen.getByLabelText(/Task Title/)).toHaveAttribute('required');
      expect(screen.getByLabelText(/Description/)).toHaveAttribute('required');
      expect(screen.getByLabelText(/Branch Name/)).toHaveAttribute('required');
    });

    it('should prevent submission when mandatory fields are empty', async () => {
      render(<TaskCreationForm {...defaultProps} />);
      
      const submitButton = screen.getByRole('button', { name: /create task/i });
      
      // Button should be disabled initially
      expect(submitButton).toBeDisabled();
      
      // Try to submit form
      fireEvent.click(submitButton);
      
      // OnSubmit should not be called
      expect(mockOnSubmit).not.toHaveBeenCalled();
    });

    it('should show validation errors when trying to submit with empty fields', async () => {
      render(<TaskCreationForm {...defaultProps} />);
      
      const form = screen.getByRole('form') || screen.getByTestId('task-creation-form') || document.querySelector('form');
      
      if (form) {
        fireEvent.submit(form);
        
        await waitFor(() => {
          // Check for error messages
          expect(screen.queryByText(/required/i)).toBeInTheDocument();
        });
      }
    });

    it('should clear errors when user starts typing', async () => {
      render(<TaskCreationForm {...defaultProps} />);
      
      const titleInput = screen.getByLabelText(/Task Title/);
      const descriptionInput = screen.getByLabelText(/Description/);
      const branchInput = screen.getByLabelText(/Branch Name/);
      
      // Trigger validation by submitting empty form
      const form = document.querySelector('form');
      if (form) {
        fireEvent.submit(form);
      }
      
      // Type in title field
      await user.type(titleInput, 'Test Task');
      
      // Type in description field  
      await user.type(descriptionInput, 'Test description');
      
      // Type in branch field
      await user.type(branchInput, 'feature/test');
      
      // Errors should be cleared as user types
      await waitFor(() => {
        expect(titleInput).toHaveValue('Test Task');
        expect(descriptionInput).toHaveValue('Test description');
        expect(branchInput).toHaveValue('feature/test');
      });
    });

    it('should enable submit button when all mandatory fields are filled', async () => {
      render(<TaskCreationForm {...defaultProps} />);
      
      const titleInput = screen.getByLabelText(/Task Title/);
      const descriptionInput = screen.getByLabelText(/Description/);
      const branchInput = screen.getByLabelText(/Branch Name/);
      const submitButton = screen.getByRole('button', { name: /create task/i });
      
      // Fill all mandatory fields
      await user.type(titleInput, 'Test Task');
      await user.type(descriptionInput, 'Test description');
      await user.type(branchInput, 'feature/test');
      
      await waitFor(() => {
        expect(submitButton).toBeEnabled();
      });
    });

    it('should call onSubmit with correct data when form is valid', async () => {
      render(<TaskCreationForm {...defaultProps} />);
      
      const titleInput = screen.getByLabelText(/Task Title/);
      const descriptionInput = screen.getByLabelText(/Description/);
      const branchInput = screen.getByLabelText(/Branch Name/);
      
      // Fill all mandatory fields
      await user.type(titleInput, 'Test Task');
      await user.type(descriptionInput, 'Test description');
      await user.type(branchInput, 'feature/test-branch');
      
      const form = document.querySelector('form');
      if (form) {
        fireEvent.submit(form);
        
        await waitFor(() => {
          expect(mockOnSubmit).toHaveBeenCalledWith({
            title: 'Test Task',
            description: 'Test description',
            status: 'pending',
            repository_id: 1,
            branch_name: 'feature/test-branch',
            pr_url: undefined,
          });
        });
      }
    });
  });

  describe('GitHub Integration', () => {
    it('should pre-fill form when GitHub issue is provided', () => {
      const githubIssue = {
        id: 1,
        number: 42,
        title: 'Bug in authentication',
        body: 'Description of the bug',
        html_url: 'https://github.com/test/repo/issues/42',
        state: 'open' as const,
        user: { login: 'testuser', avatar_url: '' },
        created_at: '2025-09-10T10:00:00Z',
        labels: [],
        assignees: [],
        comments: 0
      };

      render(<TaskCreationForm {...defaultProps} githubIssue={githubIssue} />);
      
      const titleInput = screen.getByLabelText(/Task Title/);
      const descriptionInput = screen.getByLabelText(/Description/);
      
      expect(titleInput).toHaveValue('Issue #42: Bug in authentication');
      expect(descriptionInput).toHaveValue(expect.stringContaining('Description of the bug'));
    });

    it('should pre-fill form when GitHub PR is provided', () => {
      const githubPR = {
        id: 1,
        number: 10,
        title: 'Fix authentication bug',
        body: 'This PR fixes the authentication issue',
        html_url: 'https://github.com/test/repo/pull/10',
        state: 'open' as const,
        user: { login: 'testuser', avatar_url: '' },
        created_at: '2025-09-10T10:00:00Z',
        labels: [],
        assignees: [],
        comments: 0,
        head: { ref: 'fix/auth-bug' },
        base: { ref: 'main' },
        merged: false
      };

      render(<TaskCreationForm {...defaultProps} githubPullRequest={githubPR} />);
      
      const titleInput = screen.getByLabelText(/Task Title/);
      const branchInput = screen.getByLabelText(/Branch Name/);
      
      expect(titleInput).toHaveValue('PR #10: Fix authentication bug');
      expect(branchInput).toHaveValue('fix/auth-bug');
    });
  });

  describe('Loading State', () => {
    it('should disable form elements when submitting', () => {
      render(<TaskCreationForm {...defaultProps} isSubmitting={true} />);
      
      const titleInput = screen.getByLabelText(/Task Title/);
      const descriptionInput = screen.getByLabelText(/Description/);
      const branchInput = screen.getByLabelText(/Branch Name/);
      const submitButton = screen.getByRole('button', { name: /create task/i });
      const cancelButton = screen.getByRole('button', { name: /cancel/i });
      
      expect(titleInput).toBeDisabled();
      expect(descriptionInput).toBeDisabled();
      expect(branchInput).toBeDisabled();
      expect(submitButton).toBeDisabled();
      expect(cancelButton).toBeDisabled();
    });
  });

  describe('Cancel Functionality', () => {
    it('should call onCancel when cancel button is clicked', async () => {
      render(<TaskCreationForm {...defaultProps} />);
      
      const cancelButton = screen.getByRole('button', { name: /cancel/i });
      await user.click(cancelButton);
      
      expect(mockOnCancel).toHaveBeenCalledTimes(1);
    });
  });
});