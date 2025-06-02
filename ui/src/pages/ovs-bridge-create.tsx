import React, { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ArrowLeft, Save } from 'lucide-react';
import { createBridge, type CreateBridgeRequest } from '@/api/ovs';

const OVSBridgeCreatePage: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<CreateBridgeRequest>({
    name: '',
    datapath_type: 'system',
    controller: '',
    external_ids: {},
    other_config: {},
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.name.trim()) {
      alert('Bridge name is required');
      return;
    }

    try {
      setLoading(true);
      
      // Clean up the form data
      const submitData: CreateBridgeRequest = {
        name: formData.name.trim(),
        datapath_type: formData.datapath_type,
      };

      // Add optional fields only if they have values
      if (formData.controller?.trim()) {
        submitData.controller = formData.controller.trim();
      }

      await createBridge(submitData);
      navigate({ to: '/ovs/bridges' });
    } catch (err) {
      console.error('Failed to create bridge:', err);
      alert('Failed to create bridge');
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (field: keyof CreateBridgeRequest, value: string) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
  };

  return (
    <div className="container mx-auto p-6 max-w-2xl">
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center gap-4">
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate({ to: '/ovs/bridges' })}
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Bridges
          </Button>
          <div>
            <h1 className="text-3xl font-bold">Create OVS Bridge</h1>
            <p className="text-muted-foreground">
              Create a new Open vSwitch bridge
            </p>
          </div>
        </div>

        {/* Form */}
        <Card>
          <CardHeader>
            <CardTitle>Bridge Configuration</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Bridge Name */}
              <div className="space-y-2">
                <Label htmlFor="name">Bridge Name *</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => handleInputChange('name', e.target.value)}
                  placeholder="br-ovs"
                  required
                  disabled={loading}
                />
                <p className="text-sm text-muted-foreground">
                  A unique name for the OVS bridge. Commonly prefixed with 'br-'.
                </p>
              </div>

              {/* Datapath Type */}
              <div className="space-y-2">
                <Label htmlFor="datapath_type">Datapath Type</Label>
                <Select
                  value={formData.datapath_type}
                  onValueChange={(value) => handleInputChange('datapath_type', value)}
                  disabled={loading}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="system">System</SelectItem>
                    <SelectItem value="netdev">Netdev</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-sm text-muted-foreground">
                  System datapath uses kernel networking. Netdev datapath uses userspace DPDK.
                </p>
              </div>

              {/* Controller (Optional) */}
              <div className="space-y-2">
                <Label htmlFor="controller">SDN Controller (Optional)</Label>
                <Input
                  id="controller"
                  value={formData.controller}
                  onChange={(e) => handleInputChange('controller', e.target.value)}
                  placeholder="tcp:127.0.0.1:6633"
                  disabled={loading}
                />
                <p className="text-sm text-muted-foreground">
                  Optional SDN controller endpoint (e.g., tcp:IP:PORT, ssl:IP:PORT)
                </p>
              </div>

              {/* Action Buttons */}
              <div className="flex gap-3 pt-4">
                <Button
                  type="submit"
                  disabled={loading || !formData.name.trim()}
                  className="flex items-center gap-2"
                >
                  <Save className="w-4 h-4" />
                  {loading ? 'Creating...' : 'Create Bridge'}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => navigate({ to: '/ovs/bridges' })}
                  disabled={loading}
                >
                  Cancel
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        {/* Tips */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Tips</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm text-muted-foreground">
            <p>• Bridge names should be unique and descriptive</p>
            <p>• System datapath is recommended for most use cases</p>
            <p>• Controllers are only needed for SDN deployments</p>
            <p>• You can add ports and configure flows after creating the bridge</p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default OVSBridgeCreatePage;