'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Plus, HardDrive, Trash2, Link as LinkIcon, Database, RefreshCw, CheckCircle2, XCircle } from 'lucide-react';
import Card, { CardHeader, CardTitle, CardContent } from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Badge from '@/components/ui/Badge';
import { getVolumes, deleteVolume, getStorage } from '@/lib/api';
import { formatBytes } from '@/lib/utils';
import type { Volume, Storage } from '@/lib/types';

export default function VolumesPage() {
  const [volumes, setVolumes] = useState<Volume[]>([]);
  const [storages, setStorages] = useState<Storage[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [filter, setFilter] = useState<'all' | 'available' | 'in-use'>('all');
  const [showStorage, setShowStorage] = useState(true);

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    try {
      setLoading(true);
      setError(null);
      const [volumesData, storageData] = await Promise.all([
        getVolumes(),
        getStorage({ enabled: true }),
      ]);
      setVolumes(volumesData);
      setStorages(storageData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
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
      await loadData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Delete failed');
    } finally {
      setActionLoading(null);
    }
  }

  function getStorageTypeColor(type: string): string {
    switch (type.toLowerCase()) {
      case 'dir':
        return 'text-blue-500';
      case 'lvm':
      case 'lvmthin':
        return 'text-purple-500';
      case 'zfspool':
        return 'text-green-500';
      case 'nfs':
      case 'cifs':
        return 'text-yellow-500';
      default:
        return 'text-gray-500';
    }
  }

  function getUsageColor(usedFraction?: number): string {
    if (!usedFraction) return 'bg-gray-200';
    if (usedFraction >= 0.9) return 'bg-red-500';
    if (usedFraction >= 0.75) return 'bg-yellow-500';
    return 'bg-green-500';
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
          <div className="text-error mb-2">Error loading data</div>
          <div className="text-text-muted text-sm">{error}</div>
          <Button onClick={loadData} className="mt-4" size="sm">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  // Calculate total storage statistics
  const totalStats = storages.reduce(
    (acc, storage) => {
      if (storage.total) acc.total += storage.total;
      if (storage.used) acc.used += storage.used;
      if (storage.avail) acc.avail += storage.avail;
      return acc;
    },
    { total: 0, used: 0, avail: 0 }
  );

  const totalUsedFraction = totalStats.total > 0 ? totalStats.used / totalStats.total : 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-text-primary">Volumes & Storage</h1>
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm" onClick={loadData}>
            <RefreshCw className="w-4 h-4" />
          </Button>
          <Link href="/volumes/create">
            <Button>
              <Plus className="w-4 h-4 mr-2" />
              Create Volume
            </Button>
          </Link>
        </div>
      </div>

      {/* Storage Overview Section */}
      {storages.length > 0 && (
        <>
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold text-text-primary">Storage Overview</h2>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowStorage(!showStorage)}
            >
              {showStorage ? 'Hide' : 'Show'}
            </Button>
          </div>

          {showStorage && (
            <>
              {/* Storage Summary Cards */}
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <Card>
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="text-text-muted text-sm mb-1">Total Storage</div>
                      <div className="text-2xl font-bold text-text-primary">
                        {formatBytes(totalStats.total)}
                      </div>
                      <div className="text-sm text-text-muted mt-1">
                        {storages.length} datastores
                      </div>
                    </div>
                    <div className="p-3 bg-primary/10 rounded-lg">
                      <Database className="w-5 h-5 text-primary" />
                    </div>
                  </div>
                </Card>

                <Card>
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="text-text-muted text-sm mb-1">Used Space</div>
                      <div className="text-2xl font-bold text-text-primary">
                        {formatBytes(totalStats.used)}
                      </div>
                      <div className="text-sm text-text-muted mt-1">
                        {(totalUsedFraction * 100).toFixed(1)}% used
                      </div>
                    </div>
                    <div className="p-3 bg-warning/10 rounded-lg">
                      <HardDrive className="w-5 h-5 text-warning" />
                    </div>
                  </div>
                </Card>

                <Card>
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="text-text-muted text-sm mb-1">Available Space</div>
                      <div className="text-2xl font-bold text-text-primary">
                        {formatBytes(totalStats.avail)}
                      </div>
                      <div className="text-sm text-text-muted mt-1">
                        Ready to use
                      </div>
                    </div>
                    <div className="p-3 bg-success/10 rounded-lg">
                      <CheckCircle2 className="w-5 h-5 text-success" />
                    </div>
                  </div>
                </Card>

                <Card>
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="text-text-muted text-sm mb-1">Active Stores</div>
                      <div className="text-2xl font-bold text-text-primary">
                        {storages.filter(s => s.active).length}
                      </div>
                      <div className="text-sm text-text-muted mt-1">
                        of {storages.length}
                      </div>
                    </div>
                    <div className="p-3 bg-info/10 rounded-lg">
                      <Database className="w-5 h-5 text-info" />
                    </div>
                  </div>
                </Card>
              </div>

              {/* Storage List */}
              <Card>
                <CardHeader>
                  <CardTitle>Storage Datastores</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {storages.map((storage) => (
                      <div
                        key={storage.storage}
                        className="p-4 rounded-lg border border-border bg-surface hover:bg-surface-elevated transition-colors"
                      >
                        <div className="flex items-start justify-between mb-3">
                          <div className="flex items-center gap-3">
                            <div className={`p-2 rounded-lg bg-surface-elevated ${getStorageTypeColor(storage.type)}`}>
                              <Database className="w-4 h-4" />
                            </div>
                            <div>
                              <div className="flex items-center gap-2">
                                <h3 className="font-semibold text-text-primary text-sm">
                                  {storage.storage}
                                </h3>
                                <Badge variant="default" className="text-xs">
                                  {storage.type}
                                </Badge>
                              </div>
                              <div className="text-xs text-text-muted mt-1">
                                {storage.content}
                              </div>
                            </div>
                          </div>
                          <div className="flex items-center gap-2">
                            {storage.enabled !== false ? (
                              <Badge variant="success" className="text-xs">
                                <CheckCircle2 className="w-3 h-3 mr-1" />
                                Enabled
                              </Badge>
                            ) : (
                              <Badge variant="default" className="text-xs">
                                <XCircle className="w-3 h-3 mr-1" />
                                Disabled
                              </Badge>
                            )}
                            {storage.active !== false && (
                              <Badge variant="success" className="text-xs">Active</Badge>
                            )}
                            {storage.shared && (
                              <Badge variant="info" className="text-xs">Shared</Badge>
                            )}
                          </div>
                        </div>

                        {/* Storage Usage Bar */}
                        {storage.total && storage.total > 0 && (
                          <div className="space-y-2">
                            <div className="flex items-center justify-between text-xs">
                              <span className="text-text-muted">Usage</span>
                              <span className="text-text-primary font-medium">
                                {formatBytes(storage.used || 0)} / {formatBytes(storage.total)}
                              </span>
                            </div>
                            <div className="h-2 bg-surface-elevated rounded-full overflow-hidden">
                              <div
                                className={`h-full transition-all ${getUsageColor(storage.used_fraction)}`}
                                style={{ width: `${(storage.used_fraction || 0) * 100}%` }}
                              />
                            </div>
                            <div className="flex items-center justify-between text-xs text-text-muted">
                              <span>
                                {storage.used_fraction !== undefined
                                  ? `${(storage.used_fraction * 100).toFixed(1)}% used`
                                  : 'N/A'}
                              </span>
                              <span>
                                {formatBytes(storage.avail || 0)} available
                              </span>
                            </div>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </>
          )}
        </>
      )}

      {/* Volumes Section */}
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-text-primary">Volumes</h2>
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
