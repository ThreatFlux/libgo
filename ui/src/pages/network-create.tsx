import React, { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { networkAPI, CreateNetworkParams } from '../api/network';
import { Button } from '../components/ui/button';

export function NetworkCreate() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const [formData, setFormData] = useState<CreateNetworkParams>({
    name: '',
    bridge_name: '',
    forward: {
      mode: 'nat',
    },
    ip: {
      address: '192.168.100.1',
      netmask: '255.255.255.0',
      dhcp: {
        enabled: true,
        start: '192.168.100.100',
        end: '192.168.100.200',
      },
    },
    autostart: true,
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      // If bridge name is empty, generate one
      const params = {
        ...formData,
        bridge_name: formData.bridge_name || `virbr-${formData.name}`,
      };

      await networkAPI.create(params);
      navigate({ to: '/networks' });
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create network');
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    
    if (name.includes('.')) {
      // Handle nested fields
      const [parent, child] = name.split('.');
      setFormData(prev => ({
        ...prev,
        [parent]: {
          ...(prev as any)[parent],
          [child]: type === 'checkbox' ? (e.target as HTMLInputElement).checked : value,
        },
      }));
    } else {
      setFormData(prev => ({
        ...prev,
        [name]: type === 'checkbox' ? (e.target as HTMLInputElement).checked : value,
      }));
    }
  };

  const handleDHCPToggle = (enabled: boolean) => {
    setFormData(prev => ({
      ...prev,
      ip: {
        ...prev.ip!,
        dhcp: {
          ...prev.ip!.dhcp!,
          enabled,
        },
      },
    }));
  };

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">Create Network</h1>

      <form onSubmit={handleSubmit} className="space-y-6 bg-white shadow px-6 py-6 sm:rounded-lg">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
            {error}
          </div>
        )}

        <div>
          <label htmlFor="name" className="block text-sm font-medium text-gray-700">
            Network Name
          </label>
          <input
            type="text"
            name="name"
            id="name"
            required
            value={formData.name}
            onChange={handleChange}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
            placeholder="my-network"
          />
        </div>

        <div>
          <label htmlFor="bridge_name" className="block text-sm font-medium text-gray-700">
            Bridge Name (optional)
          </label>
          <input
            type="text"
            name="bridge_name"
            id="bridge_name"
            value={formData.bridge_name}
            onChange={handleChange}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
            placeholder="virbr-my-network"
          />
          <p className="mt-1 text-sm text-gray-500">
            Leave empty to auto-generate based on network name
          </p>
        </div>

        <div>
          <label htmlFor="forward.mode" className="block text-sm font-medium text-gray-700">
            Forward Mode
          </label>
          <select
            name="forward.mode"
            id="forward.mode"
            value={formData.forward?.mode}
            onChange={handleChange}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
          >
            <option value="nat">NAT</option>
            <option value="route">Route</option>
            <option value="bridge">Bridge</option>
            <option value="private">Private</option>
          </select>
        </div>

        <fieldset>
          <legend className="text-sm font-medium text-gray-700">IP Configuration</legend>
          <div className="mt-2 space-y-4">
            <div>
              <label htmlFor="ip.address" className="block text-sm font-medium text-gray-700">
                IP Address
              </label>
              <input
                type="text"
                name="ip.address"
                id="ip.address"
                value={formData.ip?.address}
                onChange={handleChange}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                placeholder="192.168.100.1"
              />
            </div>

            <div>
              <label htmlFor="ip.netmask" className="block text-sm font-medium text-gray-700">
                Netmask
              </label>
              <input
                type="text"
                name="ip.netmask"
                id="ip.netmask"
                value={formData.ip?.netmask}
                onChange={handleChange}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                placeholder="255.255.255.0"
              />
            </div>
          </div>
        </fieldset>

        <fieldset>
          <legend className="text-sm font-medium text-gray-700">DHCP Configuration</legend>
          <div className="mt-2 space-y-4">
            <div className="flex items-center">
              <input
                type="checkbox"
                id="dhcp-enabled"
                checked={formData.ip?.dhcp?.enabled || false}
                onChange={(e) => handleDHCPToggle(e.target.checked)}
                className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
              />
              <label htmlFor="dhcp-enabled" className="ml-2 block text-sm text-gray-900">
                Enable DHCP
              </label>
            </div>

            {formData.ip?.dhcp?.enabled && (
              <>
                <div>
                  <label htmlFor="dhcp.start" className="block text-sm font-medium text-gray-700">
                    DHCP Start Range
                  </label>
                  <input
                    type="text"
                    name="ip.dhcp.start"
                    id="dhcp.start"
                    value={formData.ip?.dhcp?.start || ''}
                    onChange={handleChange}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                    placeholder="192.168.100.100"
                  />
                </div>

                <div>
                  <label htmlFor="dhcp.end" className="block text-sm font-medium text-gray-700">
                    DHCP End Range
                  </label>
                  <input
                    type="text"
                    name="ip.dhcp.end"
                    id="dhcp.end"
                    value={formData.ip?.dhcp?.end || ''}
                    onChange={handleChange}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                    placeholder="192.168.100.200"
                  />
                </div>
              </>
            )}
          </div>
        </fieldset>

        <div className="flex items-center">
          <input
            type="checkbox"
            name="autostart"
            id="autostart"
            checked={formData.autostart}
            onChange={handleChange}
            className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
          />
          <label htmlFor="autostart" className="ml-2 block text-sm text-gray-900">
            Start network automatically on boot
          </label>
        </div>

        <div className="flex justify-end space-x-3">
          <Button
            type="button"
            variant="outline"
            onClick={() => navigate({ to: '/networks' })}
            disabled={loading}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={loading}>
            {loading ? 'Creating...' : 'Create Network'}
          </Button>
        </div>
      </form>
    </div>
  );
}