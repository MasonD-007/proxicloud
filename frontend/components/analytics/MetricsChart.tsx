'use client';

import React, { useEffect, useState } from 'react';
import Card from '../ui/Card';
import { getMetrics } from '@/lib/api';
import type { MetricsData } from '@/lib/types';

interface MetricsChartProps {
  vmid: number;
  hours?: number;
}

export function MetricsChart({ vmid, hours = 24 }: MetricsChartProps) {
  const [metrics, setMetrics] = useState<MetricsData[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        setLoading(true);
        const data = await getMetrics(vmid, `${hours}h`);
        // Ensure data is always an array
        setMetrics(Array.isArray(data) ? data : []);
        setError(null);
      } catch (err) {
        console.error('Failed to fetch metrics:', err);
        setError(err instanceof Error ? err.message : 'Unknown error');
        setMetrics([]); // Set to empty array on error
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

  // Calculate min/max for scaling - use safe array access
  const cpuValues = metrics.map(m => m.cpu || 0);
  const maxCPU = cpuValues.length > 0 ? Math.max(...cpuValues, 100) : 100;

  const memValues = metrics.map(m => m.memory || 0);
  const maxMem = memValues.length > 0 ? Math.max(...memValues, 100) : 100;

  // Reverse to show oldest to newest (left to right)
  const reversedMetrics = [...metrics].reverse();

  return (
    <div className="space-y-6">
      {/* CPU Chart */}
      <Card className="p-6">
        <h3 className="text-lg font-semibold mb-4 text-white">CPU Usage</h3>
        <div className="relative h-48 flex items-end space-x-1">
          {reversedMetrics.map((metric, index) => {
            const cpuValue = metric.cpu || 0;
            const height = (cpuValue / maxCPU) * 100;
            return (
              <div
                key={index}
                className="flex-1 bg-blue-500 rounded-t hover:bg-blue-400 transition-colors relative group"
                style={{ height: `${height}%`, minHeight: '2px' }}
                title={`${new Date(metric.timestamp).toLocaleString()}: ${cpuValue.toFixed(1)}%`}
              >
                <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-gray-900 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-10">
                  {cpuValue.toFixed(1)}%
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
            const memValue = metric.memory || 0;
            const height = (memValue / maxMem) * 100;
            return (
              <div
                key={index}
                className="flex-1 bg-green-500 rounded-t hover:bg-green-400 transition-colors relative group"
                style={{ height: `${height}%`, minHeight: '2px' }}
                title={`${new Date(metric.timestamp).toLocaleString()}: ${memValue.toFixed(1)}%`}
              >
                <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-gray-900 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-10">
                  {memValue.toFixed(1)}%
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

      {/* Network Chart */}
      <Card className="p-6">
        <h3 className="text-lg font-semibold mb-4 text-white">Network Usage</h3>
        <div className="relative h-48 flex items-end space-x-1">
          {reversedMetrics.map((metric, index) => {
            const netIn = metric.net_in || 0;
            const netOut = metric.net_out || 0;
            const maxNet = Math.max(...metrics.map(m => Math.max(m.net_in || 0, m.net_out || 0)), 1);
            const heightIn = (netIn / maxNet) * 100;
            const heightOut = (netOut / maxNet) * 100;
            return (
              <div
                key={index}
                className="flex-1 flex flex-col justify-end gap-1 relative group"
              >
                <div
                  className="bg-purple-500 rounded-t hover:bg-purple-400 transition-colors"
                  style={{ height: `${heightIn}%`, minHeight: '2px' }}
                  title={`${new Date(metric.timestamp).toLocaleString()}: In ${(netIn / 1024 / 1024).toFixed(2)} MB/s`}
                />
                <div
                  className="bg-orange-500 rounded-t hover:bg-orange-400 transition-colors"
                  style={{ height: `${heightOut}%`, minHeight: '2px' }}
                  title={`${new Date(metric.timestamp).toLocaleString()}: Out ${(netOut / 1024 / 1024).toFixed(2)} MB/s`}
                />
              </div>
            );
          })}
        </div>
        <div className="flex justify-between mt-2 text-xs text-gray-400">
          <span>-{hours}h</span>
          <span>Now</span>
        </div>
        <div className="flex justify-center gap-4 mt-2 text-xs">
          <span className="flex items-center gap-1">
            <div className="w-3 h-3 bg-purple-500 rounded"></div>
            <span className="text-gray-400">In</span>
          </span>
          <span className="flex items-center gap-1">
            <div className="w-3 h-3 bg-orange-500 rounded"></div>
            <span className="text-gray-400">Out</span>
          </span>
        </div>
      </Card>
    </div>
  );
}
