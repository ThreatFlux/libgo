import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { networkAPI, NetworkInfo } from '../api/network';
import { Button } from '../components/ui/button';

export function NetworkList() {
  const navigate = useNavigate();
  const [networks, setNetworks] = useState<NetworkInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadNetworks();
  }, []);

  const loadNetworks = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await networkAPI.list();
      setNetworks(response.networks);
    } catch (err) {
      setError('Failed to load networks');
      console.error('Error loading networks:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleStart = async (name: string) => {
    try {
      await networkAPI.start(name);
      await loadNetworks(); // Reload to get updated status
    } catch (err) {
      console.error('Error starting network:', err);
    }
  };

  const handleStop = async (name: string) => {
    try {
      await networkAPI.stop(name);
      await loadNetworks(); // Reload to get updated status
    } catch (err) {
      console.error('Error stopping network:', err);
    }
  };

  const handleDelete = async (name: string) => {
    if (!confirm(`Are you sure you want to delete network "${name}"?`)) {
      return;
    }

    try {
      await networkAPI.delete(name);
      await loadNetworks(); // Reload list
    } catch (err) {
      console.error('Error deleting network:', err);
      alert('Failed to delete network. It may be in use.');
    }
  };

  if (loading) {
    return <div className="p-6">Loading networks...</div>;
  }

  if (error) {
    return <div className="p-6 text-red-600">Error: {error}</div>;
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Networks</h1>
        <Button onClick={() => navigate('/networks/create')}>
          Create Network
        </Button>
      </div>

      <div className="bg-white shadow overflow-hidden sm:rounded-md">
        <ul className="divide-y divide-gray-200">
          {networks.length === 0 ? (
            <li className="px-6 py-4 text-gray-500">No networks found</li>
          ) : (
            networks.map((network) => (
              <li key={network.name} className="px-6 py-4">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <div className="flex items-center">
                      <h3 className="text-lg font-medium text-gray-900">
                        {network.name}
                      </h3>
                      <span className={`ml-2 px-2 py-1 text-xs rounded-full ${
                        network.active 
                          ? 'bg-green-100 text-green-800' 
                          : 'bg-gray-100 text-gray-800'
                      }`}>
                        {network.active ? 'Active' : 'Inactive'}
                      </span>
                      {network.autostart && (
                        <span className="ml-2 px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded-full">
                          Autostart
                        </span>
                      )}
                    </div>
                    <div className="mt-1 text-sm text-gray-500">
                      <div>Bridge: {network.bridge_name}</div>
                      <div>Forward Mode: {network.forward.mode}</div>
                      {network.ip && (
                        <div>
                          IP: {network.ip.address}/{network.ip.netmask}
                          {network.ip.dhcp?.enabled && ' (DHCP enabled)'}
                        </div>
                      )}
                      {network.dhcp_leases && network.dhcp_leases.length > 0 && (
                        <div>{network.dhcp_leases.length} active DHCP lease(s)</div>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => navigate(`/networks/${network.name}`)}
                    >
                      Details
                    </Button>
                    {network.active ? (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleStop(network.name)}
                      >
                        Stop
                      </Button>
                    ) : (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleStart(network.name)}
                      >
                        Start
                      </Button>
                    )}
                    <Button
                      variant="outline"
                      size="sm"
                      className="text-red-600 hover:text-red-700"
                      onClick={() => handleDelete(network.name)}
                    >
                      Delete
                    </Button>
                  </div>
                </div>
              </li>
            ))
          )}
        </ul>
      </div>
    </div>
  );
}