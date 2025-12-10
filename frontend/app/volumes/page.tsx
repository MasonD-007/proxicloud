'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Plus, HardDrive, Trash2, Link as LinkIcon, Unlink } from 'lucide-react';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Badge from '@/components/ui/Badge';
import { getVolumes, deleteVolume } from '@/lib/api';
import { formatBytes } from '@/lib/utils';
import type { Volume } from '@/lib/types';

export default function VolumesPage() {
  const [volumes, setVolumes] = useState<Volume[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [filter, setFilter] = useState<'all' | 'available' | 'in-use'>('all');

  useEffect(() => {
    loadVolumes();
  }, []);

  async function loadVolumes() {
    try {
      setLoading(true);
      setError(null);
      const data = await getVolumes();
      setVolumes(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load volumes');
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete(volid: string) {
    if (!confirm('Are you sure you want to delete this volume? This action cannot be undone.')) {
      return;
    }

    try {
      setActionLoading(volid);
      await deleteVolume(volid);
      await loadVolumes();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Delete failed');
    } finally {
      setActionLoading(null);
    }
  }

  const filteredVolumes = volumes.filter((vol) => {
    if (filter === 'all') return true;
    return vol.status === filter;
  });

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-text-secondary">Loading volumes...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="text-error mb-2">Error loading volumes</div>
          <div className="text-text-muted text-sm">{error}</div>
          <Button onClick={loadVolumes} className="mt-4" size="sm">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-text-primary">Volumes</h1>
        <Link href="/volumes/create">
          <Button>
            <Plus className="w-4 h-4 mr-2" />
            Create Volume
          </Button>
        </Link>
      </div>

      {/* Filter Tabs */}
      <div className="flex gap-2">
        <button
          onClick={() => setFilter('all')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            filter === 'all'
              ? 'bg-primary text-white'
              : 'bg-surface-elevated text-text-secondary hover:bg-surface-elevated/80'
          }`}
        >
          All ({volumes.length})
        </button>
        <button
          onClick={() => setFilter('available')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            filter === 'available'
              ? 'bg-primary text-white'
              : 'bg-surface-elevated text-text-secondary hover:bg-surface-elevated/80'
          }`}
        >
          Available ({volumes.filter((v) => v.status === 'available').length})
        </button>
        <button
          onClick={() => setFilter('in-use')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            filter === 'in-use'
              ? 'bg-primary text-white'
              : 'bg-surface-elevated text-text-secondary hover:bg-surface-elevated/80'
          }`}
        >
          In Use ({volumes.filter((v) => v.status === 'in-use').length})
        </button>
      </div>

      {filteredVolumes.length === 0 ? (
        <Card>
          <div className="text-center py-12">
            <HardDrive className="w-12 h-12 mx-auto mb-4 text-text-muted" />
            <div className="text-text-muted mb-4">
              {filter === 'all' ? 'No volumes found' : `No ${filter} volumes`}
            </div>
            {filter === 'all' && (
              <Link href="/volumes/create">
                <Button>
                  <Plus className="w-4 h-4 mr-2" />
                  Create Your First Volume
                </Button>
              </Link>
            )}
          </div>
        </Card>
      ) : (
        <div className="bg-surface rounded-lg border border-border overflow-hidden">
          <table className="w-full">
            <thead className="bg-surface-elevated border-b border-border">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  Size
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  Storage
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  Attached To
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-text-muted uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {filteredVolumes.map((volume) => (
                <tr key={volume.volid} className="hover:bg-surface-elevated">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link href={`/volumes/${encodeURIComponent(volume.volid)}`}>
                      <div className="flex items-center">
                        <HardDrive className="w-4 h-4 mr-2 text-text-muted" />
                        <span className="text-sm font-medium text-primary hover:underline">
                          {volume.name}
                        </span>
                      </div>
                    </Link>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Badge
                      variant={
                        volume.status === 'available'
                          ? 'success'
                          : volume.status === 'in-use'
                          ? 'info'
                          : 'error'
                      }
                    >
                      {volume.status === 'in-use' ? (
                        <span className="flex items-center gap-1">
                          <LinkIcon className="w-3 h-3" />
                          In Use
                        </span>
                      ) : (
                        volume.status
                      )}
                    </Badge>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-text-secondary">
                    {volume.size} GB
                    {volume.used > 0 && (
                      <span className="text-xs text-text-muted ml-1">
                        ({formatBytes(volume.used * 1024 * 1024 * 1024)} used)
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-text-secondary uppercase">
                    {volume.type}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-text-secondary">
                    {volume.storage}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-text-secondary">
                    {volume.attached_to ? (
                      <Link href={`/containers/${volume.attached_to}`}>
                        <span className="text-primary hover:underline">
                          VM {volume.attached_to}
                        </span>
                      </Link>
                    ) : (
                      '-'
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        onClick={() => handleDelete(volume.volid)}
                        disabled={
                          actionLoading === volume.volid ||
                          volume.status === 'in-use'
                        }
                        className="p-2 text-error hover:bg-error/10 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                        title={
                          volume.status === 'in-use'
                            ? 'Detach volume before deleting'
                            : 'Delete'
                        }
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
