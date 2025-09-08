import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SearchFilter } from '../../presentation/components/SearchFilter'
import { Repository } from '../../domain/entities/Repository'

const mockRepositories: Repository[] = [
  {
    id: 1,
    name: 'test-repo-1',
    full_name: 'user/test-repo-1',
    description: 'A test repository',
    html_url: 'https://github.com/user/test-repo-1',
    clone_url: 'https://github.com/user/test-repo-1.git',
    ssh_url: 'git@github.com:user/test-repo-1.git',
    private: false,
    language: 'TypeScript',
    stargazers_count: 10,
    forks_count: 2,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    pushed_at: '2024-01-01T00:00:00Z',
    default_branch: 'main',
    is_connected: true,
  },
  {
    id: 2,
    name: 'test-repo-2',
    full_name: 'user/test-repo-2',
    description: 'Another test repository',
    html_url: 'https://github.com/user/test-repo-2',
    clone_url: 'https://github.com/user/test-repo-2.git',
    ssh_url: 'git@github.com:user/test-repo-2.git',
    private: true,
    language: 'JavaScript',
    stargazers_count: 5,
    forks_count: 1,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    pushed_at: '2024-01-01T00:00:00Z',
    default_branch: 'main',
    is_connected: false,
  }
]

const mockFilters = {
  query: '',
  language: '',
  visibility: 'all' as const,
  connected: 'all' as const,
}

const mockOnFiltersChange = jest.fn()

describe('SearchFilter Component', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders search input with placeholder text', () => {
    render(
      <SearchFilter 
        filters={mockFilters} 
        onFiltersChange={mockOnFiltersChange} 
        repositories={mockRepositories} 
      />
    )
    
    expect(screen.getByPlaceholderText(/search repositories/i)).toBeInTheDocument()
  })

  it('calls onFiltersChange when search query is entered', async () => {
    const user = userEvent.setup()
    render(
      <SearchFilter 
        filters={mockFilters} 
        onFiltersChange={mockOnFiltersChange} 
        repositories={mockRepositories} 
      />
    )
    
    const searchInput = screen.getByPlaceholderText(/search repositories/i)
    await user.type(searchInput, 'test')
    
    await waitFor(() => {
      expect(mockOnFiltersChange).toHaveBeenCalledWith({
        ...mockFilters,
        query: 'test'
      })
    })
  })

  it('shows filters button with count when filters are active', () => {
    const activeFilters = {
      ...mockFilters,
      query: 'test',
      language: 'TypeScript'
    }
    
    render(
      <SearchFilter 
        filters={activeFilters} 
        onFiltersChange={mockOnFiltersChange} 
        repositories={mockRepositories} 
      />
    )
    
    expect(screen.getByText('2')).toBeInTheDocument() // Filter count badge
  })

  it('expands filter panel when filters button is clicked', async () => {
    const user = userEvent.setup()
    render(
      <SearchFilter 
        filters={mockFilters} 
        onFiltersChange={mockOnFiltersChange} 
        repositories={mockRepositories} 
      />
    )
    
    const filtersButton = screen.getByRole('button', { name: /filters/i })
    await user.click(filtersButton)
    
    expect(screen.getByText('Language')).toBeInTheDocument()
    expect(screen.getByText('Visibility')).toBeInTheDocument()
    expect(screen.getByText('Status')).toBeInTheDocument()
  })

  it('shows available languages from repositories', async () => {
    const user = userEvent.setup()
    render(
      <SearchFilter 
        filters={mockFilters} 
        onFiltersChange={mockOnFiltersChange} 
        repositories={mockRepositories} 
      />
    )
    
    // Expand filters
    const filtersButton = screen.getByRole('button', { name: /filters/i })
    await user.click(filtersButton)
    
    // Open language dropdown
    const languageSelect = screen.getByRole('combobox', { name: /language/i })
    await user.click(languageSelect)
    
    expect(screen.getByText('TypeScript')).toBeInTheDocument()
    expect(screen.getByText('JavaScript')).toBeInTheDocument()
  })

  it('displays active filter badges', () => {
    const activeFilters = {
      query: 'test',
      language: 'TypeScript',
      visibility: 'private' as const,
      connected: 'connected' as const,
    }
    
    render(
      <SearchFilter 
        filters={activeFilters} 
        onFiltersChange={mockOnFiltersChange} 
        repositories={mockRepositories} 
      />
    )
    
    expect(screen.getByText('Search: test')).toBeInTheDocument()
    expect(screen.getByText('TypeScript')).toBeInTheDocument()
    expect(screen.getByText('private')).toBeInTheDocument()
    expect(screen.getByText('connected')).toBeInTheDocument()
  })

  it('clears individual filter when badge X button is clicked', async () => {
    const user = userEvent.setup()
    const activeFilters = {
      ...mockFilters,
      query: 'test'
    }
    
    render(
      <SearchFilter 
        filters={activeFilters} 
        onFiltersChange={mockOnFiltersChange} 
        repositories={mockRepositories} 
      />
    )
    
    const clearButton = screen.getByRole('button', { name: /clear search/i })
    await user.click(clearButton)
    
    expect(mockOnFiltersChange).toHaveBeenCalledWith({
      ...activeFilters,
      query: ''
    })
  })

  it('clears all filters when clear all button is clicked', async () => {
    const user = userEvent.setup()
    const activeFilters = {
      query: 'test',
      language: 'TypeScript',
      visibility: 'private' as const,
      connected: 'connected' as const,
    }
    
    render(
      <SearchFilter 
        filters={activeFilters} 
        onFiltersChange={mockOnFiltersChange} 
        repositories={mockRepositories} 
      />
    )
    
    // Expand filters to see clear all button
    const filtersButton = screen.getByRole('button', { name: /filters/i })
    await user.click(filtersButton)
    
    const clearAllButton = screen.getByRole('button', { name: /clear all filters/i })
    await user.click(clearAllButton)
    
    expect(mockOnFiltersChange).toHaveBeenCalledWith({
      query: '',
      language: '',
      visibility: 'all',
      connected: 'all',
    })
  })

  it('has proper touch targets for mobile accessibility', () => {
    const activeFilters = {
      ...mockFilters,
      query: 'test'
    }
    
    render(
      <SearchFilter 
        filters={activeFilters} 
        onFiltersChange={mockOnFiltersChange} 
        repositories={mockRepositories} 
      />
    )
    
    const clearButton = screen.getByRole('button', { name: /clear search/i })
    const styles = window.getComputedStyle(clearButton)
    
    // Check that button has minimum touch target class applied
    expect(clearButton.className).toContain('min-h-[44px]')
    expect(clearButton.className).toContain('min-w-[44px]')
  })
})