import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Modal, Form, message, Popconfirm } from 'antd';
import {
  DatabaseOutlined,
  ReloadOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import { api } from '@/api/client';
import type { SpecInfo } from '@/types';
import PlatformConfigEditor from '@/components/PlatformConfigEditor';

const SpecsPage = () => {
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [isCreateModalVisible, setIsCreateModalVisible] = useState(false);
  const [isEditModalVisible, setIsEditModalVisible] = useState(false);
  const [editingSpec, setEditingSpec] = useState<SpecInfo | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const queryClient = useQueryClient();

  const { data: specs, isLoading, refetch } = useQuery({
    queryKey: ['specs'],
    queryFn: async () => {
      const response = await api.specs.list();
      return response.data;
    },
  });

  const createMutation = useMutation({
    mutationFn: api.specs.create,
    onSuccess: () => {
      message.success('Spec created successfully');
      setIsCreateModalVisible(false);
      createForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['specs'] });
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to create spec');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ name, data }: { name: string; data: any }) => api.specs.update(name, data),
    onSuccess: () => {
      message.success('Spec updated successfully');
      setIsEditModalVisible(false);
      setEditingSpec(null);
      editForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['specs'] });
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to update spec');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: api.specs.delete,
    onSuccess: () => {
      message.success('Spec deleted successfully');
      queryClient.invalidateQueries({ queryKey: ['specs'] });
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to delete spec');
    },
  });

  const filteredSpecs = specs?.filter(
    (spec) => selectedCategory === 'all' || spec.category === selectedCategory
  );

  const categories = ['all', ...new Set(specs?.map((s) => s.category) || [])];

  const handleCreate = async () => {
    try {
      const values = await createForm.validateFields();
      createMutation.mutate({
        name: values.name,
        displayName: values.displayName,
        category: values.category,
        resourceType: values.resourceType,
        resources: {
          cpu: values.cpu,
          memory: values.memory,
          gpu: values.gpu,
          gpuType: values.gpuType,
          ephemeralStorage: values.ephemeralStorage,
          shmSize: values.shmSize,
        },
        platforms: values.platforms || { generic: { nodeSelector: {}, tolerations: [], labels: {}, annotations: {} } },
      });
    } catch (error) {
      console.error('Validation failed:', error);
    }
  };

  const handleEdit = (spec: SpecInfo) => {
    setEditingSpec(spec);
    editForm.setFieldsValue({
      displayName: spec.displayName,
      category: spec.category,
      resourceType: spec.resourceType || 'serverless',
      cpu: spec.resources.cpu,
      memory: spec.resources.memory,
      gpu: spec.resources.gpu,
      gpuType: spec.resources.gpuType,
      ephemeralStorage: spec.resources.ephemeralStorage,
      shmSize: spec.resources.shmSize,
      platforms: spec.platforms || {},
    });
    setIsEditModalVisible(true);
  };

  const handleUpdate = async () => {
    if (!editingSpec) return;
    try {
      const values = await editForm.validateFields();
      updateMutation.mutate({
        name: editingSpec.name,
        data: {
          displayName: values.displayName,
          category: values.category,
          resourceType: values.resourceType,
          resources: {
            cpu: values.cpu,
            memory: values.memory,
            gpu: values.gpu,
            gpuType: values.gpuType,
            ephemeralStorage: values.ephemeralStorage,
            shmSize: values.shmSize,
          },
          platforms: values.platforms,
        },
      });
    } catch (error) {
      console.error('Validation failed:', error);
    }
  };

  if (isLoading) {
    return <div className="loading"><div className="spinner"></div></div>;
  }

  return (
    <>
      {/* Header */}
      <div className="flex justify-between items-center mb-4">
        <div className="filters">
          {categories.map((cat) => (
            <button
              key={cat}
              className={`btn btn-sm ${selectedCategory === cat ? 'btn-blue' : 'btn-outline'}`}
              onClick={() => setSelectedCategory(cat)}
            >
              {cat.toUpperCase()}
            </button>
          ))}
        </div>
        <div className="flex gap-2">
          <button className="btn btn-outline" onClick={() => refetch()}>
            <ReloadOutlined /> Refresh
          </button>
          <button className="btn btn-blue" onClick={() => setIsCreateModalVisible(true)}>
            <PlusOutlined /> Create Spec
          </button>
        </div>
      </div>

      {/* Specs Grid */}
      <div className="specs-grid">
        {filteredSpecs?.map((spec) => (
          <div key={spec.name} className="spec-card" style={{ cursor: 'default' }}>
            <div className="flex justify-between items-center mb-2">
              <div className="spec-header">
                <DatabaseOutlined style={{ marginRight: 8, color: '#8b5cf6' }} />
                <span className="spec-vram">{spec.displayName}</span>
              </div>
              <div className="flex gap-2">
                <button className="btn btn-sm btn-outline btn-icon" onClick={() => handleEdit(spec)}>
                  <EditOutlined />
                </button>
                <Popconfirm
                  title="Delete this spec?"
                  onConfirm={() => deleteMutation.mutate(spec.name)}
                  okText="Yes"
                  cancelText="No"
                >
                  <button className="btn btn-sm btn-outline btn-icon" style={{ color: '#f56565' }}>
                    <DeleteOutlined />
                  </button>
                </Popconfirm>
              </div>
            </div>

            <div className="flex gap-2 mb-3">
              <span className={`tag ${spec.category === 'gpu' ? 'running' : 'success'}`}>
                {spec.category.toUpperCase()}
              </span>
              <span className={`tag ${spec.resourceType === 'fixed' ? 'pending' : ''}`} style={{ background: 'rgba(139,92,246,0.1)', color: '#8b5cf6' }}>
                {spec.resourceType === 'fixed' ? 'FIXED' : 'SERVERLESS'}
              </span>
            </div>

            <div style={{ fontSize: 13, color: '#6b7280' }}>
              <div className="mb-2">
                <strong>Name:</strong> <code style={{ background: '#f3f4f6', padding: '2px 6px', borderRadius: 4 }}>{spec.name}</code>
              </div>
              {spec.resources.gpu && (
                <div className="mb-1">
                  <strong>GPU:</strong> {spec.resources.gpu} × {spec.resources.gpuType}
                </div>
              )}
              <div className="mb-1">
                <strong>CPU:</strong> {spec.resources.cpu} cores
              </div>
              <div className="mb-1">
                <strong>Memory:</strong> {spec.resources.memory}
              </div>
              {spec.resources.ephemeralStorage && (
                <div className="mb-1">
                  <strong>Storage:</strong> {spec.resources.ephemeralStorage}GB
                </div>
              )}
            </div>

            {spec.platforms && Object.keys(spec.platforms).length > 0 && (
              <div style={{ marginTop: 12, paddingTop: 12, borderTop: '1px solid #f3f4f6', fontSize: 12, color: '#9ca3af' }}>
                Platforms: {Object.keys(spec.platforms).join(', ')}
              </div>
            )}
          </div>
        ))}

        {(!filteredSpecs || filteredSpecs.length === 0) && (
          <div className="empty-state" style={{ gridColumn: '1 / -1' }}>
            <DatabaseOutlined style={{ fontSize: 48, opacity: 0.3 }} />
            <p>No specs found</p>
          </div>
        )}
      </div>

      {/* Create Modal */}
      <Modal
        title="Create New Spec"
        open={isCreateModalVisible}
        onOk={handleCreate}
        onCancel={() => { setIsCreateModalVisible(false); createForm.resetFields(); }}
        okText="Create"
        confirmLoading={createMutation.isPending}
        width={700}
        footer={null}
      >
        <Form form={createForm} layout="vertical" style={{ marginTop: 16 }}>
          <div className="form-row">
            <div className="form-group">
              <label className="form-label"><span className="required">*</span> Spec Name</label>
              <Form.Item name="name" rules={[{ required: true, message: 'Required' }, { pattern: /^[a-z0-9-]+$/, message: 'Lowercase, numbers, hyphens only' }]} style={{ marginBottom: 0 }}>
                <input className="form-input" placeholder="e.g., h200-single" />
              </Form.Item>
            </div>
            <div className="form-group">
              <label className="form-label"><span className="required">*</span> Display Name</label>
              <Form.Item name="displayName" rules={[{ required: true, message: 'Required' }]} style={{ marginBottom: 0 }}>
                <input className="form-input" placeholder="e.g., H200 1 GPU" />
              </Form.Item>
            </div>
          </div>
          <div className="form-row">
            <div className="form-group">
              <label className="form-label"><span className="required">*</span> Category</label>
              <Form.Item name="category" rules={[{ required: true }]} style={{ marginBottom: 0 }}>
                <select className="form-select">
                  <option value="">Select category</option>
                  <option value="cpu">CPU</option>
                  <option value="gpu">GPU</option>
                </select>
              </Form.Item>
            </div>
            <div className="form-group">
              <label className="form-label">Resource Type</label>
              <Form.Item name="resourceType" initialValue="serverless" style={{ marginBottom: 0 }}>
                <select className="form-select">
                  <option value="serverless">Serverless</option>
                  <option value="fixed">Fixed</option>
                </select>
              </Form.Item>
            </div>
          </div>
          <div className="form-row">
            <div className="form-group">
              <label className="form-label">CPU Cores</label>
              <Form.Item name="cpu" style={{ marginBottom: 0 }}>
                <input className="form-input" placeholder="e.g., 4" />
              </Form.Item>
            </div>
            <div className="form-group">
              <label className="form-label"><span className="required">*</span> Memory</label>
              <Form.Item name="memory" rules={[{ required: true, message: 'Required' }]} style={{ marginBottom: 0 }}>
                <input className="form-input" placeholder="e.g., 32Gi" />
              </Form.Item>
            </div>
          </div>
          <div className="form-row">
            <div className="form-group">
              <label className="form-label">GPU Count</label>
              <Form.Item name="gpu" style={{ marginBottom: 0 }}>
                <input className="form-input" placeholder="e.g., 1" />
              </Form.Item>
            </div>
            <div className="form-group">
              <label className="form-label">GPU Type</label>
              <Form.Item name="gpuType" style={{ marginBottom: 0 }}>
                <input className="form-input" placeholder="e.g., nvidia.com/gpu" />
              </Form.Item>
            </div>
          </div>
          <div className="form-row">
            <div className="form-group">
              <label className="form-label">Ephemeral Storage (GB)</label>
              <Form.Item name="ephemeralStorage" style={{ marginBottom: 0 }}>
                <input className="form-input" placeholder="e.g., 100" />
              </Form.Item>
            </div>
            <div className="form-group">
              <label className="form-label">Shared Memory</label>
              <Form.Item name="shmSize" style={{ marginBottom: 0 }}>
                <input className="form-input" placeholder="e.g., 16Gi" />
              </Form.Item>
            </div>
          </div>
          <div className="form-group full" style={{ marginTop: 16 }}>
            <label className="form-label">Platform Configurations</label>
            <Form.Item name="platforms" style={{ marginBottom: 0 }}>
              <PlatformConfigEditor />
            </Form.Item>
          </div>
          <div className="modal-footer" style={{ margin: '24px -24px -24px', padding: '16px 24px' }}>
            <button type="button" className="btn btn-outline" onClick={() => { setIsCreateModalVisible(false); createForm.resetFields(); }}>
              Cancel
            </button>
            <button type="button" className="btn btn-blue" onClick={handleCreate} disabled={createMutation.isPending}>
              {createMutation.isPending ? 'Creating...' : 'Create'}
            </button>
          </div>
        </Form>
      </Modal>

      {/* Edit Modal */}
      <Modal
        title={`Edit Spec: ${editingSpec?.name}`}
        open={isEditModalVisible}
        onOk={handleUpdate}
        onCancel={() => { setIsEditModalVisible(false); setEditingSpec(null); editForm.resetFields(); }}
        okText="Update"
        confirmLoading={updateMutation.isPending}
        width={700}
        footer={null}
      >
        <Form form={editForm} layout="vertical" style={{ marginTop: 16 }}>
          <div className="form-row">
            <div className="form-group">
              <label className="form-label"><span className="required">*</span> Display Name</label>
              <Form.Item name="displayName" rules={[{ required: true, message: 'Required' }]} style={{ marginBottom: 0 }}>
                <input className="form-input" />
              </Form.Item>
            </div>
            <div className="form-group">
              <label className="form-label"><span className="required">*</span> Category</label>
              <Form.Item name="category" rules={[{ required: true }]} style={{ marginBottom: 0 }}>
                <select className="form-select">
                  <option value="cpu">CPU</option>
                  <option value="gpu">GPU</option>
                </select>
              </Form.Item>
            </div>
          </div>
          <div className="form-group">
            <label className="form-label">Resource Type</label>
            <Form.Item name="resourceType" style={{ marginBottom: 0 }}>
              <select className="form-select">
                <option value="serverless">Serverless</option>
                <option value="fixed">Fixed</option>
              </select>
            </Form.Item>
          </div>
          <div className="form-row">
            <div className="form-group">
              <label className="form-label">CPU Cores</label>
              <Form.Item name="cpu" style={{ marginBottom: 0 }}>
                <input className="form-input" />
              </Form.Item>
            </div>
            <div className="form-group">
              <label className="form-label"><span className="required">*</span> Memory</label>
              <Form.Item name="memory" rules={[{ required: true, message: 'Required' }]} style={{ marginBottom: 0 }}>
                <input className="form-input" />
              </Form.Item>
            </div>
          </div>
          <div className="form-row">
            <div className="form-group">
              <label className="form-label">GPU Count</label>
              <Form.Item name="gpu" style={{ marginBottom: 0 }}>
                <input className="form-input" />
              </Form.Item>
            </div>
            <div className="form-group">
              <label className="form-label">GPU Type</label>
              <Form.Item name="gpuType" style={{ marginBottom: 0 }}>
                <input className="form-input" />
              </Form.Item>
            </div>
          </div>
          <div className="form-row">
            <div className="form-group">
              <label className="form-label">Ephemeral Storage (GB)</label>
              <Form.Item name="ephemeralStorage" style={{ marginBottom: 0 }}>
                <input className="form-input" />
              </Form.Item>
            </div>
            <div className="form-group">
              <label className="form-label">Shared Memory</label>
              <Form.Item name="shmSize" style={{ marginBottom: 0 }}>
                <input className="form-input" />
              </Form.Item>
            </div>
          </div>
          <div className="form-group full" style={{ marginTop: 16 }}>
            <label className="form-label">Platform Configurations</label>
            <Form.Item name="platforms" style={{ marginBottom: 0 }}>
              <PlatformConfigEditor />
            </Form.Item>
          </div>
          <div className="modal-footer" style={{ margin: '24px -24px -24px', padding: '16px 24px' }}>
            <button type="button" className="btn btn-outline" onClick={() => { setIsEditModalVisible(false); setEditingSpec(null); editForm.resetFields(); }}>
              Cancel
            </button>
            <button type="button" className="btn btn-blue" onClick={handleUpdate} disabled={updateMutation.isPending}>
              {updateMutation.isPending ? 'Updating...' : 'Update'}
            </button>
          </div>
        </Form>
      </Modal>
    </>
  );
};

export default SpecsPage;
