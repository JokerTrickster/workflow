'use client';

import { useState, useMemo } from 'react';
import { Input } from '../../../components/ui/input';
import { Button } from '../../../components/ui/button';
import { Badge } from '../../../components/ui/badge';
import { Card, CardContent } from '../../../components/ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../../../components/ui/select';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '../../../components/ui/collapsible';
import { Search, Filter, ChevronDown, X } from 'lucide-react';
import { Repository } from '../../domain/entities/Repository';

interface SearchFilters {
  query: string;
  language: string;
  visibility: 'all' | 'public' | 'private';
  connected: 'all' | 'connected' | 'disconnected';
}

interface SearchFilterProps {
  filters: SearchFilters;
  onFiltersChange: (filters: SearchFilters) => void;
  repositories: Repository[];
}

export function SearchFilter({ filters, onFiltersChange, repositories }: SearchFilterProps) {
  const [isFilterOpen, setIsFilterOpen] = useState(false);

  // Extract unique languages from repositories
  const availableLanguages = useMemo(() => {
    const languages = repositories
      .map(repo => repo.language)
      .filter((lang): lang is string => lang !== null)
      .filter((lang, index, array) => array.indexOf(lang) === index)
      .sort();
    return languages;
  }, [repositories]);

  const handleQueryChange = (value: string) => {
    onFiltersChange({ ...filters, query: value });
  };

  const handleLanguageChange = (value: string) => {
    onFiltersChange({ 
      ...filters, 
      language: value === 'all' ? '' : value 
    });
  };

  const handleVisibilityChange = (value: 'all' | 'public' | 'private') => {
    onFiltersChange({ ...filters, visibility: value });
  };

  const handleConnectedChange = (value: 'all' | 'connected' | 'disconnected') => {
    onFiltersChange({ ...filters, connected: value });
  };

  const clearAllFilters = () => {
    onFiltersChange({
      query: '',
      language: '',
      visibility: 'all',
      connected: 'all',
    });
    setIsFilterOpen(false);
  };

  const hasActiveFilters = 
    filters.query !== '' || 
    filters.language !== '' || 
    filters.visibility !== 'all' || 
    filters.connected !== 'all';

  const activeFilterCount = [
    filters.query !== '',
    filters.language !== '',
    filters.visibility !== 'all',
    filters.connected !== 'all',
  ].filter(Boolean).length;

  return (
    <div className="space-y-4">
      {/* Search Bar */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
        <Input
          placeholder="Search repositories by name, description, or owner..."
          value={filters.query}
          onChange={(e) => handleQueryChange(e.target.value)}
          className="pl-10 pr-4"
        />
      </div>

      {/* Filter Controls */}
      <div className="flex items-center gap-2 flex-wrap">
        <Collapsible open={isFilterOpen} onOpenChange={setIsFilterOpen}>
          <CollapsibleTrigger asChild>
            <Button variant="outline" size="sm" className="gap-2">
              <Filter className="h-4 w-4" />
              Filters
              {activeFilterCount > 0 && (
                <Badge variant="secondary" className="ml-1 h-5 px-1.5 text-xs">
                  {activeFilterCount}
                </Badge>
              )}
              <ChevronDown className={`h-4 w-4 transition-transform ${isFilterOpen ? 'rotate-180' : ''}`} />
            </Button>
          </CollapsibleTrigger>
          
          <CollapsibleContent className="mt-4">
            <Card>
              <CardContent className="pt-6">
                <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                  {/* Language Filter */}
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Language</label>
                    <Select
                      value={filters.language || 'all'}
                      onValueChange={handleLanguageChange}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="All languages" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">All languages</SelectItem>
                        {availableLanguages.map(language => (
                          <SelectItem key={language} value={language}>
                            {language}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {/* Visibility Filter */}
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Visibility</label>
                    <Select
                      value={filters.visibility}
                      onValueChange={handleVisibilityChange}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">All repositories</SelectItem>
                        <SelectItem value="public">Public only</SelectItem>
                        <SelectItem value="private">Private only</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  {/* Connection Status Filter */}
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Status</label>
                    <Select
                      value={filters.connected}
                      onValueChange={handleConnectedChange}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">All statuses</SelectItem>
                        <SelectItem value="connected">Connected only</SelectItem>
                        <SelectItem value="disconnected">Disconnected only</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                {hasActiveFilters && (
                  <div className="mt-4 flex justify-end">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={clearAllFilters}
                      className="gap-2"
                    >
                      <X className="h-4 w-4" />
                      Clear all filters
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>
          </CollapsibleContent>
        </Collapsible>

        {/* Active Filter Badges */}
        {filters.query && (
          <Badge variant="secondary" className="gap-1">
            Search: {filters.query}
            <button
              onClick={() => handleQueryChange('')}
              className="hover:bg-muted-foreground/20 rounded-full p-0.5"
            >
              <X className="h-3 w-3" />
            </button>
          </Badge>
        )}

        {filters.language && (
          <Badge variant="secondary" className="gap-1">
            {filters.language}
            <button
              onClick={() => handleLanguageChange('all')}
              className="hover:bg-muted-foreground/20 rounded-full p-0.5"
            >
              <X className="h-3 w-3" />
            </button>
          </Badge>
        )}

        {filters.visibility !== 'all' && (
          <Badge variant="secondary" className="gap-1 capitalize">
            {filters.visibility}
            <button
              onClick={() => handleVisibilityChange('all')}
              className="hover:bg-muted-foreground/20 rounded-full p-0.5"
            >
              <X className="h-3 w-3" />
            </button>
          </Badge>
        )}

        {filters.connected !== 'all' && (
          <Badge variant="secondary" className="gap-1 capitalize">
            {filters.connected}
            <button
              onClick={() => handleConnectedChange('all')}
              className="hover:bg-muted-foreground/20 rounded-full p-0.5"
            >
              <X className="h-3 w-3" />
            </button>
          </Badge>
        )}
      </div>
    </div>
  );
}