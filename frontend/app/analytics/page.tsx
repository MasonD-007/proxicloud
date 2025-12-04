'use client';

import React, { useEffect, useState } from 'react';
import { getContainers } from '@/lib/api';
import { MetricsChart } from '@/components/analytics/MetricsChart';
import { MetricsSummary } from '@/components/analytics/MetricsSummary';
import Card from '@/components/ui/Card';
import Select from '@/components/ui/Select';
import type { Container } from '@/lib/types';

export default function AnalyticsPage() {
  const [containers, setContainers] = useState<Container[]>([]);
  const [selectedVMID, setSelectedVMID] = useState<number | null>(null);
  const [timeRange, setTimeRange] = useState<number>(24);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchContainers = async () => {
      try {
        setLoading(true);
        const data = await getContainers();
        setContainers(data);
        
        // Auto-select first running container
        const runningContainer = data.find(c => c.status === 'running');
        if (runningContainer) {
          setSelectedVMID(runningContainer.vmid);
        } else if (data.length > 0) {
          setSelectedVMID(data[0].vmid);
        }
        
        setError(null);
      } catch (err) {
        console.error('Failed to fetch containers:', err);
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    fetchContainers();
  }, []);

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-white">Analytics</h1>
        </div>
        <Card className="p-6">
          <div className="animate-pulse space-y-4">
            <div className="h-8 bg-gray-700 rounded w-1/4"></div>
            <div className="h-64 bg-gray-700 rounded"></div>
          </div>
        </Card>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-white">Analytics</h1>
        </div>
        <Card className="p-6">
          <p className="text-red-400">Error loading containers: {error}</p>
        </Card>
      </div>
    );
  }

  if (containers.length === 0) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-white">Analytics</h1>
        </div>
        <Card className="p-6">
          <p className="text-gray-400">
            No containers available. Create a container to view analytics.
          </p>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-white">Analytics</h1>
      </div>

      {/* Controls */}
      <Card className="p-6">
        <div className="flex flex-col md:flex-row gap-4">
          <div className="flex-1">
            <Select
              label="Container"
              value={selectedVMID?.toString() || ''}
              onChange={(e) => setSelectedVMID(Number(e.target.value))}
              options={[
                { value: '', label: 'Select a container' },
                ...containers.map((container) => ({
                  value: container.vmid.toString(),
                  label: `${container.name} (VMID: ${container.vmid}) - ${container.status}`,
                })),
              ]}
            />
          </div>

          <div className="w-full md:w-48">
            <Select
              label="Time Range"
              value={timeRange.toString()}
              onChange={(e) => setTimeRange(Number(e.target.value))}
              options={[
                { value: '1', label: 'Last Hour' },
                { value: '6', label: 'Last 6 Hours' },
                { value: '24', label: 'Last 24 Hours' },
                { value: '168', label: 'Last Week' },
                { value: '720', label: 'Last 30 Days' },
              ]}
            />
          </div>
        </div>
      </Card>

      {/* Metrics Display */}
      {selectedVMID ? (
        <>
          <MetricsSummary vmid={selectedVMID} hours={timeRange} />
          <MetricsChart vmid={selectedVMID} hours={timeRange} />
        </>
      ) : (
        <Card className="p-6">
          <p className="text-gray-400">Select a container to view metrics.</p>
        </Card>
      )}

      {/* Info Note */}
      <Card className="p-4 bg-blue-900/20 border-blue-700/50">
        <p className="text-sm text-blue-300">
          <strong>Note:</strong> Analytics data is collected every 30 seconds and retained for 30 days. 
          It may take a few minutes for data to appear for new containers.
        </p>
      </Card>
    </div>
  );
}
