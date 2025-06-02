import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { networkAPI, NetworkInfo } from '../api/network';
import { Button } from '../components/ui/button';

export function NetworkDetail() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const [network, setNetwork] = useState<NetworkInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (name) {
      loadNetwork();
    }
  }, [name]);

  const loadNetwork = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await networkAPI.get(name!);
      setNetwork(data);
    } catch (err) {
      setError('Failed to load network details');
      console.error('Error loading network:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleStart = async () => {
    try {
      await networkAPI.start(name!);
      await loadNetwork(); // Reload to get updated status
    } catch (err) {
      console.error('Error starting network:', err);
    }
  };

  const handleStop = async () => {
    try {
      await networkAPI.stop(name!);
      await loadNetwork(); // Reload to get updated status
    } catch (err) {
      console.error('Error stopping network:', err);
    }
  };

  const handleDelete = async () => {
    if (!confirm(`Are you sure you want to delete network "${name}"?`)) {
      return;
    }

    try {
      await networkAPI.delete(name!);
      navigate('/networks');
    } catch (err) {
      console.error('Error deleting network:', err);
      alert('Failed to delete network. It may be in use.');
    }
  };

  const handleToggleAutostart = async () => {
    if (!network) return;

    try {
      await networkAPI.update(name!, {
        autostart: !network.autostart,
      });
      await loadNetwork(); // Reload to get updated status
    } catch (err) {
      console.error('Error updating network:', err);
    }
  };

  if (loading) {
    return <div className="p-6">Loading network details...</div>;
  }

  if (error || !network) {
    return <div className="p-6 text-red-600">Error: {error || 'Network not found'}</div>;
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <Button
          variant="outline"
          onClick={() => navigate('/networks')}
          className="mb-4"
        >
          ← Back to Networks
        </Button>

        <div className="flex justify-between items-start">
          <div>
            <h1 className="text-2xl font-bold">{network.name}</h1>
            <div className="mt-2 flex items-center space-x-2">
              <span className={`px-2 py-1 text-xs rounded-full ${
                network.active 
                  ? 'bg-green-100 text-green-800' 
                  : 'bg-gray-100 text-gray-800'
              }`}>
                {network.active ? 'Active' : 'Inactive'}
              </span>
              {network.persistent && (
                <span className="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded-full">
                  Persistent
                </span>
              )}
              {network.autostart && (
                <span className="px-2 py-1 text-xs bg-purple-100 text-purple-800 rounded-full">
                  Autostart
                </span>
              )}
            </div>
          </div>

          <div className="flex items-center space-x-2">
            {network.active ? (
              <Button variant="outline" onClick={handleStop}>
                Stop Network
              </Button>
            ) : (
              <Button variant="outline" onClick={handleStart}>
                Start Network
              </Button>
            )}
            <Button
              variant="outline"
              onClick={handleToggleAutostart}
            >
              {network.autostart ? 'Disable' : 'Enable'} Autostart
            </Button>
            <Button
              variant="outline"
              className="text-red-600 hover:text-red-700"
              onClick={handleDelete}
            >
              Delete Network
            </Button>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* General Information */}
        <div className="bg-white shadow overflow-hidden sm:rounded-lg">
          <div className="px-4 py-5 sm:px-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900">
              General Information
            </h3>
          </div>
          <div className="border-t border-gray-200">
            <dl>
              <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="text-sm font-medium text-gray-500">UUID</dt>
                <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2 font-mono">
                  {network.uuid}
                </dd>
              </div>
              <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="text-sm font-medium text-gray-500">Bridge Name</dt>
                <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                  {network.bridge_name}
                </dd>
              </div>
              <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="text-sm font-medium text-gray-500">Forward Mode</dt>
                <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                  {network.forward.mode}
                  {network.forward.dev && ` (${network.forward.dev})`}
                </dd>
              </div>
            </dl>
          </div>
        </div>

        {/* IP Configuration */}
        {network.ip && (
          <div className="bg-white shadow overflow-hidden sm:rounded-lg">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                IP Configuration
              </h3>
            </div>
            <div className="border-t border-gray-200">
              <dl>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Network</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {network.ip.address}/{network.ip.netmask}
                  </dd>
                </div>
                {network.ip.dhcp && (
                  <>
                    <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                      <dt className="text-sm font-medium text-gray-500">DHCP</dt>
                      <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                        {network.ip.dhcp.enabled ? 'Enabled' : 'Disabled'}
                      </dd>
                    </div>
                    {network.ip.dhcp.enabled && network.ip.dhcp.start && (
                      <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                        <dt className="text-sm font-medium text-gray-500">DHCP Range</dt>
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                          {network.ip.dhcp.start} - {network.ip.dhcp.end}
                        </dd>
                      </div>
                    )}
                  </>
                )}
              </dl>
            </div>
          </div>
        )}

        {/* DHCP Leases */}
        {network.dhcp_leases && network.dhcp_leases.length > 0 && (
          <div className="bg-white shadow overflow-hidden sm:rounded-lg md:col-span-2">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                Active DHCP Leases
              </h3>
            </div>
            <div className="border-t border-gray-200">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      IP Address
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      MAC Address
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Hostname
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Expiry
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {network.dhcp_leases.map((lease, index) => (
                    <tr key={index}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">
                        {lease.ip_address}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">
                        {lease.mac_address}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {lease.hostname || '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {new Date(lease.expiry_time * 1000).toLocaleString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* Static DHCP Hosts */}
        {network.ip?.dhcp?.hosts && network.ip.dhcp.hosts.length > 0 && (
          <div className="bg-white shadow overflow-hidden sm:rounded-lg md:col-span-2">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                Static DHCP Hosts
              </h3>
            </div>
            <div className="border-t border-gray-200">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      MAC Address
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      IP Address
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Hostname
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {network.ip.dhcp.hosts.map((host, index) => (
                    <tr key={index}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">
                        {host.mac}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">
                        {host.ip}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {host.name || '-'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}