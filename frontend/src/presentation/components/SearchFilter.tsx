'use client';

import { useState } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import { Search, Filter, ChevronDown, X } from 'lucide-react';
import { Repository } from '../../domain/entities/Repository';

interface SearchFilterProps {
  searchQuery: string;
  onSearchChange: (query: string) => void;
  languageFilter: string;
  onLanguageChange: (language: string) => void;
  connectionFilter: 'all' | 'connected' | 'disconnected';
  onConnectionFilterChange: (filter: 'all' | 'connected' | 'disconnected') => void;
  availableLanguages: string[];
  isLoading?: boolean;
}

export function SearchFilter({ 
  searchQuery, 
  onSearchChange, 
  languageFilter, 
  onLanguageChange, 
  connectionFilter, 
  onConnectionFilterChange, 
  availableLanguages, 
  isLoading = false 
}: SearchFilterProps) {
  const [isFilterOpen, setIsFilterOpen] = useState(false);

  const handleLanguageChange = (value: string) => {
    onLanguageChange(value === 'all' ? '' : value);
  };

  const handleConnectedChange = (value: 'all' | 'connected' | 'disconnected') => {
    onConnectionFilterChange(value);
  };

  const clearAllFilters = () => {
    onSearchChange('');
    onLanguageChange('');
    onConnectionFilterChange('all');
    setIsFilterOpen(false);
  };

  const hasActiveFilters = 
    searchQuery !== '' || 
    languageFilter !== '' || 
    connectionFilter !== 'all';

  const activeFilterCount = [
    searchQuery !== '',
    languageFilter !== '',
    connectionFilter !== 'all',
  ].filter(Boolean).length;

  return (
    <div className="space-y-4">
      {/* Search Bar */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
        <Input
          placeholder="Search repositories by name, description, or owner..."
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
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
                      value={languageFilter || 'all'}
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


                  {/* Connection Status Filter */}
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Status</label>
                    <Select
                      value={connectionFilter}
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
        {searchQuery && (
          <Badge variant="secondary" className="gap-1">
            Search: {searchQuery}
            <button
              onClick={() => onSearchChange('')}
              className="hover:bg-muted-foreground/20 rounded-full p-1 min-h-[44px] min-w-[44px] flex items-center justify-center touch-manipulation"
              aria-label="Clear search"
            >
              <X className="h-4 w-4" />
            </button>
          </Badge>
        )}

        {languageFilter && (
          <Badge variant="secondary" className="gap-1">
            {languageFilter}
            <button
              onClick={() => handleLanguageChange('all')}
              className="hover:bg-muted-foreground/20 rounded-full p-1 min-h-[44px] min-w-[44px] flex items-center justify-center touch-manipulation"
              aria-label="Clear language filter"
            >
              <X className="h-4 w-4" />
            </button>
          </Badge>
        )}

        {connectionFilter !== 'all' && (
          <Badge variant="secondary" className="gap-1 capitalize">
            {connectionFilter}
            <button
              onClick={() => handleConnectedChange('all')}
              className="hover:bg-muted-foreground/20 rounded-full p-1 min-h-[44px] min-w-[44px] flex items-center justify-center touch-manipulation"
              aria-label="Clear connection status filter"
            >
              <X className="h-4 w-4" />
            </button>
          </Badge>
        )}
      </div>
    </div>
  );
}