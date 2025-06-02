import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Plus, HardDrive, Play, Square, Trash2, RefreshCw } from 'lucide-react';
import { Button } from '../components/ui/button';
import { storagePoolApi, formatBytes, getPoolStateColor, StoragePoolInfo } from '../api/storage';

export function StorageList() {
  const navigate = useNavigate();
  const [pools, setPools] = useState<StoragePoolInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  useEffect(() => {
    loadPools();
  }, []);

  const loadPools = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await storagePoolApi.list();
      setPools(data || []);
    } catch (err) {
      setError('Failed to load storage pools');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleStart = async (poolName: string) => {
    try {
      setActionLoading(poolName);
      await storagePoolApi.start(poolName);
      await loadPools();
    } catch (err) {
      console.error('Failed to start pool:', err);
    } finally {
      setActionLoading(null);
    }
  };

  const handleStop = async (poolName: string) => {
    try {
      setActionLoading(poolName);
      await storagePoolApi.stop(poolName);
      await loadPools();
    } catch (err) {
      console.error('Failed to stop pool:', err);
    } finally {
      setActionLoading(null);
    }
  };

  const handleDelete = async (poolName: string) => {
    if (!confirm(`Are you sure you want to delete storage pool "${poolName}"?`)) {
      return;
    }

    try {
      setActionLoading(poolName);
      await storagePoolApi.delete(poolName);
      await loadPools();
    } catch (err) {
      console.error('Failed to delete pool:', err);
    } finally {
      setActionLoading(null);
    }
  };

  const getUsagePercentage = (pool: StoragePoolInfo) => {
    if (pool.capacity === 0) return 0;
    return Math.round((pool.allocation / pool.capacity) * 100);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">Loading storage pools...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-destructive">{error}</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Storage Pools</h1>
        <Button onClick={() => navigate('/storage/create')}>
          <Plus className="mr-2 h-4 w-4" />
          Create Pool
        </Button>
      </div>

      <div className="grid gap-4">
        {pools.length === 0 ? (
          <div className="text-center py-12 bg-muted/50 rounded-lg">
            <HardDrive className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium mb-2">No storage pools</h3>
            <p className="text-muted-foreground mb-4">
              Get started by creating your first storage pool.
            </p>
            <Button onClick={() => navigate('/storage/create')}>
              <Plus className="mr-2 h-4 w-4" />
              Create Pool
            </Button>
          </div>
        ) : (
          pools.map((pool) => (
            <div
              key={pool.name}
              className="bg-card rounded-lg border p-6 hover:shadow-lg transition-shadow"
            >
              <div className="flex items-start justify-between mb-4">
                <div>
                  <Link
                    to={`/storage/pools/${pool.name}`}
                    className="text-xl font-semibold hover:underline"
                  >
                    {pool.name}
                  </Link>
                  <div className="flex items-center gap-4 mt-2 text-sm text-muted-foreground">
                    <span>Type: {pool.type}</span>
                    {pool.path && <span>Path: {pool.path}</span>}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <span
                    className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-${getPoolStateColor(
                      pool.state
                    )}-500/20 text-${getPoolStateColor(pool.state)}-700`}
                  >
                    {pool.state}
                  </span>
                  {pool.autostart && (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-500/20 text-blue-700">
                      Autostart
                    </span>
                  )}
                </div>
              </div>

              <div className="space-y-3">
                <div>
                  <div className="flex items-center justify-between text-sm mb-1">
                    <span className="text-muted-foreground">Storage Usage</span>
                    <span>
                      {formatBytes(pool.allocation)} / {formatBytes(pool.capacity)} (
                      {getUsagePercentage(pool)}%)
                    </span>
                  </div>
                  <div className="w-full bg-muted rounded-full h-2">
                    <div
                      className="bg-primary rounded-full h-2 transition-all"
                      style={{ width: `${getUsagePercentage(pool)}%` }}
                    />
                  </div>
                  <div className="text-xs text-muted-foreground mt-1">
                    {formatBytes(pool.available)} available
                  </div>
                </div>

                <div className="flex items-center gap-2 pt-2">
                  {pool.state === 'running' ? (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleStop(pool.name)}
                      disabled={actionLoading === pool.name}
                    >
                      <Square className="mr-2 h-4 w-4" />
                      Stop
                    </Button>
                  ) : (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleStart(pool.name)}
                      disabled={actionLoading === pool.name}
                    >
                      <Play className="mr-2 h-4 w-4" />
                      Start
                    </Button>
                  )}
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => navigate(`/storage/pools/${pool.name}`)}
                  >
                    View Volumes
                  </Button>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() => handleDelete(pool.name)}
                    disabled={actionLoading === pool.name}
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    Delete
                  </Button>
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {pools.length > 0 && (
        <div className="flex justify-end">
          <Button variant="outline" onClick={loadPools}>
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
        </div>
      )}
    </div>
  );
}