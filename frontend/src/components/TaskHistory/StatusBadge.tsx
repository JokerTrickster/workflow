import React from 'react';

interface StatusBadgeProps {
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled';
  className?: string;
}

const statusConfig = {
  pending: {
    label: 'Pending',
    className: 'bg-yellow-100 text-yellow-800 border-yellow-200',
    icon: (
      <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
      </svg>
    )
  },
  processing: {
    label: 'Processing',
    className: 'bg-blue-100 text-blue-800 border-blue-200',
    icon: (
      <svg className="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24">
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"/>
        <path className="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
      </svg>
    )
  },
  completed: {
    label: 'Completed',
    className: 'bg-green-100 text-green-800 border-green-200',
    icon: (
      <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
      </svg>
    )
  },
  failed: {
    label: 'Failed',
    className: 'bg-red-100 text-red-800 border-red-200',
    icon: (
      <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
      </svg>
    )
  },
  cancelled: {
    label: 'Cancelled',
    className: 'bg-gray-100 text-gray-800 border-gray-200',
    icon: (
      <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
      </svg>
    )
  }
};

/**
 * Status badge component for displaying task status with appropriate colors and icons
 */
export function StatusBadge({ status, className = '' }: StatusBadgeProps) {
  const config = statusConfig[status];

  return (
    <span className={`
      inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium border
      ${config.className}
      ${className}
    `}>
      {config.icon}
      {config.label}
    </span>
  );
}