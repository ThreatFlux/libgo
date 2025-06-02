import React, { useState, useEffect } from 'react';
import { Link } from '@tanstack/react-router';
import { Button } from '../components/ui/button';
import { listBridgeNetworks, deleteBridgeNetwork, BridgeNetwork } from '../api/bridge-network';

export function BridgeNetworkList() {
  const [networks, setNetworks] = useState<BridgeNetwork[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadNetworks = async () => {
    try {
      setLoading(true);
      const response = await listBridgeNetworks();
      setNetworks(response.networks);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load bridge networks');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteNetwork = async (name: string) => {
    if (!confirm(`Are you sure you want to delete bridge network "${name}"?`)) {
      return;
    }

    try {
      await deleteBridgeNetwork(name);
      setNetworks(networks.filter(n => n.name !== name));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete bridge network');
    }
  };

  useEffect(() => {
    loadNetworks();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-lg">Loading bridge networks...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4">
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          Error: {error}
        </div>
        <Button onClick={loadNetworks}>Retry</Button>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Bridge Networks</h1>
        <Link to="/bridge-networks/create">
          <Button>Create Bridge Network</Button>
        </Link>
      </div>

      {networks.length === 0 ? (
        <div className="text-center py-8">
          <p className="text-gray-500 mb-4">No bridge networks found</p>
          <Link to="/bridge-networks/create">
            <Button>Create Your First Bridge Network</Button>
          </Link>
        </div>
      ) : (
        <div className="bg-white shadow rounded-lg overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Bridge Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Auto Start
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {networks.map((network) => (
                <tr key={network.name} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link 
                      to="/bridge-networks/$name" 
                      params={{ name: network.name }}
                      className="text-blue-600 hover:text-blue-800 font-medium"
                    >
                      {network.name}
                    </Link>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {network.bridge_name}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                      network.active 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-red-100 text-red-800'
                    }`}>
                      {network.active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {network.auto_start ? 'Yes' : 'No'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
                    <Link 
                      to="/bridge-networks/$name" 
                      params={{ name: network.name }}
                      className="text-blue-600 hover:text-blue-800"
                    >
                      View
                    </Link>
                    <button
                      onClick={() => handleDeleteNetwork(network.name)}
                      className="text-red-600 hover:text-red-800"
                    >
                      Delete
                    </button>
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