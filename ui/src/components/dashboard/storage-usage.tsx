import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/lib/components';
import { StoragePoolInfo } from '@/types/api';
import { formatBytes } from '@/lib/utils';

interface StorageUsageProps {
  pools: StoragePoolInfo[];
}

export const StorageUsage: React.FC<StorageUsageProps> = ({ pools }) => {
  return (
    <Card className="col-span-full md:col-span-2">
      <CardHeader>
        <CardTitle>Storage Pools</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {pools.map((pool) => {
            const usagePercentage = (pool.allocation / pool.capacity) * 100;
            const availablePercentage = (pool.available / pool.capacity) * 100;
            
            return (
              <div key={pool.name} className="space-y-2">
                <div className="flex justify-between items-center">
                  <div className="font-medium">{pool.name}</div>
                  <div className="text-sm text-muted-foreground">
                    {formatBytes(pool.allocation)} / {formatBytes(pool.capacity)}
                  </div>
                </div>
                
                <div className="h-2 bg-muted rounded-full overflow-hidden">
                  <div 
                    className={`h-full ${usagePercentage > 80 ? 'bg-red-500' : usagePercentage > 60 ? 'bg-yellow-500' : 'bg-green-500'}`}
                    style={{ width: `${usagePercentage}%` }}
                  />
                </div>
                
                <div className="flex justify-between text-xs text-muted-foreground">
                  <div>{usagePercentage.toFixed(1)}% used</div>
                  <div>{availablePercentage.toFixed(1)}% available</div>
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
};