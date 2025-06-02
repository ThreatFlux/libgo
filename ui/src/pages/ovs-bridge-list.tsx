import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
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
import { Plus, Trash2, Eye, Activity } from 'lucide-react';
import { listBridges, deleteBridge, type OVSBridge } from '@/api/ovs';

const OVSBridgeListPage: React.FC = () => {
  const [bridges, setBridges] = useState<OVSBridge[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const loadBridges = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await listBridges();
      setBridges(response.bridges);
    } catch (err) {
      console.error('Failed to load OVS bridges:', err);
      setError('Failed to load OVS bridges');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadBridges();
  }, []);

  const handleDelete = async (bridgeName: string) => {
    if (!confirm(`Are you sure you want to delete bridge "${bridgeName}"?`)) {
      return;
    }

    try {
      await deleteBridge(bridgeName);
      await loadBridges();
    } catch (err) {
      console.error('Failed to delete bridge:', err);
      alert('Failed to delete bridge');
    }
  };

  const formatDatapathType = (type: string) => {
    switch (type) {
      case 'system':
        return <Badge variant="default">System</Badge>;
      case 'netdev':
        return <Badge variant="secondary">Netdev</Badge>;
      default:
        return <Badge variant="outline">{type}</Badge>;
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="flex items-center justify-center h-64">
          <div className="text-lg">Loading OVS bridges...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto p-6">
        <Card>
          <CardContent className="pt-6">
            <div className="text-center text-red-600">
              <p>{error}</p>
              <Button onClick={loadBridges} className="mt-4">
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
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">OVS Bridges</h1>
          <p className="text-muted-foreground mt-2">
            Manage Open vSwitch bridges and their configuration
          </p>
        </div>
        <Button onClick={() => navigate({ to: '/ovs/bridges/create' })} className="flex items-center gap-2">
          <Plus className="w-4 h-4" />
          Create Bridge
        </Button>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Bridges</CardTitle>
            <Activity className="w-4 h-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{bridges.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Ports</CardTitle>
            <Activity className="w-4 h-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {bridges.reduce((sum, bridge) => sum + bridge.ports.length, 0)}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">System Bridges</CardTitle>
            <Activity className="w-4 h-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {bridges.filter(b => b.datapath_type === 'system').length}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">With Controllers</CardTitle>
            <Activity className="w-4 h-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {bridges.filter(b => b.controller).length}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Bridges Table */}
      <Card>
        <CardHeader>
          <CardTitle>Bridges</CardTitle>
        </CardHeader>
        <CardContent>
          {bridges.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <p>No OVS bridges found</p>
              <p className="text-sm mt-2">Create your first bridge to get started</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>UUID</TableHead>
                  <TableHead>Datapath Type</TableHead>
                  <TableHead>Ports</TableHead>
                  <TableHead>Controller</TableHead>
                  <TableHead>Flows</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {bridges.map((bridge) => (
                  <TableRow key={bridge.name}>
                    <TableCell className="font-medium">
                      <Link
                        to="/ovs/bridges/$name"
                        params={{ name: bridge.name }}
                        className="text-blue-600 hover:text-blue-800 hover:underline"
                      >
                        {bridge.name}
                      </Link>
                    </TableCell>
                    <TableCell className="font-mono text-sm text-muted-foreground">
                      {bridge.uuid.substring(0, 8)}...
                    </TableCell>
                    <TableCell>{formatDatapathType(bridge.datapath_type)}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{bridge.ports.length}</Badge>
                    </TableCell>
                    <TableCell>
                      {bridge.controller ? (
                        <Badge variant="default">{bridge.controller}</Badge>
                      ) : (
                        <span className="text-muted-foreground">None</span>
                      )}
                    </TableCell>
                    <TableCell>
                      {bridge.statistics ? (
                        <Badge variant="outline">{bridge.statistics.flow_count}</Badge>
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => navigate({ to: '/ovs/bridges/$name', params: { name: bridge.name } })}
                        >
                          <Eye className="w-4 h-4" />
                        </Button>
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={() => handleDelete(bridge.name)}
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default OVSBridgeListPage;