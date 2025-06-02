import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { ArrowLeft, HardDrive } from 'lucide-react';
import { Button } from '../components/ui/button';
import { storagePoolApi, CreatePoolParams } from '../api/storage';

export function StorageCreate() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [formData, setFormData] = useState<CreatePoolParams>({
    name: '',
    type: 'dir',
    path: '',
    autostart: true,
    metadata: {},
  });

  const poolTypes = [
    { value: 'dir', label: 'Directory', requiresPath: true },
    { value: 'fs', label: 'Filesystem', requiresPath: true },
    { value: 'netfs', label: 'Network Filesystem', requiresPath: false },
    { value: 'logical', label: 'Logical Volume', requiresPath: false },
    { value: 'disk', label: 'Disk', requiresPath: false },
    { value: 'iscsi', label: 'iSCSI', requiresPath: false },
    { value: 'rbd', label: 'RBD (Ceph)', requiresPath: false },
    { value: 'gluster', label: 'GlusterFS', requiresPath: false },
    { value: 'zfs', label: 'ZFS', requiresPath: true },
  ];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.name) {
      setError('Pool name is required');
      return;
    }

    const selectedType = poolTypes.find((t) => t.value === formData.type);
    if (selectedType?.requiresPath && !formData.path) {
      setError(`Path is required for ${selectedType.label} pools`);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      await storagePoolApi.create(formData);
      navigate({ to: '/storage' });
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create storage pool');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const selectedType = poolTypes.find((t) => t.value === formData.type);

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => navigate({ to: '/storage' })}
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h1 className="text-3xl font-bold">Create Storage Pool</h1>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-card rounded-lg border p-6 space-y-6">
          <div className="flex items-center gap-3 mb-4">
            <HardDrive className="h-6 w-6 text-primary" />
            <h2 className="text-xl font-semibold">Pool Configuration</h2>
          </div>

          {error && (
            <div className="bg-destructive/15 text-destructive px-4 py-3 rounded-lg">
              {error}
            </div>
          )}

          <div>
            <label htmlFor="name" className="block text-sm font-medium mb-2">
              Pool Name
            </label>
            <input
              type="text"
              id="name"
              value={formData.name}
              onChange={(e) =>
                setFormData({ ...formData, name: e.target.value })
              }
              className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
              placeholder="my-storage-pool"
              required
            />
            <p className="text-sm text-muted-foreground mt-1">
              A unique name for the storage pool
            </p>
          </div>

          <div>
            <label htmlFor="type" className="block text-sm font-medium mb-2">
              Pool Type
            </label>
            <select
              id="type"
              value={formData.type}
              onChange={(e) =>
                setFormData({ ...formData, type: e.target.value })
              }
              className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
            >
              {poolTypes.map((type) => (
                <option key={type.value} value={type.value}>
                  {type.label}
                </option>
              ))}
            </select>
            <p className="text-sm text-muted-foreground mt-1">
              The type of storage pool to create
            </p>
          </div>

          {selectedType?.requiresPath && (
            <div>
              <label htmlFor="path" className="block text-sm font-medium mb-2">
                Storage Path
              </label>
              <input
                type="text"
                id="path"
                value={formData.path}
                onChange={(e) =>
                  setFormData({ ...formData, path: e.target.value })
                }
                className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="/var/lib/libvirt/images"
                required={selectedType?.requiresPath}
              />
              <p className="text-sm text-muted-foreground mt-1">
                The directory path where storage volumes will be created
              </p>
            </div>
          )}

          {formData.type === 'netfs' && (
            <div className="space-y-4 border-t pt-4">
              <h3 className="font-medium">Network Filesystem Settings</h3>
              <div>
                <label htmlFor="host" className="block text-sm font-medium mb-2">
                  Server Host
                </label>
                <input
                  type="text"
                  id="host"
                  value={formData.source?.host || ''}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      source: { ...formData.source, host: e.target.value },
                    })
                  }
                  className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
                  placeholder="nfs-server.example.com"
                />
              </div>
              <div>
                <label htmlFor="dir" className="block text-sm font-medium mb-2">
                  Export Directory
                </label>
                <input
                  type="text"
                  id="dir"
                  value={formData.source?.dir || ''}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      source: { ...formData.source, dir: e.target.value },
                    })
                  }
                  className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
                  placeholder="/exports/vms"
                />
              </div>
            </div>
          )}

          <div>
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={formData.autostart}
                onChange={(e) =>
                  setFormData({ ...formData, autostart: e.target.checked })
                }
                className="rounded border-gray-300"
              />
              <span className="text-sm font-medium">
                Start pool automatically
              </span>
            </label>
            <p className="text-sm text-muted-foreground mt-1 ml-6">
              Automatically start this pool when the host system boots
            </p>
          </div>
        </div>

        <div className="flex justify-end gap-4">
          <Button
            type="button"
            variant="outline"
            onClick={() => navigate({ to: '/storage' })}
            disabled={loading}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={loading}>
            {loading ? 'Creating...' : 'Create Pool'}
          </Button>
        </div>
      </form>
    </div>
  );
}