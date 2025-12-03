'use client';

import React, { useEffect, useState } from 'react';
import Card from '../ui/Card';

interface Metric {
  vmid: number;
  timestamp: string;
  cpu_usage: number;
  mem_usage: number;
  mem_total: number;
  disk_usage: number;
  disk_total: number;
  net_in: number;
  net_out: number;
  uptime: number;
  status: string;
}

interface MetricsChartProps {
  vmid: number;
  hours?: number;
}

export function MetricsChart({ vmid, hours = 24 }: MetricsChartProps) {
  const [metrics, setMetrics] = useState<Metric[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        setLoading(true);
        const response = await fetch(
          `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'}/containers/${vmid}/metrics?hours=${hours}&limit=100`
        );
        
        if (!response.ok) {
          throw new Error('Failed to fetch metrics');
        }
        
        const data = await response.json();
        setMetrics(data || []);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    fetchMetrics();
    const interval = setInterval(fetchMetrics, 60000); // Refresh every minute

    return () => clearInterval(interval);
  }, [vmid, hours]);

  if (loading) {
    return (
      <Card className="p-6">
        <div className="animate-pulse space-y-4">
          <div className="h-4 bg-gray-700 rounded w-1/4"></div>
          <div className="h-48 bg-gray-700 rounded"></div>
        </div>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="p-6">
        <p className="text-red-400">Error loading metrics: {error}</p>
      </Card>
    );
  }

  if (!metrics || metrics.length === 0) {
    return (
      <Card className="p-6">
        <p className="text-gray-400">No metrics data available yet. Metrics will appear after the collector runs.</p>
      </Card>
    );
  }

  // Calculate min/max for scaling
  const cpuValues = metrics.map(m => m.cpu_usage);
  const maxCPU = Math.max(...cpuValues, 100);

  const memValues = metrics.map(m => (m.mem_usage / m.mem_total) * 100);
  const maxMem = Math.max(...memValues, 100);

  // Reverse to show oldest to newest (left to right)
  const reversedMetrics = [...metrics].reverse();

  return (
    <div className="space-y-6">
      {/* CPU Chart */}
      <Card className="p-6">
        <h3 className="text-lg font-semibold mb-4 text-white">CPU Usage</h3>
        <div className="relative h-48 flex items-end space-x-1">
          {reversedMetrics.map((metric, index) => {
            const height = (metric.cpu_usage / maxCPU) * 100;
            return (
              <div
                key={index}
                className="flex-1 bg-blue-500 rounded-t hover:bg-blue-400 transition-colors relative group"
                style={{ height: `${height}%`, minHeight: '2px' }}
                title={`${new Date(metric.timestamp).toLocaleString()}: ${metric.cpu_usage.toFixed(1)}%`}
              >
                <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-gray-900 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-10">
                  {metric.cpu_usage.toFixed(1)}%
                </div>
              </div>
            );
          })}
        </div>
        <div className="flex justify-between mt-2 text-xs text-gray-400">
          <span>-{hours}h</span>
          <span>Now</span>
        </div>
      </Card>

      {/* Memory Chart */}
      <Card className="p-6">
        <h3 className="text-lg font-semibold mb-4 text-white">Memory Usage</h3>
        <div className="relative h-48 flex items-end space-x-1">
          {reversedMetrics.map((metric, index) => {
            const percentage = (metric.mem_usage / metric.mem_total) * 100;
            const height = (percentage / maxMem) * 100;
            return (
              <div
                key={index}
                className="flex-1 bg-green-500 rounded-t hover:bg-green-400 transition-colors relative group"
                style={{ height: `${height}%`, minHeight: '2px' }}
                title={`${new Date(metric.timestamp).toLocaleString()}: ${percentage.toFixed(1)}%`}
              >
                <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-gray-900 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-10">
                  {percentage.toFixed(1)}%
                </div>
              </div>
            );
          })}
        </div>
        <div className="flex justify-between mt-2 text-xs text-gray-400">
          <span>-{hours}h</span>
          <span>Now</span>
        </div>
      </Card>

      {/* Disk Chart */}
      <Card className="p-6">
        <h3 className="text-lg font-semibold mb-4 text-white">Disk Usage</h3>
        <div className="relative h-48 flex items-end space-x-1">
          {reversedMetrics.map((metric, index) => {
            const percentage = (metric.disk_usage / metric.disk_total) * 100;
            const height = percentage;
            return (
              <div
                key={index}
                className="flex-1 bg-purple-500 rounded-t hover:bg-purple-400 transition-colors relative group"
                style={{ height: `${height}%`, minHeight: '2px' }}
                title={`${new Date(metric.timestamp).toLocaleString()}: ${percentage.toFixed(1)}%`}
              >
                <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-gray-900 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-10">
                  {percentage.toFixed(1)}%
                </div>
              </div>
            );
          })}
        </div>
        <div className="flex justify-between mt-2 text-xs text-gray-400">
          <span>-{hours}h</span>
          <span>Now</span>
        </div>
      </Card>
    </div>
  );
}
