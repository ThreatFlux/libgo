import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from '@tanstack/react-router';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ArrowLeft, Plus, Trash2, Network, Activity, Settings } from 'lucide-react';
import {
  getBridge,
  listPorts,
  deletePort,
  type OVSBridge,
  type OVSPort
} from '@/api/ovs';

const OVSBridgeDetailPage: React.FC = () => {
  const { name } = useParams({ from: '/ovs/bridges/$name' });
  const navigate = useNavigate();
  const [bridge, setBridge] = useState<OVSBridge | null>(null);
  const [ports, setPorts] = useState<OVSPort[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadBridgeData = async () => {
    if (!name) return;

    try {
      setLoading(true);
      setError(null);
      
      const [bridgeResponse, portsResponse] = await Promise.all([
        getBridge(name),
        listPorts(name)
      ]);
      
      setBridge(bridgeResponse.bridge);
      setPorts(portsResponse.ports);
    } catch (err) {
      console.error('Failed to load bridge data:', err);
      setError('Failed to load bridge data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadBridgeData();
  }, [name]);

  const handleDeletePort = async (portName: string) => {
    if (!name || !confirm(`Are you sure you want to delete port "${portName}"?`)) {
      return;
    }

    try {
      await deletePort(name, portName);
      await loadBridgeData();
    } catch (err) {
      console.error('Failed to delete port:', err);
      alert('Failed to delete port');
    }
  };

  const formatPortType = (type: string) => {
    switch (type) {
      case 'internal':
        return <Badge variant="default">Internal</Badge>;
      case 'patch':
        return <Badge variant="secondary">Patch</Badge>;
      case 'vxlan':
        return <Badge variant="outline">VXLAN</Badge>;
      case 'gre':
        return <Badge variant="outline">GRE</Badge>;
      case 'geneve':
        return <Badge variant="outline">Geneve</Badge>;
      default:
        return <Badge variant="outline">{type || 'System'}</Badge>;
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  if (loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="flex items-center justify-center h-64">
          <div className="text-lg">Loading bridge details...</div>
        </div>
      </div>
    );
  }

  if (error || !bridge) {
    return (
      <div className="container mx-auto p-6">
        <Card>
          <CardContent className="pt-6">
            <div className="text-center text-red-600">
              <p>{error || 'Bridge not found'}</p>
              <Button onClick={loadBridgeData} className="mt-4">
                Retry
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
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
            <h1 className="text-3xl font-bold">{bridge.name}</h1>
            <p className="text-muted-foreground">OVS Bridge Details</p>
          </div>
        </div>
        <Button onClick={() => navigate({ to: '/ovs/bridges/$bridgeName/ports/create', params: { bridgeName: name } })}>
          <Plus className="w-4 h-4 mr-2" />
          Add Port
        </Button>
      </div>

      {/* Bridge Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">UUID</CardTitle>
            <Settings className="w-4 h-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-sm font-mono">{bridge.uuid}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Datapath Type</CardTitle>
            <Network className="w-4 h-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-sm">{bridge.datapath_type}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Ports</CardTitle>
            <Activity className="w-4 h-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{ports.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Flow Count</CardTitle>
            <Activity className="w-4 h-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {bridge.statistics?.flow_count || 0}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="ports" className="space-y-4">
        <TabsList>
          <TabsTrigger value="ports">Ports</TabsTrigger>
          <TabsTrigger value="flows">Flows</TabsTrigger>
          <TabsTrigger value="config">Configuration</TabsTrigger>
        </TabsList>

        <TabsContent value="ports">
          <Card>
            <CardHeader>
              <CardTitle>Ports</CardTitle>
            </CardHeader>
            <CardContent>
              {ports.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <p>No ports configured</p>
                  <p className="text-sm mt-2">Add a port to get started</p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>VLAN Tag</TableHead>
                      <TableHead>Trunk VLANs</TableHead>
                      <TableHead>RX/TX Bytes</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {ports.map((port) => (
                      <TableRow key={port.name}>
                        <TableCell className="font-medium">{port.name}</TableCell>
                        <TableCell>{formatPortType(port.type)}</TableCell>
                        <TableCell>
                          {port.tag ? (
                            <Badge variant="outline">VLAN {port.tag}</Badge>
                          ) : (
                            <span className="text-muted-foreground">None</span>
                          )}
                        </TableCell>
                        <TableCell>
                          {port.trunks && port.trunks.length > 0 ? (
                            <div className="flex gap-1 flex-wrap">
                              {port.trunks.map(vlan => (
                                <Badge key={vlan} variant="outline" className="text-xs">
                                  {vlan}
                                </Badge>
                              ))}
                            </div>
                          ) : (
                            <span className="text-muted-foreground">None</span>
                          )}
                        </TableCell>
                        <TableCell>
                          {port.statistics ? (
                            <div className="text-sm">
                              <div>RX: {formatBytes(port.statistics.rx_bytes)}</div>
                              <div>TX: {formatBytes(port.statistics.tx_bytes)}</div>
                            </div>
                          ) : (
                            <span className="text-muted-foreground">-</span>
                          )}
                        </TableCell>
                        <TableCell className="text-right">
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => handleDeletePort(port.name)}
                          >
                            <Trash2 className="w-4 h-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="flows">
          <Card>
            <CardHeader>
              <CardTitle>Flow Rules</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-center py-8 text-muted-foreground">
                <p>Flow management coming soon</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="config">
          <Card>
            <CardHeader>
              <CardTitle>Configuration</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <h4 className="font-medium mb-2">Controller</h4>
                <p className="text-sm text-muted-foreground">
                  {bridge.controller || 'No controller configured'}
                </p>
              </div>
              
              {bridge.external_ids && Object.keys(bridge.external_ids).length > 0 && (
                <div>
                  <h4 className="font-medium mb-2">External IDs</h4>
                  <div className="space-y-1">
                    {Object.entries(bridge.external_ids).map(([key, value]) => (
                      <div key={key} className="text-sm">
                        <span className="font-mono text-muted-foreground">{key}:</span> {value}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {bridge.other_config && Object.keys(bridge.other_config).length > 0 && (
                <div>
                  <h4 className="font-medium mb-2">Other Config</h4>
                  <div className="space-y-1">
                    {Object.entries(bridge.other_config).map(([key, value]) => (
                      <div key={key} className="text-sm">
                        <span className="font-mono text-muted-foreground">{key}:</span> {value}
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default OVSBridgeDetailPage;