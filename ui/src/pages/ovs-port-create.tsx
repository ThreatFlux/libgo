import React, { useState } from 'react';
import { useNavigate, useParams } from '@tanstack/react-router';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ArrowLeft, Save } from 'lucide-react';
import { createPort, type CreatePortRequest } from '@/api/ovs';

const OVSPortCreatePage: React.FC = () => {
  const navigate = useNavigate();
  const { bridgeName } = useParams({ from: '/ovs/bridges/$bridgeName/ports/create' });
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<CreatePortRequest>({
    name: '',
    bridge: bridgeName || '',
    type: 'system',
    tag: undefined,
    trunks: [],
    peer_port: '',
    remote_ip: '',
    tunnel_type: '',
    external_ids: {},
    other_config: {},
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.name.trim() || !formData.bridge.trim()) {
      alert('Port name and bridge are required');
      return;
    }

    try {
      setLoading(true);
      
      // Clean up the form data
      const submitData: CreatePortRequest = {
        name: formData.name.trim(),
        bridge: formData.bridge.trim(),
      };

      // Add optional fields only if they have values
      if (formData.type && formData.type !== 'system') {
        submitData.type = formData.type;
      }

      if (formData.tag !== undefined && formData.tag > 0) {
        submitData.tag = formData.tag;
      }

      if (formData.trunks && formData.trunks.length > 0) {
        submitData.trunks = formData.trunks;
      }

      if (formData.type === 'patch' && formData.peer_port?.trim()) {
        submitData.peer_port = formData.peer_port.trim();
      }

      if ((formData.type === 'vxlan' || formData.type === 'gre' || formData.type === 'geneve') && formData.remote_ip?.trim()) {
        submitData.remote_ip = formData.remote_ip.trim();
      }

      await createPort(submitData);
      navigate({ to: '/ovs/bridges/$name', params: { name: formData.bridge } });
    } catch (err) {
      console.error('Failed to create port:', err);
      alert('Failed to create port');
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (field: keyof CreatePortRequest, value: string | number | number[] | undefined) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleTrunksChange = (value: string) => {
    if (!value.trim()) {
      handleInputChange('trunks', []);
      return;
    }

    try {
      const trunks = value.split(',').map(v => parseInt(v.trim())).filter(v => !isNaN(v) && v > 0);
      handleInputChange('trunks', trunks);
    } catch (err) {
      // Invalid input, ignore
    }
  };

  const isVLANAccessPort = formData.tag !== undefined && formData.tag > 0;
  const isTunnelPort = ['vxlan', 'gre', 'geneve'].includes(formData.type || '');
  const isPatchPort = formData.type === 'patch';

  return (
    <div className="container mx-auto p-6 max-w-2xl">
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center gap-4">
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate({ to: '/ovs/bridges/$name', params: { name: bridgeName } })}
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Bridge
          </Button>
          <div>
            <h1 className="text-3xl font-bold">Create Port</h1>
            <p className="text-muted-foreground">
              Add a new port to bridge {bridgeName}
            </p>
          </div>
        </div>

        {/* Form */}
        <Card>
          <CardHeader>
            <CardTitle>Port Configuration</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Port Name */}
              <div className="space-y-2">
                <Label htmlFor="name">Port Name *</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => handleInputChange('name', e.target.value)}
                  placeholder="eth0"
                  required
                  disabled={loading}
                />
                <p className="text-sm text-muted-foreground">
                  Name of the network interface or port
                </p>
              </div>

              {/* Port Type */}
              <div className="space-y-2">
                <Label htmlFor="type">Port Type</Label>
                <Select
                  value={formData.type}
                  onValueChange={(value) => handleInputChange('type', value)}
                  disabled={loading}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="system">System (Physical/Virtual Interface)</SelectItem>
                    <SelectItem value="internal">Internal Port</SelectItem>
                    <SelectItem value="patch">Patch Port</SelectItem>
                    <SelectItem value="vxlan">VXLAN Tunnel</SelectItem>
                    <SelectItem value="gre">GRE Tunnel</SelectItem>
                    <SelectItem value="geneve">Geneve Tunnel</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-sm text-muted-foreground">
                  Type of port to create. System ports connect to existing interfaces.
                </p>
              </div>

              {/* VLAN Configuration */}
              <div className="space-y-4">
                <Label>VLAN Configuration</Label>
                
                {/* VLAN Tag */}
                <div className="space-y-2">
                  <Label htmlFor="tag">Access VLAN Tag</Label>
                  <Input
                    id="tag"
                    type="number"
                    min="1"
                    max="4094"
                    value={formData.tag || ''}
                    onChange={(e) => {
                      const value = e.target.value;
                      handleInputChange('tag', value ? parseInt(value) : undefined);
                    }}
                    placeholder="Optional (1-4094)"
                    disabled={loading}
                  />
                  <p className="text-sm text-muted-foreground">
                    VLAN tag for access port (mutually exclusive with trunk VLANs)
                  </p>
                </div>

                {/* Trunk VLANs */}
                {!isVLANAccessPort && (
                  <div className="space-y-2">
                    <Label htmlFor="trunks">Trunk VLANs</Label>
                    <Input
                      id="trunks"
                      value={formData.trunks?.join(', ') || ''}
                      onChange={(e) => handleTrunksChange(e.target.value)}
                      placeholder="1,2,3,100 (comma-separated)"
                      disabled={loading}
                    />
                    <p className="text-sm text-muted-foreground">
                      Comma-separated list of VLAN IDs for trunk port
                    </p>
                  </div>
                )}
              </div>

              {/* Patch Port Configuration */}
              {isPatchPort && (
                <div className="space-y-2">
                  <Label htmlFor="peer_port">Peer Port</Label>
                  <Input
                    id="peer_port"
                    value={formData.peer_port}
                    onChange={(e) => handleInputChange('peer_port', e.target.value)}
                    placeholder="patch-br1-to-br2"
                    disabled={loading}
                  />
                  <p className="text-sm text-muted-foreground">
                    Name of the peer patch port on the other bridge
                  </p>
                </div>
              )}

              {/* Tunnel Configuration */}
              {isTunnelPort && (
                <div className="space-y-2">
                  <Label htmlFor="remote_ip">Remote IP Address</Label>
                  <Input
                    id="remote_ip"
                    value={formData.remote_ip}
                    onChange={(e) => handleInputChange('remote_ip', e.target.value)}
                    placeholder="192.168.1.100"
                    disabled={loading}
                  />
                  <p className="text-sm text-muted-foreground">
                    IP address of the remote tunnel endpoint
                  </p>
                </div>
              )}

              {/* Action Buttons */}
              <div className="flex gap-3 pt-4">
                <Button
                  type="submit"
                  disabled={loading || !formData.name.trim()}
                  className="flex items-center gap-2"
                >
                  <Save className="w-4 h-4" />
                  {loading ? 'Creating...' : 'Create Port'}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => navigate({ to: '/ovs/bridges/$name', params: { name: bridgeName } })}
                  disabled={loading}
                >
                  Cancel
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        {/* Port Type Information */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Port Types</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm text-muted-foreground">
            <p><strong>System:</strong> Connects to existing network interfaces (eth0, tap devices, etc.)</p>
            <p><strong>Internal:</strong> Creates a virtual interface accessible from the host</p>
            <p><strong>Patch:</strong> Connects two OVS bridges together</p>
            <p><strong>VXLAN/GRE/Geneve:</strong> Creates overlay tunnel connections to remote hosts</p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default OVSPortCreatePage;