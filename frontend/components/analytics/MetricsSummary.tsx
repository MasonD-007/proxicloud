'use client';

import React, { useEffect, useState } from 'react';
import Card from '../ui/Card';

interface MetricsSummary {
  vmid: number;
  start_time: string;
  end_time: string;
  avg_cpu: number;
  max_cpu: number;
  avg_mem_usage: number;
  max_mem_usage: number;
  avg_disk_usage: number;
  total_net_in: number;
  total_net_out: number;
  data_points: number;
}

interface MetricsSummaryProps {
  vmid: number;
  hours?: number;
}

export function MetricsSummary({ vmid, hours = 24 }: MetricsSummaryProps) {
  const [summary, setSummary] = useState<MetricsSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchSummary = async () => {
      try {
        setLoading(true);
        const response = await fetch(
          `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'}/containers/${vmid}/metrics/summary?hours=${hours}`
        );
        
        if (!response.ok) {
          throw new Error('Failed to fetch metrics summary');
        }
        
        const data = await response.json();
        setSummary(data);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    fetchSummary();
    const interval = setInterval(fetchSummary, 60000); // Refresh every minute

    return () => clearInterval(interval);
  }, [vmid, hours]);

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
  };

  if (loading) {
    return (
      <Card className="p-6">
        <div className="animate-pulse space-y-4">
          <div className="h-4 bg-gray-700 rounded w-1/4"></div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="h-20 bg-gray-700 rounded"></div>
            ))}
          </div>
        </div>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="p-6">
        <p className="text-red-400">Error loading summary: {error}</p>
      </Card>
    );
  }

  if (!summary) {
    return (
      <Card className="p-6">
        <p className="text-gray-400">No metrics summary available yet.</p>
      </Card>
    );
  }

  return (
    <Card className="p-6">
      <h3 className="text-lg font-semibold mb-4 text-white">
        Metrics Summary (Last {hours}h)
      </h3>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-gray-800 rounded-lg p-4">
          <p className="text-xs text-gray-400 mb-1">Average CPU</p>
          <p className="text-2xl font-bold text-blue-400">{summary.avg_cpu.toFixed(1)}%</p>
          <p className="text-xs text-gray-500 mt-1">Max: {summary.max_cpu.toFixed(1)}%</p>
        </div>

        <div className="bg-gray-800 rounded-lg p-4">
          <p className="text-xs text-gray-400 mb-1">Average Memory</p>
          <p className="text-2xl font-bold text-green-400">{summary.avg_mem_usage.toFixed(1)}%</p>
          <p className="text-xs text-gray-500 mt-1">Max: {formatBytes(summary.max_mem_usage)}</p>
        </div>

        <div className="bg-gray-800 rounded-lg p-4">
          <p className="text-xs text-gray-400 mb-1">Average Disk</p>
          <p className="text-2xl font-bold text-purple-400">{summary.avg_disk_usage.toFixed(1)}%</p>
        </div>

        <div className="bg-gray-800 rounded-lg p-4">
          <p className="text-xs text-gray-400 mb-1">Network Transfer</p>
          <p className="text-sm font-semibold text-orange-400">
            ↓ {formatBytes(summary.total_net_in)}
          </p>
          <p className="text-sm font-semibold text-orange-400">
            ↑ {formatBytes(summary.total_net_out)}
          </p>
        </div>
      </div>

      <div className="mt-4 flex justify-between text-xs text-gray-500">
        <span>{summary.data_points} data points</span>
        <span>
          {new Date(summary.start_time).toLocaleString()} - {new Date(summary.end_time).toLocaleString()}
        </span>
      </div>
    </Card>
  );
}
