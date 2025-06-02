import React, { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { Button } from '../components/ui/button';
import { createBridgeNetwork, CreateBridgeNetworkRequest } from '../api/bridge-network';

export function BridgeNetworkCreate() {
  const navigate = useNavigate();
  const [formData, setFormData] = useState<CreateBridgeNetworkRequest>({
    name: '',
    bridge_name: '',
    auto_start: false,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.name.trim() || !formData.bridge_name.trim()) {
      setError('Network name and bridge name are required');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      await createBridgeNetwork(formData);
      navigate({ to: '/bridge-networks' });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create bridge network');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-6 max-w-md mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-2">Create Bridge Network</h1>
        <p className="text-gray-600">
          Create a new libvirt bridge network that connects to an existing Linux bridge.
        </p>
      </div>

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
            Network Name
          </label>
          <input
            type="text"
            id="name"
            name="name"
            value={formData.name}
            onChange={handleInputChange}
            placeholder="e.g., bridge-network"
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
          <p className="text-xs text-gray-500 mt-1">
            Unique name for the libvirt network
          </p>
        </div>

        <div>
          <label htmlFor="bridge_name" className="block text-sm font-medium text-gray-700 mb-1">
            Bridge Name
          </label>
          <input
            type="text"
            id="bridge_name"
            name="bridge_name"
            value={formData.bridge_name}
            onChange={handleInputChange}
            placeholder="e.g., br0"
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
          <p className="text-xs text-gray-500 mt-1">
            Name of the existing Linux bridge to connect to
          </p>
        </div>

        <div className="flex items-center">
          <input
            type="checkbox"
            id="auto_start"
            name="auto_start"
            checked={formData.auto_start}
            onChange={handleInputChange}
            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
          />
          <label htmlFor="auto_start" className="ml-2 block text-sm text-gray-700">
            Auto-start network
          </label>
        </div>

        <div className="flex space-x-4 pt-4">
          <Button
            type="submit"
            disabled={loading}
            className="flex-1"
          >
            {loading ? 'Creating...' : 'Create Network'}
          </Button>
          <Button
            type="button"
            onClick={() => navigate({ to: '/bridge-networks' })}
            className="flex-1 bg-gray-500 hover:bg-gray-600"
          >
            Cancel
          </Button>
        </div>
      </form>

      <div className="mt-8 p-4 bg-blue-50 rounded-md">
        <h3 className="text-sm font-medium text-blue-800 mb-2">Important Notes:</h3>
        <ul className="text-xs text-blue-700 space-y-1">
          <li>• The Linux bridge must already exist on the host system</li>
          <li>• VMs using this network will be directly connected to the bridge</li>
          <li>• This enables bridged networking for VMs on the same network as the host</li>
          <li>• Use commands like "brctl show" to list existing bridges</li>
        </ul>
      </div>
    </div>
  );
}