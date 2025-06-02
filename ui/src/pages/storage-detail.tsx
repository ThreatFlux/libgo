import { useState, useEffect } from 'react';
import { useParams, useNavigate } from '@tanstack/react-router';
import { 
  ArrowLeft, 
  HardDrive, 
  Plus, 
  Trash2, 
  Upload, 
  RefreshCw,
  FileIcon,
  Database 
} from 'lucide-react';
import { Button } from '../components/ui/button';
import { 
  storagePoolApi, 
  storageVolumeApi, 
  formatBytes, 
  getPoolStateColor,
  StoragePoolInfo, 
  StorageVolumeInfo,
  CreateVolumeParams 
} from '../api/storage';

export function StorageDetail() {
  const { poolName } = useParams({ from: '/storage/pools/$poolName' });
  const navigate = useNavigate();
  const [pool, setPool] = useState<StoragePoolInfo | null>(null);
  const [volumes, setVolumes] = useState<StorageVolumeInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [showCreateVolume, setShowCreateVolume] = useState(false);
  const [createVolumeData, setCreateVolumeData] = useState<CreateVolumeParams>({
    name: '',
    capacity_bytes: 10 * 1024 * 1024 * 1024, // 10GB default
    format: 'qcow2',
  });
  const [uploadProgress, setUploadProgress] = useState<{ [key: string]: number }>({});

  useEffect(() => {
    if (poolName) {
      loadPoolAndVolumes();
    }
  }, [poolName]);

  const loadPoolAndVolumes = async () => {
    if (!poolName) return;
    
    try {
      setLoading(true);
      setError(null);
      
      const [poolData, volumesData] = await Promise.all([
        storagePoolApi.get(poolName),
        storageVolumeApi.list(poolName),
      ]);
      
      setPool(poolData);
      setVolumes(volumesData || []);
    } catch (err) {
      setError('Failed to load storage pool details');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateVolume = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!poolName) return;

    try {
      setActionLoading('create');
      await storageVolumeApi.create(poolName, createVolumeData);
      setShowCreateVolume(false);
      setCreateVolumeData({
        name: '',
        capacity_bytes: 10 * 1024 * 1024 * 1024,
        format: 'qcow2',
      });
      await loadPoolAndVolumes();
    } catch (err) {
      console.error('Failed to create volume:', err);
    } finally {
      setActionLoading(null);
    }
  };

  const handleDeleteVolume = async (volumeName: string) => {
    if (!poolName) return;
    
    if (!confirm(`Are you sure you want to delete volume "${volumeName}"?`)) {
      return;
    }

    try {
      setActionLoading(volumeName);
      await storageVolumeApi.delete(poolName, volumeName);
      await loadPoolAndVolumes();
    } catch (err) {
      console.error('Failed to delete volume:', err);
    } finally {
      setActionLoading(null);
    }
  };

  const handleFileUpload = async (volumeName: string, file: File) => {
    if (!poolName) return;

    try {
      setActionLoading(volumeName);
      await storageVolumeApi.upload(poolName, volumeName, file, (progress) => {
        setUploadProgress((prev) => ({ ...prev, [volumeName]: progress }));
      });
      setUploadProgress((prev) => {
        const { [volumeName]: _, ...rest } = prev;
        return rest;
      });
      await loadPoolAndVolumes();
    } catch (err) {
      console.error('Failed to upload file:', err);
    } finally {
      setActionLoading(null);
    }
  };

  const formatVolumeFormat = (format: string) => {
    const formats: { [key: string]: string } = {
      qcow2: 'QCOW2',
      raw: 'RAW',
      vmdk: 'VMDK',
      vdi: 'VDI',
      vpc: 'VPC',
      vhdx: 'VHDX',
    };
    return formats[format] || format.toUpperCase();
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">Loading storage pool...</div>
      </div>
    );
  }

  if (error || !pool) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-destructive">{error || 'Storage pool not found'}</div>
      </div>
    );
  }

  const usagePercentage = pool.capacity > 0 ? Math.round((pool.allocation / pool.capacity) * 100) : 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => navigate({ to: '/storage' })}
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <h1 className="text-3xl font-bold">{pool.name}</h1>
            <span
              className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-${getPoolStateColor(
                pool.state
              )}-500/20 text-${getPoolStateColor(pool.state)}-700`}
            >
              {pool.state}
            </span>
          </div>
          <div className="flex items-center gap-4 mt-2 text-sm text-muted-foreground">
            <span>Type: {pool.type}</span>
            {pool.path && <span>Path: {pool.path}</span>}
            {pool.autostart && <span>Autostart enabled</span>}
          </div>
        </div>
      </div>

      {/* Storage Usage Card */}
      <div className="bg-card rounded-lg border p-6">
        <div className="flex items-center gap-3 mb-4">
          <Database className="h-6 w-6 text-primary" />
          <h2 className="text-xl font-semibold">Storage Usage</h2>
        </div>
        <div className="space-y-3">
          <div>
            <div className="flex items-center justify-between text-sm mb-1">
              <span className="text-muted-foreground">Used Space</span>
              <span>
                {formatBytes(pool.allocation)} / {formatBytes(pool.capacity)} ({usagePercentage}%)
              </span>
            </div>
            <div className="w-full bg-muted rounded-full h-3">
              <div
                className="bg-primary rounded-full h-3 transition-all"
                style={{ width: `${usagePercentage}%` }}
              />
            </div>
            <div className="text-xs text-muted-foreground mt-1">
              {formatBytes(pool.available)} available
            </div>
          </div>
        </div>
      </div>

      {/* Volumes Section */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-semibold">Volumes</h2>
          <div className="flex items-center gap-2">
            <Button onClick={loadPoolAndVolumes} variant="outline" size="sm">
              <RefreshCw className="mr-2 h-4 w-4" />
              Refresh
            </Button>
            <Button onClick={() => setShowCreateVolume(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Create Volume
            </Button>
          </div>
        </div>

        {showCreateVolume && (
          <form onSubmit={handleCreateVolume} className="bg-card rounded-lg border p-6 space-y-4">
            <h3 className="text-lg font-semibold">Create New Volume</h3>
            
            <div>
              <label htmlFor="volume-name" className="block text-sm font-medium mb-2">
                Volume Name
              </label>
              <input
                type="text"
                id="volume-name"
                value={createVolumeData.name}
                onChange={(e) =>
                  setCreateVolumeData({ ...createVolumeData, name: e.target.value })
                }
                className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="my-volume"
                required
              />
            </div>

            <div>
              <label htmlFor="volume-size" className="block text-sm font-medium mb-2">
                Size (GB)
              </label>
              <input
                type="number"
                id="volume-size"
                value={createVolumeData.capacity_bytes / (1024 * 1024 * 1024)}
                onChange={(e) =>
                  setCreateVolumeData({
                    ...createVolumeData,
                    capacity_bytes: Number(e.target.value) * 1024 * 1024 * 1024,
                  })
                }
                className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
                min="1"
                required
              />
            </div>

            <div>
              <label htmlFor="volume-format" className="block text-sm font-medium mb-2">
                Format
              </label>
              <select
                id="volume-format"
                value={createVolumeData.format}
                onChange={(e) =>
                  setCreateVolumeData({ ...createVolumeData, format: e.target.value })
                }
                className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
              >
                <option value="qcow2">QCOW2 (Recommended)</option>
                <option value="raw">RAW</option>
                <option value="vmdk">VMDK</option>
                <option value="vdi">VDI</option>
              </select>
            </div>

            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setShowCreateVolume(false)}
                disabled={actionLoading === 'create'}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={actionLoading === 'create'}>
                {actionLoading === 'create' ? 'Creating...' : 'Create'}
              </Button>
            </div>
          </form>
        )}

        {volumes.length === 0 ? (
          <div className="text-center py-12 bg-muted/50 rounded-lg">
            <HardDrive className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium mb-2">No volumes</h3>
            <p className="text-muted-foreground mb-4">
              This storage pool doesn't have any volumes yet.
            </p>
            <Button onClick={() => setShowCreateVolume(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Create Volume
            </Button>
          </div>
        ) : (
          <div className="grid gap-4">
            {volumes.map((volume) => (
              <div
                key={volume.name}
                className="bg-card rounded-lg border p-4 hover:shadow-lg transition-shadow"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <FileIcon className="h-8 w-8 text-muted-foreground" />
                    <div>
                      <h3 className="font-semibold">{volume.name}</h3>
                      <div className="text-sm text-muted-foreground">
                        {formatBytes(volume.capacity)} • {formatVolumeFormat(volume.format)}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {uploadProgress[volume.name] !== undefined && (
                      <div className="flex items-center gap-2">
                        <div className="w-32 bg-muted rounded-full h-2">
                          <div
                            className="bg-primary rounded-full h-2 transition-all"
                            style={{ width: `${uploadProgress[volume.name]}%` }}
                          />
                        </div>
                        <span className="text-sm">{uploadProgress[volume.name]}%</span>
                      </div>
                    )}
                    <input
                      type="file"
                      id={`upload-${volume.name}`}
                      className="hidden"
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (file) {
                          handleFileUpload(volume.name, file);
                        }
                      }}
                      disabled={actionLoading === volume.name}
                    />
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => document.getElementById(`upload-${volume.name}`)?.click()}
                      disabled={actionLoading === volume.name}
                    >
                      <Upload className="mr-2 h-4 w-4" />
                      Upload
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => handleDeleteVolume(volume.name)}
                      disabled={actionLoading === volume.name}
                    >
                      <Trash2 className="mr-2 h-4 w-4" />
                      Delete
                    </Button>
                  </div>
                </div>
                <div className="mt-2 text-xs text-muted-foreground">
                  <span>Path: {volume.path}</span>
                  {volume.allocation < volume.capacity && (
                    <span className="ml-4">
                      Allocated: {formatBytes(volume.allocation)}
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}