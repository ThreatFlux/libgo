import React, { useEffect, useState } from 'react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js';
import { Line } from 'react-chartjs-2';
import { formatBytes } from '@/lib/utils';
import { VMMetrics } from '@/api/websocket';
import { Card, CardContent, CardHeader, CardTitle } from '@/lib/components';
import { LuCpu, LuHardDrive, LuMemoryStick, LuNetwork } from 'react-icons/lu';

// Register ChartJS components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

// Number of data points to keep in history
const MAX_DATA_POINTS = 20;

// Chart options
const getChartOptions = (title: string, yAxisLabel: string, min: number = 0, max: number = 100) => ({
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    x: {
      display: true,
      grid: {
        display: false,
      },
      ticks: {
        display: false,
      },
    },
    y: {
      display: true,
      min,
      max,
      title: {
        display: true,
        text: yAxisLabel,
      },
    },
  },
  plugins: {
    legend: {
      display: false,
    },
    title: {
      display: true,
      text: title,
    },
    tooltip: {
      enabled: true,
    },
  },
  animation: {
    duration: 200,
  },
});

// Chart colors
const chartColors = {
  cpu: {
    backgroundColor: 'rgba(66, 153, 225, 0.1)',
    borderColor: 'rgba(66, 153, 225, 1)',
  },
  memory: {
    backgroundColor: 'rgba(72, 187, 120, 0.1)',
    borderColor: 'rgba(72, 187, 120, 1)',
  },
  network: {
    rx: {
      backgroundColor: 'rgba(246, 173, 85, 0.1)',
      borderColor: 'rgba(246, 173, 85, 1)',
    },
    tx: {
      backgroundColor: 'rgba(237, 100, 166, 0.1)',
      borderColor: 'rgba(237, 100, 166, 1)',
    },
  },
  disk: {
    read: {
      backgroundColor: 'rgba(159, 122, 234, 0.1)',
      borderColor: 'rgba(159, 122, 234, 1)',
    },
    write: {
      backgroundColor: 'rgba(237, 137, 54, 0.1)',
      borderColor: 'rgba(237, 137, 54, 1)',
    },
  },
};

// Helper for generating empty chart data
const generateEmptyChartData = (label: string, color: { backgroundColor: string; borderColor: string }) => {
  const labels = Array(MAX_DATA_POINTS).fill('');
  const data = Array(MAX_DATA_POINTS).fill(null);
  
  return {
    labels,
    datasets: [
      {
        label,
        data,
        borderColor: color.borderColor,
        backgroundColor: color.backgroundColor,
        fill: true,
        tension: 0.4,
        pointRadius: 0,
      },
    ],
  };
};

// Helper to update chart data
const updateChartData = (
  prevData: any,
  newValue: number | null,
  label?: string,
) => {
  const updatedLabels = [...prevData.labels.slice(1), new Date().toLocaleTimeString()];
  
  const updatedData = {
    ...prevData,
    labels: updatedLabels,
    datasets: prevData.datasets.map((dataset: any) => ({
      ...dataset,
      label: label || dataset.label,
      data: [...dataset.data.slice(1), newValue],
    })),
  };
  
  return updatedData;
};

interface VMMetricsChartProps {
  vmName: string;
  metrics: VMMetrics | null;
  className?: string;
}

