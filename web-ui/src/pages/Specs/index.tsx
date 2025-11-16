import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Card,
  Row,
  Col,
  Typography,
  Tag,
  Descriptions,
  Divider,
  Empty,
  Spin,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  message,
  Popconfirm,
} from 'antd';
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

const { Title, Paragraph, Text } = Typography;
const { Option } = Select;

const SpecsPage = () => {
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [isCreateModalVisible, setIsCreateModalVisible] = useState(false);
  const [isEditModalVisible, setIsEditModalVisible] = useState(false);
  const [editingSpec, setEditingSpec] = useState<SpecInfo | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const queryClient = useQueryClient();

  // Fetch specs from database (specs API)
  const { data: specs, isLoading, refetch } = useQuery({
    queryKey: ['specs'],
    queryFn: async () => {
      const response = await api.specs.list();
      return response.data;
    },
  });

  // Create spec mutation
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

  // Update spec mutation
  const updateMutation = useMutation({
    mutationFn: ({ name, data }: { name: string; data: any }) =>
      api.specs.update(name, data),
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

  // Delete spec mutation
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

  // Filter specs by category
  const filteredSpecs = specs?.filter(
    (spec) => selectedCategory === 'all' || spec.category === selectedCategory
  );

  // Get unique categories
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
        platforms: values.platforms || {
          generic: {
            nodeSelector: {},
            tolerations: [],
            labels: {},
            annotations: {},
          },
        },
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

  const handleDelete = (name: string) => {
    deleteMutation.mutate(name);
  };

  const renderSpecCard = (spec: SpecInfo) => (
    <Card
      key={spec.name}
      title={
        <Space>
          <DatabaseOutlined />
          <span>{spec.displayName}</span>
        </Space>
      }
      extra={
        <Space>
          <Tag color={spec.category === 'gpu' ? 'blue' : 'green'}>{spec.category.toUpperCase()}</Tag>
          <Tag color={spec.resourceType === 'fixed' ? 'purple' : 'cyan'}>
            {spec.resourceType === 'fixed' ? 'FIXED' : 'SERVERLESS'}
          </Tag>
          <Button
            type="text"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(spec)}
          />
          <Popconfirm
            title="Delete Spec"
            description="Are you sure you want to delete this spec?"
            onConfirm={() => handleDelete(spec.name)}
            okText="Yes"
            cancelText="No"
          >
            <Button
              type="text"
              size="small"
              danger
              icon={<DeleteOutlined />}
            />
          </Popconfirm>
        </Space>
      }
      hoverable
      style={{ marginBottom: 16 }}
    >
      <Descriptions column={1} size="small">
        <Descriptions.Item label="Name">
          <Text code>{spec.name}</Text>
        </Descriptions.Item>

        {spec.resources.gpu && (
          <Descriptions.Item label="GPU">
            <Tag color="blue">
              {spec.resources.gpu} × {spec.resources.gpuType}
            </Tag>
          </Descriptions.Item>
        )}

        <Descriptions.Item label="CPU">
          <Tag>{spec.resources.cpu} cores</Tag>
        </Descriptions.Item>

        <Descriptions.Item label="Memory">
          <Tag>{spec.resources.memory}</Tag>
        </Descriptions.Item>

        {spec.resources.ephemeralStorage && (
          <Descriptions.Item label="Ephemeral Storage">
            <Tag>{spec.resources.ephemeralStorage}GB</Tag>
          </Descriptions.Item>
        )}

        {spec.resources.shmSize && (
          <Descriptions.Item label="Shared Memory (SHM)">
            <Tag>{spec.resources.shmSize}</Tag>
          </Descriptions.Item>
        )}
      </Descriptions>

      {/* Show platform configs */}
      {spec.platforms && Object.keys(spec.platforms).length > 0 && (
        <>
          <Divider style={{ margin: '12px 0' }} />
          <Text type="secondary" style={{ fontSize: 12 }}>
            Supported Platforms: {Object.keys(spec.platforms).join(', ')}
          </Text>
        </>
      )}
    </Card>
  );

  return (
    <div>
      <Row justify="space-between" align="middle" style={{ marginBottom: 12 }}>
        <Col>
          <Title level={3} style={{ margin: 0, marginBottom: 4 }}>
            <DatabaseOutlined /> GPU Specs
          </Title>
          <Paragraph type="secondary" style={{ margin: 0, fontSize: 13 }}>
            Manage GPU and CPU specifications for deployment
          </Paragraph>
        </Col>
        <Col>
          <Space>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setIsCreateModalVisible(true)}
            >
              Create Spec
            </Button>
            <Button icon={<ReloadOutlined />} onClick={() => refetch()} loading={isLoading}>
              Refresh
            </Button>
          </Space>
        </Col>
      </Row>

      {/* Category Filter */}
      <Card style={{ marginBottom: 16 }}>
        <Space wrap>
          <Text strong>Filter by category:</Text>
          {categories.map((cat) => (
            <Tag.CheckableTag
              key={cat}
              checked={selectedCategory === cat}
              onChange={() => setSelectedCategory(cat)}
              style={{ padding: '4px 12px', fontSize: 14 }}
            >
              {cat.toUpperCase()}
            </Tag.CheckableTag>
          ))}
        </Space>
      </Card>

      <Spin spinning={isLoading}>
        {filteredSpecs && filteredSpecs.length > 0 ? (
          <Row gutter={[16, 16]}>
            {filteredSpecs.map((spec) => (
              <Col xs={24} sm={12} lg={8} key={spec.name}>
                {renderSpecCard(spec)}
              </Col>
            ))}
          </Row>
        ) : (
          <Card>
            <Empty description="No specs found" />
          </Card>
        )}
      </Spin>

      {/* Create Modal */}
      <Modal
        title="Create New Spec"
        open={isCreateModalVisible}
        onOk={handleCreate}
        onCancel={() => {
          setIsCreateModalVisible(false);
          createForm.resetFields();
        }}
        okText="Create"
        confirmLoading={createMutation.isPending}
        width={900}
      >
        <Form form={createForm} layout="vertical">
          <Form.Item
            name="name"
            label="Spec Name"
            rules={[{ required: true, message: 'Please enter spec name' }]}
          >
            <Input placeholder="e.g., h200-single" />
          </Form.Item>

          <Form.Item
            name="displayName"
            label="Display Name"
            rules={[{ required: true, message: 'Please enter display name' }]}
          >
            <Input placeholder="e.g., H200 1 GPU" />
          </Form.Item>

          <Form.Item
            name="category"
            label="Category"
            rules={[{ required: true, message: 'Please select category' }]}
          >
            <Select placeholder="Select category">
              <Option value="cpu">CPU</Option>
              <Option value="gpu">GPU</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="resourceType"
            label="Resource Type"
            rules={[{ required: true, message: 'Please select resource type' }]}
            tooltip="Fixed resource pools have dedicated nodes and support debugging features like ptrace. Serverless resources are on-demand."
            initialValue="serverless"
          >
            <Select placeholder="Select resource type">
              <Option value="serverless">Serverless (on-demand)</Option>
              <Option value="fixed">Fixed (dedicated pool)</Option>
            </Select>
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="cpu" label="CPU Cores">
                <Input placeholder="e.g., 4" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="memory"
                label="Memory"
                rules={[{ required: true, message: 'Please enter memory' }]}
              >
                <Input placeholder="e.g., 8Gi" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="gpu" label="GPU Count">
                <Input placeholder="e.g., 1" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="gpuType" label="GPU Type">
                <Input placeholder="e.g., NVIDIA-H200" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="ephemeralStorage"
            label="Ephemeral Storage (GB)"
            rules={[{ required: true, message: 'Please enter ephemeral storage' }]}
          >
            <Input placeholder="e.g., 300" />
          </Form.Item>

          <Form.Item
            name="shmSize"
            label="Shared Memory Size (Optional)"
            tooltip="Shared memory size for /dev/shm (e.g., 1Gi, 512Mi). Useful for ML workloads."
          >
            <Input placeholder="e.g., 1Gi, 512Mi" />
          </Form.Item>

          <Divider>Platform Configuration</Divider>

          <Form.Item
            name="platforms"
            label="Platform-specific Settings"
            tooltip="Configure node selectors, tolerations, labels, and annotations for different platforms"
            initialValue={{
              generic: {
                nodeSelector: {},
                tolerations: [],
                labels: {},
                annotations: {},
              },
            }}
          >
            <PlatformConfigEditor />
          </Form.Item>
        </Form>
      </Modal>

      {/* Edit Modal */}
      <Modal
        title={`Edit Spec: ${editingSpec?.name}`}
        open={isEditModalVisible}
        onOk={handleUpdate}
        onCancel={() => {
          setIsEditModalVisible(false);
          setEditingSpec(null);
          editForm.resetFields();
        }}
        okText="Update"
        confirmLoading={updateMutation.isPending}
        width={900}
      >
        <Form form={editForm} layout="vertical">
          <Form.Item
            name="displayName"
            label="Display Name"
            rules={[{ required: true, message: 'Please enter display name' }]}
          >
            <Input placeholder="e.g., H200 1 GPU" />
          </Form.Item>

          <Form.Item
            name="category"
            label="Category"
            rules={[{ required: true, message: 'Please select category' }]}
          >
            <Select placeholder="Select category">
              <Option value="cpu">CPU</Option>
              <Option value="gpu">GPU</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="resourceType"
            label="Resource Type"
            rules={[{ required: true, message: 'Please select resource type' }]}
            tooltip="Fixed resource pools have dedicated nodes and support debugging features like ptrace. Serverless resources are on-demand."
          >
            <Select placeholder="Select resource type">
              <Option value="serverless">Serverless (on-demand)</Option>
              <Option value="fixed">Fixed (dedicated pool)</Option>
            </Select>
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="cpu" label="CPU Cores">
                <Input placeholder="e.g., 4" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="memory"
                label="Memory"
                rules={[{ required: true, message: 'Please enter memory' }]}
              >
                <Input placeholder="e.g., 8Gi" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="gpu" label="GPU Count">
                <Input placeholder="e.g., 1" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="gpuType" label="GPU Type">
                <Input placeholder="e.g., NVIDIA-H200" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="ephemeralStorage"
            label="Ephemeral Storage (GB)"
            rules={[{ required: true, message: 'Please enter ephemeral storage' }]}
          >
            <Input placeholder="e.g., 300" />
          </Form.Item>

          <Form.Item
            name="shmSize"
            label="Shared Memory Size (Optional)"
            tooltip="Shared memory size for /dev/shm (e.g., 1Gi, 512Mi). Useful for ML workloads."
          >
            <Input placeholder="e.g., 1Gi, 512Mi" />
          </Form.Item>

          <Divider>Platform Configuration</Divider>

          <Form.Item
            name="platforms"
            label="Platform-specific Settings"
            tooltip="Configure node selectors, tolerations, labels, and annotations for different platforms"
          >
            <PlatformConfigEditor />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default SpecsPage;
