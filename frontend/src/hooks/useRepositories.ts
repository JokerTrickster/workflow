'use client'

import { useState, useEffect, useCallback } from 'react'
import { useAuth } from '@/contexts/AuthContext'
import { Repository, RepositoriesQuery, RepositoriesResponse } from '@/types/github'

interface UseRepositoriesState {
  repositories: Repository[]
  loading: boolean
  error: string | null
  totalCount: number
  hasNextPage: boolean
  rateLimit: {
    remaining: number
    resetAt: string
  } | null
}

interface UseRepositoriesActions {
  fetchRepositories: (query?: RepositoriesQuery) => Promise<void>
  fetchAllRepositories: (maxPages?: number) => Promise<void>
  searchRepositories: (searchQuery: string, options?: any) => Promise<void>
  refresh: () => Promise<void>
  reset: () => void
}

export function useRepositories(initialQuery: RepositoriesQuery = {}) {
  const { isAuthenticated } = useAuth()
  
  const [state, setState] = useState<UseRepositoriesState>({
    repositories: [],
    loading: false,
    error: null,
    totalCount: 0,
    hasNextPage: false,
    rateLimit: null,
  })

  const [currentQuery, setCurrentQuery] = useState<RepositoriesQuery>(initialQuery)

  const fetchRepositories = useCallback(async (query: RepositoriesQuery = currentQuery) => {
    if (!isAuthenticated) {
      setState(prev => ({ ...prev, error: 'Authentication required' }))
      return
    }

    setState(prev => ({ ...prev, loading: true, error: null }))

    try {
      const searchParams = new URLSearchParams()
      
      if (query.page) searchParams.set('page', query.page.toString())
      if (query.per_page) searchParams.set('per_page', query.per_page.toString())
      if (query.sort) searchParams.set('sort', query.sort)
      if (query.direction) searchParams.set('direction', query.direction)
      if (query.type) searchParams.set('type', query.type)

      const response = await fetch(`/api/github/repositories?${searchParams.toString()}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'Unknown error' }))
        throw new Error(errorData.error || `HTTP ${response.status}`)
      }

      const data: RepositoriesResponse = await response.json()

      setState(prev => ({
        ...prev,
        repositories: data.repositories,
        totalCount: data.totalCount,
        hasNextPage: data.hasNextPage,
        rateLimit: data.rateLimit,
        loading: false,
        error: null,
      }))

      setCurrentQuery(query)

    } catch (error) {
      console.error('Error fetching repositories:', error)
      setState(prev => ({
        ...prev,
        loading: false,
        error: error instanceof Error ? error.message : 'Failed to fetch repositories',
      }))
    }
  }, [isAuthenticated, currentQuery])

  const fetchAllRepositories = useCallback(async (maxPages: number = 10) => {
    if (!isAuthenticated) {
      setState(prev => ({ ...prev, error: 'Authentication required' }))
      return
    }

    setState(prev => ({ ...prev, loading: true, error: null }))

    try {
      const response = await fetch('/api/github/repositories', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          action: 'fetchAll',
          maxPages,
          sort: currentQuery.sort || 'updated',
          direction: currentQuery.direction || 'desc',
          type: currentQuery.type || 'all',
        }),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'Unknown error' }))
        throw new Error(errorData.error || `HTTP ${response.status}`)
      }

      const data = await response.json()

      setState(prev => ({
        ...prev,
        repositories: data.repositories,
        totalCount: data.totalCount,
        hasNextPage: false, // All fetched
        loading: false,
        error: null,
      }))

    } catch (error) {
      console.error('Error fetching all repositories:', error)
      setState(prev => ({
        ...prev,
        loading: false,
        error: error instanceof Error ? error.message : 'Failed to fetch all repositories',
      }))
    }
  }, [isAuthenticated, currentQuery])

  const searchRepositories = useCallback(async (
    searchQuery: string, 
    options: any = {}
  ) => {
    if (!isAuthenticated) {
      setState(prev => ({ ...prev, error: 'Authentication required' }))
      return
    }

    setState(prev => ({ ...prev, loading: true, error: null }))

    try {
      const response = await fetch('/api/github/repositories', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          action: 'search',
          query: searchQuery,
          ...options,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'Unknown error' }))
        throw new Error(errorData.error || `HTTP ${response.status}`)
      }

      const data = await response.json()

      setState(prev => ({
        ...prev,
        repositories: data.repositories,
        totalCount: data.totalCount,
        hasNextPage: false, // Search doesn't use pagination in the same way
        loading: false,
        error: null,
      }))

    } catch (error) {
      console.error('Error searching repositories:', error)
      setState(prev => ({
        ...prev,
        loading: false,
        error: error instanceof Error ? error.message : 'Failed to search repositories',
      }))
    }
  }, [isAuthenticated])

  const refresh = useCallback(async () => {
    await fetchRepositories(currentQuery)
  }, [fetchRepositories, currentQuery])

  const reset = useCallback(() => {
    setState({
      repositories: [],
      loading: false,
      error: null,
      totalCount: 0,
      hasNextPage: false,
      rateLimit: null,
    })
    setCurrentQuery(initialQuery)
  }, [initialQuery])

  // Auto-fetch on mount if authenticated
  useEffect(() => {
    if (isAuthenticated) {
      fetchRepositories(initialQuery)
    }
  }, [isAuthenticated]) // Only depend on isAuthenticated to avoid infinite loops

  return {
    ...state,
    fetchRepositories,
    fetchAllRepositories,
    searchRepositories,
    refresh,
    reset,
  } as UseRepositoriesState & UseRepositoriesActions
}