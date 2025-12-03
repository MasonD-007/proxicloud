'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Plus, Play, Square, RotateCw, Trash2 } from 'lucide-react';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Badge from '@/components/ui/Badge';
import { getContainers, startContainer, stopContainer, rebootContainer, deleteContainer } from '@/lib/api';
import { formatBytes, formatCPU, formatUptime } from '@/lib/utils';
import type { Container } from '@/lib/types';

export default function ContainersPage() {
  const [containers, setContainers] = useState<Container[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<number | null>(null);

  useEffect(() => {
    loadContainers();
  }, []);

  async function loadContainers() {
    try {
      setLoading(true);
      setError(null);
      const data = await getContainers();
      setContainers(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load containers');
    } finally {
      setLoading(false);
    }
  }

  async function handleAction(vmid: number, action: 'start' | 'stop' | 'reboot' | 'delete') {
    if (action === 'delete' && !confirm('Are you sure you want to delete this container?')) {
      return;
    }

    try {
      setActionLoading(vmid);
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
          break;
      }
      await loadContainers();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Action failed');
    } finally {
      setActionLoading(null);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-text-secondary">Loading containers...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="text-error mb-2">Error loading containers</div>
          <div className="text-text-muted text-sm">{error}</div>
          <Button onClick={loadContainers} className="mt-4" size="sm">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-text-primary">Containers</h1>
        <Link href="/containers/create">
          <Button>
            <Plus className="w-4 h-4 mr-2" />
            Create Container
          </Button>
        </Link>
      </div>

      {containers.length === 0 ? (
        <Card>
          <div className="text-center py-12">
            <div className="text-text-muted mb-4">No containers found</div>
            <Link href="/containers/create">
              <Button>
                <Plus className="w-4 h-4 mr-2" />
                Create Your First Container
              </Button>
            </Link>
          </div>
        </Card>
      ) : (
        <div className="bg-surface rounded-lg border border-border overflow-hidden">
          <table className="w-full">
            <thead className="bg-surface-elevated border-b border-border">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  ID
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  CPU
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  Memory
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  Uptime
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-text-muted uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {containers.map((container) => (
                <tr key={container.vmid} className="hover:bg-surface-elevated">
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-text-primary">
                    {container.vmid}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link href={`/containers/${container.vmid}`}>
                      <span className="text-sm font-medium text-primary hover:underline">
                        {container.name}
                      </span>
                    </Link>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Badge variant={container.status === 'running' ? 'success' : 'default'}>
                      {container.status}
                    </Badge>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-text-secondary">
                    {formatCPU(container.cpu)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-text-secondary">
                    {formatBytes(container.mem)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-text-secondary">
                    {container.status === 'running' ? formatUptime(container.uptime) : '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <div className="flex items-center justify-end gap-2">
                      {container.status === 'stopped' ? (
                        <button
                          onClick={() => handleAction(container.vmid, 'start')}
                          disabled={actionLoading === container.vmid}
                          className="p-2 text-success hover:bg-success/10 rounded transition-colors disabled:opacity-50"
                          title="Start"
                        >
                          <Play className="w-4 h-4" />
                        </button>
                      ) : (
                        <button
                          onClick={() => handleAction(container.vmid, 'stop')}
                          disabled={actionLoading === container.vmid}
                          className="p-2 text-warning hover:bg-warning/10 rounded transition-colors disabled:opacity-50"
                          title="Stop"
                        >
                          <Square className="w-4 h-4" />
                        </button>
                      )}
                      <button
                        onClick={() => handleAction(container.vmid, 'reboot')}
                        disabled={actionLoading === container.vmid || container.status === 'stopped'}
                        className="p-2 text-info hover:bg-info/10 rounded transition-colors disabled:opacity-50"
                        title="Reboot"
                      >
                        <RotateCw className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => handleAction(container.vmid, 'delete')}
                        disabled={actionLoading === container.vmid}
                        className="p-2 text-error hover:bg-error/10 rounded transition-colors disabled:opacity-50"
                        title="Delete"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
