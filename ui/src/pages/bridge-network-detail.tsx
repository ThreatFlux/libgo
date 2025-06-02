import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from '@tanstack/react-router';
import { Button } from '../components/ui/button';
import { getBridgeNetwork, deleteBridgeNetwork, BridgeNetwork } from '../api/bridge-network';

export function BridgeNetworkDetail() {
  const { name } = useParams({ from: '/bridge-networks/$name' });
  const navigate = useNavigate();
  const [network, setNetwork] = useState<BridgeNetwork | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadNetwork = async () => {
    try {
      setLoading(true);
      const networkData = await getBridgeNetwork(name);
      setNetwork(networkData);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load bridge network');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!network || !confirm(`Are you sure you want to delete bridge network "${network.name}"?`)) {
      return;
    }

    try {
      await deleteBridgeNetwork(network.name);
      navigate({ to: '/bridge-networks' });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete bridge network');
    }
  };

  useEffect(() => {
    loadNetwork();
  }, [name]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-lg">Loading bridge network...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4">
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          Error: {error}
        </div>
        <Button onClick={loadNetwork}>Retry</Button>
      </div>
    );
  }

  if (!network) {
    return (
      <div className="p-4">
        <div className="text-center py-8">
          <p className="text-gray-500">Bridge network not found</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Bridge Network: {network.name}</h1>
        <div className="space-x-2">
          <Button
            onClick={() => navigate({ to: '/bridge-networks' })}
            className="bg-gray-500 hover:bg-gray-600"
          >
            Back to List
          </Button>
          <Button
            onClick={handleDelete}
            className="bg-red-500 hover:bg-red-600"
          >
            Delete Network
          </Button>
        </div>
      </div>

      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-4">Network Information</h2>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <h3 className="text-sm font-medium text-gray-500 uppercase tracking-wider mb-2">
              Basic Information
            </h3>
            <dl className="space-y-3">
              <div>
                <dt className="text-sm font-medium text-gray-700">Name</dt>
                <dd className="text-sm text-gray-900">{network.name}</dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-700">Bridge Name</dt>
                <dd className="text-sm text-gray-900">{network.bridge_name}</dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-700">Forward Mode</dt>
                <dd className="text-sm text-gray-900">{network.forward_mode}</dd>
              </div>
            </dl>
          </div>

          <div>
            <h3 className="text-sm font-medium text-gray-500 uppercase tracking-wider mb-2">
              Status & Configuration
            </h3>
            <dl className="space-y-3">
              <div>
                <dt className="text-sm font-medium text-gray-700">Status</dt>
                <dd>
                  <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                    network.active 
                      ? 'bg-green-100 text-green-800' 
                      : 'bg-red-100 text-red-800'
                  }`}>
                    {network.active ? 'Active' : 'Inactive'}
                  </span>
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-700">Auto Start</dt>
                <dd className="text-sm text-gray-900">
                  {network.auto_start ? 'Enabled' : 'Disabled'}
                </dd>
              </div>
            </dl>
          </div>
        </div>

        <div className="mt-8 p-4 bg-blue-50 rounded-md">
          <h3 className="text-sm font-medium text-blue-800 mb-2">Bridge Network Information:</h3>
          <div className="text-xs text-blue-700 space-y-1">
            <p>• This network connects VMs directly to the Linux bridge "{network.bridge_name}"</p>
            <p>• VMs using this network will appear on the same network segment as the bridge</p>
            <p>• The bridge must exist on the host system for this network to function</p>
            {network.active && <p>• Network is currently active and available for VM connections</p>}
            {!network.active && <p>• Network is inactive - VMs cannot currently connect to this network</p>}
          </div>
        </div>
      </div>
    </div>
  );
}