export const VMMetricsChart: React.FC<VMMetricsChartProps> = ({
  vmName,
  metrics,
  className = '',
}) => {
  // Chart data state
  const [cpuData, setCpuData] = useState(
    generateEmptyChartData('CPU Usage (%)', chartColors.cpu)
  );
  
  const [memoryData, setMemoryData] = useState(
    generateEmptyChartData('Memory Usage', chartColors.memory)
  );
  
  const [networkData, setNetworkData] = useState({
    labels: Array(MAX_DATA_POINTS).fill(''),
    datasets: [
      {
        label: 'Received',
        data: Array(MAX_DATA_POINTS).fill(null),
        borderColor: chartColors.network.rx.borderColor,
        backgroundColor: chartColors.network.rx.backgroundColor,
        fill: true,
        tension: 0.4,
        pointRadius: 0,
      },
      {
        label: 'Transmitted',
        data: Array(MAX_DATA_POINTS).fill(null),
        borderColor: chartColors.network.tx.borderColor,
        backgroundColor: chartColors.network.tx.backgroundColor,
        fill: true,
        tension: 0.4,
        pointRadius: 0,
      },
    ],
  });
  
  const [diskData, setDiskData] = useState({
    labels: Array(MAX_DATA_POINTS).fill(''),
    datasets: [
      {
        label: 'Read',
        data: Array(MAX_DATA_POINTS).fill(null),
        borderColor: chartColors.disk.read.borderColor,
        backgroundColor: chartColors.disk.read.backgroundColor,
        fill: true,
        tension: 0.4,
        pointRadius: 0,
      },
      {
        label: 'Write',
        data: Array(MAX_DATA_POINTS).fill(null),
        borderColor: chartColors.disk.write.borderColor,
        backgroundColor: chartColors.disk.write.backgroundColor,
        fill: true,
        tension: 0.4,
        pointRadius: 0,
      },
    ],
  });
  
  // Track previous metrics to calculate rates
  const [prevMetrics, setPrevMetrics] = useState<VMMetrics | null>(null);
  const [prevTimestamp, setPrevTimestamp] = useState<number | null>(null);
  
  // State for current metrics summary display
  const [memoryUsage, setMemoryUsage] = useState<string>('N/A');
  const [memoryPercent, setMemoryPercent] = useState<number>(0);
  const [networkRates, setNetworkRates] = useState({ rx: 'N/A', tx: 'N/A' });
  const [diskRates, setDiskRates] = useState({ read: 'N/A', write: 'N/A' });
  
  // Update charts when metrics change
  useEffect(() => {
    if (!metrics) return;
    
    const now = Date.now();
    const elapsedMs = prevTimestamp ? now - prevTimestamp : 1000; // Default to 1 second if no previous timestamp
    const elapsedSec = elapsedMs / 1000;
    
    // Update CPU chart
    setCpuData(prevData => updateChartData(prevData, metrics.cpu.utilization));
    
    // Update Memory chart
    const memPercent = metrics.memory.total ? (metrics.memory.used / metrics.memory.total) * 100 : 0;
    setMemoryData(prevData => updateChartData(prevData, memPercent));
    setMemoryUsage(`${formatBytes(metrics.memory.used)} / ${formatBytes(metrics.memory.total)}`);
    setMemoryPercent(memPercent);
    
    // Calculate network rates
    if (prevMetrics && elapsedSec > 0) {
      // Calculate network transfer rates
      const rxRate = prevMetrics.network.rxBytes 
        ? (metrics.network.rxBytes - prevMetrics.network.rxBytes) / elapsedSec 
        : 0;
      const txRate = prevMetrics.network.txBytes 
        ? (metrics.network.txBytes - prevMetrics.network.txBytes) / elapsedSec
        : 0;
      
      // Update network chart
      setNetworkData(prevData => ({
        ...prevData,
        labels: [...prevData.labels.slice(1), new Date().toLocaleTimeString()],
        datasets: [
          {
            ...prevData.datasets[0],
            data: [...prevData.datasets[0].data.slice(1), rxRate],
          },
          {
            ...prevData.datasets[1],
            data: [...prevData.datasets[1].data.slice(1), txRate],
          },
        ],
      }));
      
      setNetworkRates({
        rx: formatBytes(rxRate) + '/s',
        tx: formatBytes(txRate) + '/s',
      });
      
      // Calculate disk rates
      const readRate = prevMetrics.disk.readBytes 
        ? (metrics.disk.readBytes - prevMetrics.disk.readBytes) / elapsedSec
        : 0;
      const writeRate = prevMetrics.disk.writeBytes 
        ? (metrics.disk.writeBytes - prevMetrics.disk.writeBytes) / elapsedSec
        : 0;
      
      // Update disk chart
      setDiskData(prevData => ({
        ...prevData,
        labels: [...prevData.labels.slice(1), new Date().toLocaleTimeString()],
        datasets: [
          {
            ...prevData.datasets[0],
            data: [...prevData.datasets[0].data.slice(1), readRate],
          },
          {
            ...prevData.datasets[1],
            data: [...prevData.datasets[1].data.slice(1), writeRate],
          },
        ],
      }));
      
      setDiskRates({
        read: formatBytes(readRate) + '/s',
        write: formatBytes(writeRate) + '/s',
      });
    }
    
    // Store current metrics for next calculation
    setPrevMetrics(metrics);
    setPrevTimestamp(now);
  }, [metrics]);
  
  return (
    <div className={className}>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* CPU Usage */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="flex items-center text-lg">
              <LuCpu className="mr-2 h-4 w-4" />
              CPU Usage
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between mb-1">
              <div className="text-2xl font-bold">
                {metrics?.cpu.utilization?.toFixed(1) ?? 'N/A'}%
              </div>
            </div>
            <div className="h-[150px]">
              <Line
                data={cpuData}
                options={getChartOptions('CPU Usage', 'Percent (%)')}
              />
            </div>
          </CardContent>
        </Card>
        
        {/* Memory Usage */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="flex items-center text-lg">
              <LuMemoryStick className="mr-2 h-4 w-4" />
              Memory Usage
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between mb-1">
              <div className="text-sm">{memoryUsage}</div>
              <div className="font-bold">
                {memoryPercent.toFixed(1)}%
              </div>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2.5 mb-2">
              <div 
                className="bg-blue-600 h-2.5 rounded-full"
                style={{ width: `${memoryPercent}%` }}
              ></div>
            </div>
            <div className="h-[150px]">
              <Line
                data={memoryData}
                options={getChartOptions('Memory Usage', 'Percent (%)')}
              />
            </div>
          </CardContent>
        </Card>
        
        {/* Network Usage */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="flex items-center text-lg">
              <LuNetwork className="mr-2 h-4 w-4" />
              Network Activity
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between mb-2 text-sm">
              <div className="flex items-center">
                <div className="w-3 h-3 rounded-full mr-1" style={{ backgroundColor: chartColors.network.rx.borderColor }}></div>
                <span>RX: {networkRates.rx}</span>
              </div>
              <div className="flex items-center">
                <div className="w-3 h-3 rounded-full mr-1" style={{ backgroundColor: chartColors.network.tx.borderColor }}></div>
                <span>TX: {networkRates.tx}</span>
              </div>
            </div>
            <div className="h-[150px]">
              <Line
                data={networkData}
                options={getChartOptions('Network I/O', 'Bytes/s', undefined, undefined)}
              />
            </div>
          </CardContent>
        </Card>
        
        {/* Disk Usage */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="flex items-center text-lg">
              <LuHardDrive className="mr-2 h-4 w-4" />
              Disk Activity
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between mb-2 text-sm">
              <div className="flex items-center">
                <div className="w-3 h-3 rounded-full mr-1" style={{ backgroundColor: chartColors.disk.read.borderColor }}></div>
                <span>Read: {diskRates.read}</span>
              </div>
              <div className="flex items-center">
                <div className="w-3 h-3 rounded-full mr-1" style={{ backgroundColor: chartColors.disk.write.borderColor }}></div>
                <span>Write: {diskRates.write}</span>
              </div>
            </div>
            <div className="h-[150px]">
              <Line
                data={diskData}
                options={getChartOptions('Disk I/O', 'Bytes/s', undefined, undefined)}
              />
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};