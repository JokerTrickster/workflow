import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { RepositoryCard } from '../../presentation/components/RepositoryCard'
import { Repository } from '../../domain/entities/Repository'

const mockRepository: Repository = {
  id: 1,
  name: 'test-repo',
  full_name: 'user/test-repo',
  description: 'A test repository for unit testing',
  html_url: 'https://github.com/user/test-repo',
  clone_url: 'https://github.com/user/test-repo.git',
  ssh_url: 'git@github.com:user/test-repo.git',
  private: false,
  language: 'TypeScript',
  stargazers_count: 42,
  forks_count: 7,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-09-01T12:30:00Z',
  pushed_at: '2024-09-01T12:30:00Z',
  default_branch: 'main',
  is_connected: false,
}

const mockConnectedRepository: Repository = {
  ...mockRepository,
  id: 2,
  name: 'connected-repo',
  full_name: 'user/connected-repo',
  is_connected: true,
}

const mockPrivateRepository: Repository = {
  ...mockRepository,
  id: 3,
  private: true,
}

const mockOnConnect = jest.fn()
const mockOnDisconnect = jest.fn()
const mockOnOpenWorkspace = jest.fn()

describe('RepositoryCard Component', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders repository information correctly', () => {
    render(
      <RepositoryCard
        repository={mockRepository}
        onConnect={mockOnConnect}
        onDisconnect={mockOnDisconnect}
        onOpenWorkspace={mockOnOpenWorkspace}
      />
    )

    expect(screen.getByText('test-repo')).toBeInTheDocument()
    expect(screen.getByText('user/test-repo')).toBeInTheDocument()
    expect(screen.getByText('A test repository for unit testing')).toBeInTheDocument()
    expect(screen.getByText('TypeScript')).toBeInTheDocument()
    expect(screen.getByText('42')).toBeInTheDocument()
    expect(screen.getByText('7')).toBeInTheDocument()
  })

  it('shows external link to GitHub repository', () => {
    render(
      <RepositoryCard
        repository={mockRepository}
        onConnect={mockOnConnect}
        onDisconnect={mockOnDisconnect}
        onOpenWorkspace={mockOnOpenWorkspace}
      />
    )

    const githubLink = screen.getByRole('link', { name: /user\/test-repo/i })
    expect(githubLink).toHaveAttribute('href', 'https://github.com/user/test-repo')
    expect(githubLink).toHaveAttribute('target', '_blank')
    expect(githubLink).toHaveAttribute('rel', 'noopener noreferrer')
  })

  it('shows "Private" badge for private repositories', () => {
    render(
      <RepositoryCard
        repository={mockPrivateRepository}
        onConnect={mockOnConnect}
        onDisconnect={mockOnDisconnect}
        onOpenWorkspace={mockOnOpenWorkspace}
      />
    )

    expect(screen.getByText('Private')).toBeInTheDocument()
  })

  it('shows "Connect Repository" button for disconnected repositories', () => {
    render(
      <RepositoryCard
        repository={mockRepository}
        onConnect={mockOnConnect}
        onDisconnect={mockOnDisconnect}
        onOpenWorkspace={mockOnOpenWorkspace}
      />
    )

    expect(screen.getByRole('button', { name: /connect repository/i })).toBeInTheDocument()
    expect(screen.queryByText('Connected')).not.toBeInTheDocument()
  })

  it('calls onConnect when connect button is clicked', async () => {
    const user = userEvent.setup()
    render(
      <RepositoryCard
        repository={mockRepository}
        onConnect={mockOnConnect}
        onDisconnect={mockOnDisconnect}
        onOpenWorkspace={mockOnOpenWorkspace}
      />
    )

    const connectButton = screen.getByRole('button', { name: /connect repository/i })
    await user.click(connectButton)

    expect(mockOnConnect).toHaveBeenCalledWith(mockRepository.id)
  })

  it('shows "Connected" badge and workspace buttons for connected repositories', () => {
    render(
      <RepositoryCard
        repository={mockConnectedRepository}
        onConnect={mockOnConnect}
        onDisconnect={mockOnDisconnect}
        onOpenWorkspace={mockOnOpenWorkspace}
      />
    )

    expect(screen.getByText('Connected')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /open workspace/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /disconnect/i })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /connect repository/i })).not.toBeInTheDocument()
  })

  it('calls onOpenWorkspace when workspace button is clicked', async () => {
    const user = userEvent.setup()
    render(
      <RepositoryCard
        repository={mockConnectedRepository}
        onConnect={mockOnConnect}
        onDisconnect={mockOnDisconnect}
        onOpenWorkspace={mockOnOpenWorkspace}
      />
    )

    const workspaceButton = screen.getByRole('button', { name: /open workspace/i })
    await user.click(workspaceButton)

    expect(mockOnOpenWorkspace).toHaveBeenCalledWith(mockConnectedRepository)
  })

  it('calls onDisconnect when disconnect button is clicked', async () => {
    const user = userEvent.setup()
    render(
      <RepositoryCard
        repository={mockConnectedRepository}
        onConnect={mockOnConnect}
        onDisconnect={mockOnDisconnect}
        onOpenWorkspace={mockOnOpenWorkspace}
      />
    )

    const disconnectButton = screen.getByRole('button', { name: /disconnect/i })
    await user.click(disconnectButton)

    expect(mockOnDisconnect).toHaveBeenCalledWith(mockConnectedRepository.id)
  })

  it('shows disconnect button with proper styling and icon', () => {
    render(
      <RepositoryCard
        repository={mockConnectedRepository}
        onConnect={mockOnConnect}
        onDisconnect={mockOnDisconnect}
        onOpenWorkspace={mockOnOpenWorkspace}
      />
    )

    const disconnectButton = screen.getByRole('button', { name: /disconnect/i })
    expect(disconnectButton.className).toContain('text-muted-foreground')
    expect(disconnectButton.className).toContain('hover:text-destructive')
    
    // Check that the Unplug icon is present
    const unplugIcon = disconnectButton.querySelector('svg')
    expect(unplugIcon).toBeInTheDocument()
  })

  it('displays formatted update date', () => {
    render(
      <RepositoryCard
        repository={mockRepository}
        onConnect={mockOnConnect}
        onDisconnect={mockOnDisconnect}
        onOpenWorkspace={mockOnOpenWorkspace}
      />
    )

    // Check that a date is displayed (format may vary by locale)
    const dateElement = screen.getByText(/2024/)
    expect(dateElement).toBeInTheDocument()
  })

  it('handles repositories without description gracefully', () => {
    const repoWithoutDescription = {
      ...mockRepository,
      description: null,
    }

    render(
      <RepositoryCard
        repository={repoWithoutDescription}
        onConnect={mockOnConnect}
        onDisconnect={mockOnDisconnect}
        onOpenWorkspace={mockOnOpenWorkspace}
      />
    )

    // Component should still render without errors
    expect(screen.getByText('test-repo')).toBeInTheDocument()
    expect(screen.queryByText('A test repository for unit testing')).not.toBeInTheDocument()
  })

  it('handles repositories without language gracefully', () => {
    const repoWithoutLanguage = {
      ...mockRepository,
      language: null,
    }

    render(
      <RepositoryCard
        repository={repoWithoutLanguage}
        onConnect={mockOnConnect}
        onDisconnect={mockOnDisconnect}
        onOpenWorkspace={mockOnOpenWorkspace}
      />
    )

    // Component should still render without errors
    expect(screen.getByText('test-repo')).toBeInTheDocument()
    expect(screen.queryByText('TypeScript')).not.toBeInTheDocument()
  })

  it('displays both action buttons with proper spacing for connected repositories', () => {
    render(
      <RepositoryCard
        repository={mockConnectedRepository}
        onConnect={mockOnConnect}
        onDisconnect={mockOnDisconnect}
        onOpenWorkspace={mockOnOpenWorkspace}
      />
    )

    const buttonContainer = screen.getByRole('button', { name: /open workspace/i }).parentElement
    expect(buttonContainer?.className).toContain('gap-2')
    expect(buttonContainer?.className).toContain('flex')
  })
})