import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/lib/components';
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js';
import { Doughnut } from 'react-chartjs-2';

ChartJS.register(ArcElement, Tooltip, Legend);

interface VMStatusChartProps {
  running: number;
  stopped: number;
  paused: number;
  other: number;
}

export const VMStatusChart: React.FC<VMStatusChartProps> = ({
  running,
  stopped,
  paused,
  other,
}) => {
  const data = {
    labels: ['Running', 'Stopped', 'Paused', 'Other'],
    datasets: [
      {
        data: [running, stopped, paused, other],
        backgroundColor: [
          'rgba(34, 197, 94, 0.6)',  // green-500 with transparency
          'rgba(239, 68, 68, 0.6)',   // red-500 with transparency
          'rgba(234, 179, 8, 0.6)',   // yellow-500 with transparency
          'rgba(148, 163, 184, 0.6)',  // slate-400 with transparency
        ],
        borderColor: [
          'rgb(34, 197, 94)',      // green-500
          'rgb(239, 68, 68)',      // red-500
          'rgb(234, 179, 8)',      // yellow-500
          'rgb(148, 163, 184)',    // slate-400
        ],
        borderWidth: 1,
      },
    ],
  };

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'bottom' as const,
        labels: {
          usePointStyle: true,
          padding: 20,
        },
      },
      tooltip: {
        callbacks: {
          label: function(context: any) {
            const label = context.label || '';
            const value = context.raw || 0;
            const total = context.dataset.data.reduce((a: number, b: number) => a + b, 0);
            const percentage = Math.round((value / total) * 100);
            return `${label}: ${value} (${percentage}%)`;
          }
        }
      }
    },
    cutout: '70%',
  };

  // Calculate total
  const total = running + stopped + paused + other;

  return (
    <Card className="col-span-1">
      <CardHeader>
        <CardTitle>VM Status</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-64 relative">
          <Doughnut data={data} options={options} />
          <div className="absolute inset-0 flex items-center justify-center flex-col">
            <span className="text-3xl font-bold">{total}</span>
            <span className="text-sm text-muted-foreground">Total VMs</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};