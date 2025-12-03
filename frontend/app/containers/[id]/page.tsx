'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { ArrowLeft, Server, Play, Square, RotateCw, Trash2, HardDrive, Cpu, MemoryStick, Network } from 'lucide-react';
import Link from 'next/link';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Badge from '@/components/ui/Badge';
import { MetricsChart } from '@/components/analytics/MetricsChart';
import { MetricsSummary } from '@/components/analytics/MetricsSummary';
import { getContainer, startContainer, stopContainer, rebootContainer, deleteContainer } from '@/lib/api';
import { formatBytes, formatCPU, formatUptime } from '@/lib/utils';
import type { Container } from '@/lib/types';

export default function ContainerDetailPage() {
  const router = useRouter();
  const params = useParams();
  const vmid = parseInt(params.id as string);

  const [container, setContainer] = useState<Container | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (vmid) {
      loadContainer();
      const interval = setInterval(loadContainer, 5000); // Refresh every 5 seconds
      return () => clearInterval(interval);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [vmid]);

  async function loadContainer() {
    try {
      setError(null);
      const data = await getContainer(vmid);
      setContainer(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load container');
    } finally {
      setLoading(false);
    }
  }

  async function handleAction(action: 'start' | 'stop' | 'reboot' | 'delete') {
    if (action === 'delete') {
      if (!confirm(`Are you sure you want to delete container ${container?.name} (${vmid})?`)) {
        return;
      }
    }

    try {
      setActionLoading(true);
      switch (action) {
        case 'start':
          await startContainer(vmid);
          break;
        case 'stop':
          await stopContainer(vmid);
          break;
        case 'reboot':
          await rebootContainer(vmid);
          break;
        case 'delete':
          await deleteContainer(vmid);
          router.push('/containers');
          return;
      }
      await loadContainer();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Action failed');
    } finally {
      setActionLoading(false);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-text-secondary">Loading container...</div>
      </div>
    );
  }

  if (error || !container) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="text-error mb-2">Error loading container</div>
          <div className="text-text-muted text-sm">{error || 'Container not found'}</div>
          <Link href="/containers">
            <Button className="mt-4" size="sm">
              Back to Containers
            </Button>
          </Link>
        </div>
      </div>
    );
  }

  const cpuPercent = (container.cpu * 100).toFixed(1);
  const memPercent = ((container.mem / container.maxmem) * 100).toFixed(1);
  const diskPercent = ((container.disk / container.maxdisk) * 100).toFixed(1);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link href="/containers">
            <Button variant="ghost" size="sm">
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back
            </Button>
          </Link>
          <div>
            <div className="flex items-center gap-3">
              <Server className="w-8 h-8 text-text-secondary" />
              <h1 className="text-3xl font-bold text-text-primary">{container.name}</h1>
              <Badge variant={container.status === 'running' ? 'success' : 'default'}>
                {container.status}
              </Badge>
            </div>
            <p className="text-text-muted mt-1">
              Container {vmid} on node {container.node}
            </p>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center gap-2">
          {container.status === 'stopped' ? (
            <Button onClick={() => handleAction('start')} disabled={actionLoading}>
              <Play className="w-4 h-4 mr-2" />
              Start
            </Button>
          ) : (
            <Button onClick={() => handleAction('stop')} disabled={actionLoading} variant="secondary">
              <Square className="w-4 h-4 mr-2" />
              Stop
            </Button>
          )}
          <Button
            onClick={() => handleAction('reboot')}
            disabled={actionLoading || container.status === 'stopped'}
            variant="secondary"
          >
            <RotateCw className="w-4 h-4 mr-2" />
            Reboot
          </Button>
          <Button
            onClick={() => handleAction('delete')}
            disabled={actionLoading}
            variant="danger"
          >
            <Trash2 className="w-4 h-4 mr-2" />
            Delete
          </Button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {/* CPU */}
        <Card>
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <Cpu className="w-5 h-5 text-primary" />
              <h3 className="font-semibold text-text-primary">CPU</h3>
            </div>
            <span className="text-2xl font-bold text-text-primary">{cpuPercent}%</span>
          </div>
          <div className="w-full bg-surface-elevated rounded-full h-2 overflow-hidden">
            <div
              className="bg-primary h-full transition-all duration-300"
              style={{ width: `${cpuPercent}%` }}
            />
          </div>
          <p className="text-sm text-text-muted mt-2">{formatCPU(container.cpu)}</p>
        </Card>

        {/* Memory */}
        <Card>
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <MemoryStick className="w-5 h-5 text-success" />
              <h3 className="font-semibold text-text-primary">Memory</h3>
            </div>
            <span className="text-2xl font-bold text-text-primary">{memPercent}%</span>
          </div>
          <div className="w-full bg-surface-elevated rounded-full h-2 overflow-hidden">
            <div
              className="bg-success h-full transition-all duration-300"
              style={{ width: `${memPercent}%` }}
            />
          </div>
          <p className="text-sm text-text-muted mt-2">
            {formatBytes(container.mem)} / {formatBytes(container.maxmem)}
          </p>
        </Card>

        {/* Disk */}
        <Card>
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <HardDrive className="w-5 h-5 text-warning" />
              <h3 className="font-semibold text-text-primary">Disk</h3>
            </div>
            <span className="text-2xl font-bold text-text-primary">{diskPercent}%</span>
          </div>
          <div className="w-full bg-surface-elevated rounded-full h-2 overflow-hidden">
            <div
              className="bg-warning h-full transition-all duration-300"
              style={{ width: `${diskPercent}%` }}
            />
          </div>
          <p className="text-sm text-text-muted mt-2">
            {formatBytes(container.disk)} / {formatBytes(container.maxdisk)}
          </p>
        </Card>

        {/* Uptime */}
        <Card>
          <div className="flex items-center gap-2 mb-2">
            <Network className="w-5 h-5 text-info" />
            <h3 className="font-semibold text-text-primary">Uptime</h3>
          </div>
          <div className="text-2xl font-bold text-text-primary mt-4">
            {container.status === 'running' ? formatUptime(container.uptime) : 'Stopped'}
          </div>
          <p className="text-sm text-text-muted mt-2">
            {container.status === 'running' ? 'Running' : 'Not running'}
          </p>
        </Card>
      </div>

      {/* Configuration */}
      <Card>
        <h2 className="text-xl font-semibold text-text-primary mb-4">Configuration</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <dt className="text-sm font-medium text-text-muted">VM ID</dt>
            <dd className="text-text-primary mt-1">{container.vmid}</dd>
          </div>
          <div>
            <dt className="text-sm font-medium text-text-muted">Name</dt>
            <dd className="text-text-primary mt-1">{container.name}</dd>
          </div>
          <div>
            <dt className="text-sm font-medium text-text-muted">Node</dt>
            <dd className="text-text-primary mt-1">{container.node}</dd>
          </div>
          <div>
            <dt className="text-sm font-medium text-text-muted">Status</dt>
            <dd className="mt-1">
              <Badge variant={container.status === 'running' ? 'success' : 'default'}>
                {container.status}
              </Badge>
            </dd>
          </div>
          {container.template && (
            <div>
              <dt className="text-sm font-medium text-text-muted">Template</dt>
              <dd className="text-text-primary mt-1">{container.template}</dd>
            </div>
          )}
          {container.os && (
            <div>
              <dt className="text-sm font-medium text-text-muted">Operating System</dt>
              <dd className="text-text-primary mt-1">{container.os}</dd>
            </div>
          )}
        </div>
      </Card>

      {/* Resource Details */}
      <Card>
        <h2 className="text-xl font-semibold text-text-primary mb-4">Resource Details</h2>
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <dt className="text-sm font-medium text-text-muted">CPU Usage</dt>
              <dd className="text-text-primary mt-1">{formatCPU(container.cpu)}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-text-muted">CPU Load</dt>
              <dd className="text-text-primary mt-1">{cpuPercent}%</dd>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <dt className="text-sm font-medium text-text-muted">Memory Used</dt>
              <dd className="text-text-primary mt-1">{formatBytes(container.mem)}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-text-muted">Memory Total</dt>
              <dd className="text-text-primary mt-1">{formatBytes(container.maxmem)}</dd>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <dt className="text-sm font-medium text-text-muted">Disk Used</dt>
              <dd className="text-text-primary mt-1">{formatBytes(container.disk)}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-text-muted">Disk Total</dt>
              <dd className="text-text-primary mt-1">{formatBytes(container.maxdisk)}</dd>
            </div>
          </div>
        </div>
      </Card>

      {/* Quick Actions */}
      <Card className="bg-surface-elevated">
        <h2 className="text-xl font-semibold text-text-primary mb-4">Quick Actions</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Button
            onClick={() => handleAction('start')}
            disabled={actionLoading || container.status === 'running'}
            className="w-full"
          >
            <Play className="w-4 h-4 mr-2" />
            Start
          </Button>
          <Button
            onClick={() => handleAction('stop')}
            disabled={actionLoading || container.status === 'stopped'}
            variant="secondary"
            className="w-full"
          >
            <Square className="w-4 h-4 mr-2" />
            Stop
          </Button>
          <Button
            onClick={() => handleAction('reboot')}
            disabled={actionLoading || container.status === 'stopped'}
            variant="secondary"
            className="w-full"
          >
            <RotateCw className="w-4 h-4 mr-2" />
            Reboot
          </Button>
          <Button
            onClick={() => handleAction('delete')}
            disabled={actionLoading}
            variant="danger"
            className="w-full"
          >
            <Trash2 className="w-4 h-4 mr-2" />
            Delete
          </Button>
        </div>
      </Card>

      {/* Analytics Section */}
      <div className="space-y-6">
        <h2 className="text-2xl font-semibold text-text-primary">Performance Analytics</h2>
        
        {/* Metrics Summary */}
        <MetricsSummary vmid={vmid} hours={24} />
        
        {/* Historical Charts */}
        <MetricsChart vmid={vmid} hours={24} />
      </div>
    </div>
  );
}
