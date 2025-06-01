import React from 'react';
import { cn, getStatusColor } from '@/lib/utils';
import { VMStatus } from '@/types/api';

interface VMStatusBadgeProps {
  status: VMStatus;
  className?: string;
  children?: React.ReactNode;
}

export const VMStatusBadge: React.FC<VMStatusBadgeProps> = ({ status, className, children }) => {
  let statusText: string;
  
  switch (status) {
    case 'running':
      statusText = 'Running';
      break;
    case 'stopped':
      statusText = 'Stopped';
      break;
    case 'paused':
      statusText = 'Paused';
      break;
    case 'shutdown':
      statusText = 'Shutdown';
      break;
    case 'crashed':
      statusText = 'Crashed';
      break;
    default:
      statusText = 'Unknown';
  }
  
  return (
    <span
      className={cn(
        "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium",
        getStatusColor(status),
        "text-white",
        className
      )}
    >
      <span className="flex-shrink-0 h-1.5 w-1.5 rounded-full bg-white mr-1.5" />
      {children || statusText}
    </span>
  );
};