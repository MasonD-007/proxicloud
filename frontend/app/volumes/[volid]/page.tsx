'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import {
  ArrowLeft,
  HardDrive,
  Trash2,
  Link as LinkIcon,
  Unlink,
  Camera,
  RotateCcw,
  Copy,
  AlertCircle,
} from 'lucide-react';
import Link from 'next/link';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Badge from '@/components/ui/Badge';
import Input from '@/components/ui/Input';
import {
  getVolume,
  deleteVolume,
  attachVolume,
  detachVolume,
  getSnapshots,
  createSnapshot,
  restoreSnapshot,
  cloneSnapshot,
  getContainers,
} from '@/lib/api';
import { formatBytes } from '@/lib/utils';
import type { Volume, Snapshot, Container } from '@/lib/types';

export default function VolumeDetailPage() {
  const router = useRouter();
  const params = useParams();
  const volid = decodeURIComponent(params.volid as string);

  const [volume, setVolume] = useState<Volume | null>(null);
  const [snapshots, setSnapshots] = useState<Snapshot[]>([]);
  const [containers, setContainers] = useState<Container[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Snapshot form state
  const [showSnapshotForm, setShowSnapshotForm] = useState(false);
  const [snapshotName, setSnapshotName] = useState('');
  const [snapshotDescription, setSnapshotDescription] = useState('');

  // Attach form state
  const [showAttachForm, setShowAttachForm] = useState(false);
  const [selectedContainer, setSelectedContainer] = useState<number | null>(null);

  useEffect(() => {
    if (volid) {
      loadData();
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [volid]);

  async function loadData() {
    try {
      setError(null);
      const [volumeData, snapshotsData, containersData] = await Promise.all([
        getVolume(volid),
        getSnapshots(volid).catch(() => []),
        getContainers(),
      ]);
      setVolume(volumeData);
      setSnapshots(snapshotsData);
      setContainers(containersData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load volume');
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete() {
    if (!volume) return;
    if (volume.status === 'in-use') {
      alert('Cannot delete a volume that is in use. Please detach it first.');
      return;
    }
    if (!confirm(`Are you sure you want to delete volume ${volume.name}? This action cannot be undone.`)) {
      return;
    }

    try {
      setActionLoading(true);
      await deleteVolume(volid);
      router.push('/volumes');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Delete failed');
      setActionLoading(false);
    }
  }

  async function handleAttach() {
    if (!selectedContainer) return;

    try {
      setActionLoading(true);
      await attachVolume(volid, selectedContainer);
      setShowAttachForm(false);
      setSelectedContainer(null);
      await loadData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Attach failed');
    } finally {
      setActionLoading(false);
    }
  }

  async function handleDetach() {
    if (!volume?.attached_to) return;
    if (!confirm('Are you sure you want to detach this volume?')) return;

    try {
      setActionLoading(true);
      await detachVolume(volid, volume.attached_to);
      await loadData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Detach failed');
    } finally {
      setActionLoading(false);
    }
  }

  async function handleCreateSnapshot() {
    if (!snapshotName.trim()) {
      alert('Please enter a snapshot name');
      return;
    }

    try {
      setActionLoading(true);
      await createSnapshot(volid, {
        name: snapshotName,
        description: snapshotDescription,
      });
      setShowSnapshotForm(false);
      setSnapshotName('');
      setSnapshotDescription('');
      await loadData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Snapshot creation failed');
    } finally {
      setActionLoading(false);
    }
  }

  async function handleRestoreSnapshot(snapshotName: string) {
    if (!confirm(`Are you sure you want to restore from snapshot "${snapshotName}"? Current data will be lost.`)) {
      return;
    }

    try {
      setActionLoading(true);
      await restoreSnapshot(volid, { snapshot_name: snapshotName });
      alert('Snapshot restored successfully');
      await loadData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Restore failed');
    } finally {
      setActionLoading(false);
    }
  }

  async function handleCloneSnapshot(snapshotName: string) {
    const newName = prompt('Enter name for the cloned volume:');
    if (!newName) return;

    try {
      setActionLoading(true);
      const newVolume = await cloneSnapshot(volid, {
        snapshot_name: snapshotName,
        new_name: newName,
      });
      alert(`Volume cloned successfully as ${newVolume.name}`);
      router.push(`/volumes/${encodeURIComponent(newVolume.volid)}`);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Clone failed');
      setActionLoading(false);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-text-secondary">Loading volume...</div>
      </div>
    );
  }

  if (error || !volume) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="text-error mb-2">Error loading volume</div>
          <div className="text-text-muted text-sm">{error || 'Volume not found'}</div>
          <Link href="/volumes">
            <Button className="mt-4" size="sm">
              Back to Volumes
            </Button>
          </Link>
        </div>
      </div>
    );
  }

  const usagePercent = volume.size > 0 ? ((volume.used / volume.size) * 100).toFixed(1) : '0';
  const availableContainers = containers.filter(c => c.status === 'running');

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link href="/volumes">
            <Button variant="ghost" size="sm">
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back
            </Button>
          </Link>
          <div>
            <div className="flex items-center gap-3">
              <HardDrive className="w-8 h-8 text-text-secondary" />
              <h1 className="text-3xl font-bold text-text-primary">{volume.name}</h1>
              <Badge
                variant={
                  volume.status === 'available'
                    ? 'success'
                    : volume.status === 'in-use'
                    ? 'info'
                    : 'error'
                }
              >
                {volume.status}
              </Badge>
            </div>
            <p className="text-text-muted mt-1">
              {volume.volid} on {volume.storage} ({volume.type.toUpperCase()})
            </p>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center gap-2">
          {volume.status === 'available' ? (
            <Button onClick={() => setShowAttachForm(true)} disabled={actionLoading}>
              <LinkIcon className="w-4 h-4 mr-2" />
              Attach
            </Button>
          ) : (
            <Button onClick={handleDetach} disabled={actionLoading} variant="secondary">
              <Unlink className="w-4 h-4 mr-2" />
              Detach
            </Button>
          )}
          <Button
            onClick={handleDelete}
            disabled={actionLoading || volume.status === 'in-use'}
            variant="danger"
          >
            <Trash2 className="w-4 h-4 mr-2" />
            Delete
          </Button>
        </div>
      </div>

      {/* Volume Info Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <div className="space-y-2">
            <div className="text-text-muted text-sm">Total Size</div>
            <div className="text-2xl font-bold text-text-primary">{volume.size} GB</div>
            <div className="text-text-muted text-xs">
              {formatBytes(volume.size * 1024 * 1024 * 1024)}
            </div>
          </div>
        </Card>

        <Card>
          <div className="space-y-2">
            <div className="text-text-muted text-sm">Used Space</div>
            <div className="text-2xl font-bold text-text-primary">
              {volume.used > 0 ? `${volume.used} GB` : '0 GB'}
            </div>
            <div className="flex items-center gap-2">
              <div className="flex-1 bg-surface-elevated rounded-full h-2">
                <div
                  className="bg-primary rounded-full h-2 transition-all"
                  style={{ width: `${usagePercent}%` }}
                />
              </div>
              <span className="text-text-muted text-xs">{usagePercent}%</span>
            </div>
          </div>
        </Card>

        <Card>
          <div className="space-y-2">
            <div className="text-text-muted text-sm">Attached To</div>
            <div className="text-2xl font-bold text-text-primary">
              {volume.attached_to ? (
                <Link href={`/containers/${volume.attached_to}`}>
                  <span className="text-primary hover:underline">VM {volume.attached_to}</span>
                </Link>
              ) : (
                <span className="text-text-muted">Not attached</span>
              )}
            </div>
            {volume.mountpoint && (
              <div className="text-text-muted text-xs">Mount: {volume.mountpoint}</div>
            )}
          </div>
        </Card>
      </div>

      {/* Attach Form */}
      {showAttachForm && (
        <Card>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold text-text-primary">Attach Volume</h2>
              <Button
                onClick={() => setShowAttachForm(false)}
                variant="ghost"
                size="sm"
              >
                Cancel
              </Button>
            </div>

            {availableContainers.length === 0 ? (
              <div className="text-center py-6 text-text-muted">
                No running containers available. Start a container first.
              </div>
            ) : (
              <>
                <div>
                  <label className="block text-sm font-medium text-text-primary mb-2">
                    Select Container
                  </label>
                  <select
                    value={selectedContainer || ''}
                    onChange={(e) => setSelectedContainer(parseInt(e.target.value))}
                    className="w-full px-3 py-2 bg-surface border border-border rounded-lg text-text-primary focus:outline-none focus:ring-2 focus:ring-primary"
                  >
                    <option value="">Choose a container...</option>
                    {availableContainers.map((container) => (
                      <option key={container.vmid} value={container.vmid}>
                        {container.name} (VM {container.vmid})
                      </option>
                    ))}
                  </select>
                </div>

                <Button
                  onClick={handleAttach}
                  disabled={!selectedContainer || actionLoading}
                  className="w-full"
                >
                  Attach Volume
                </Button>
              </>
            )}
          </div>
        </Card>
      )}

      {/* Snapshots Section */}
      <Card>
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold text-text-primary">Snapshots</h2>
            <Button
              onClick={() => setShowSnapshotForm(!showSnapshotForm)}
              size="sm"
              disabled={volume.status === 'in-use'}
            >
              <Camera className="w-4 h-4 mr-2" />
              Create Snapshot
            </Button>
          </div>

          {volume.status === 'in-use' && (
            <div className="flex items-start gap-2 p-3 bg-warning/10 border border-warning/20 rounded-lg">
              <AlertCircle className="w-5 h-5 text-warning flex-shrink-0 mt-0.5" />
              <div className="text-sm text-text-secondary">
                Volume must be detached before creating snapshots. Detach the volume to enable snapshot operations.
              </div>
            </div>
          )}

          {/* Snapshot Form */}
          {showSnapshotForm && (
            <div className="p-4 bg-surface-elevated rounded-lg border border-border space-y-4">
              <div>
                <label className="block text-sm font-medium text-text-primary mb-2">
                  Snapshot Name *
                </label>
                <Input
                  value={snapshotName}
                  onChange={(e) => setSnapshotName(e.target.value)}
                  placeholder="snapshot-1"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-text-primary mb-2">
                  Description (Optional)
                </label>
                <Input
                  value={snapshotDescription}
                  onChange={(e) => setSnapshotDescription(e.target.value)}
                  placeholder="Backup before update..."
                />
              </div>

              <div className="flex gap-2">
                <Button onClick={handleCreateSnapshot} disabled={actionLoading}>
                  Create
                </Button>
                <Button
                  onClick={() => {
                    setShowSnapshotForm(false);
                    setSnapshotName('');
                    setSnapshotDescription('');
                  }}
                  variant="ghost"
                >
                  Cancel
                </Button>
              </div>
            </div>
          )}

          {/* Snapshots List */}
          {snapshots.length === 0 ? (
            <div className="text-center py-12 text-text-muted">
              <Camera className="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p>No snapshots yet</p>
              <p className="text-sm mt-2">
                Create snapshots to save the current state of your volume
              </p>
            </div>
          ) : (
            <div className="space-y-2">
              {snapshots.map((snapshot) => (
                <div
                  key={snapshot.name}
                  className="flex items-center justify-between p-4 bg-surface-elevated rounded-lg border border-border hover:border-primary/50 transition-colors"
                >
                  <div className="flex-1">
                    <div className="font-medium text-text-primary">{snapshot.name}</div>
                    {snapshot.description && (
                      <div className="text-sm text-text-muted mt-1">{snapshot.description}</div>
                    )}
                    <div className="text-xs text-text-muted mt-1">
                      Created: {new Date(snapshot.created_at * 1000).toLocaleString()}
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => handleRestoreSnapshot(snapshot.name)}
                      disabled={actionLoading}
                      className="p-2 text-info hover:bg-info/10 rounded transition-colors disabled:opacity-50"
                      title="Restore"
                    >
                      <RotateCcw className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => handleCloneSnapshot(snapshot.name)}
                      disabled={actionLoading}
                      className="p-2 text-success hover:bg-success/10 rounded transition-colors disabled:opacity-50"
                      title="Clone"
                    >
                      <Copy className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </Card>

      {/* Volume Details */}
      <Card>
        <h2 className="text-xl font-semibold text-text-primary mb-4">Volume Details</h2>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <div className="text-sm text-text-muted">Volume ID</div>
            <div className="text-text-primary font-mono text-sm">{volume.volid}</div>
          </div>
          <div>
            <div className="text-sm text-text-muted">Storage Pool</div>
            <div className="text-text-primary">{volume.storage}</div>
          </div>
          <div>
            <div className="text-sm text-text-muted">Format</div>
            <div className="text-text-primary uppercase">{volume.format}</div>
          </div>
          <div>
            <div className="text-sm text-text-muted">Node</div>
            <div className="text-text-primary">{volume.node}</div>
          </div>
          {volume.created_at && (
            <div>
              <div className="text-sm text-text-muted">Created</div>
              <div className="text-text-primary">
                {new Date(volume.created_at * 1000).toLocaleString()}
              </div>
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}
