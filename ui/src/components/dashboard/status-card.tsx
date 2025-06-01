import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/lib/components';
import { cn } from '@/lib/utils';

interface StatusCardProps {
  title: string;
  value: string | number;
  icon?: React.ReactNode;
  description?: string;
  trend?: 'up' | 'down' | 'neutral';
  trendValue?: string;
  variant?: 'default' | 'success' | 'warning' | 'danger';
  className?: string;
}

export const StatusCard: React.FC<StatusCardProps> = ({
  title,
  value,
  icon,
  description,
  trend,
  trendValue,
  variant = 'default',
  className,
}) => {
  return (
    <Card 
      className={cn(
        "overflow-hidden",
        variant === 'success' && "border-green-500 dark:border-green-700",
        variant === 'warning' && "border-yellow-500 dark:border-yellow-700",
        variant === 'danger' && "border-red-500 dark:border-red-700",
        className
      )}
    >
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        {icon && (
          <div className={cn(
            "p-1.5 bg-muted rounded-lg",
            variant === 'success' && "text-green-500 dark:text-green-400",
            variant === 'warning' && "text-yellow-500 dark:text-yellow-400",
            variant === 'danger' && "text-red-500 dark:text-red-400"
          )}>
            {icon}
          </div>
        )}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        {(description || trend) && (
          <p className="text-xs text-muted-foreground mt-1 flex items-center">
            {description}
            {trend && (
              <span className={cn(
                "ml-1 inline-flex items-center",
                trend === 'up' && "text-green-500 dark:text-green-400",
                trend === 'down' && "text-red-500 dark:text-red-400"
              )}>
                {trend === 'up' && "↑"}
                {trend === 'down' && "↓"}
                {trendValue}
              </span>
            )}
          </p>
        )}
      </CardContent>
    </Card>
  );
};