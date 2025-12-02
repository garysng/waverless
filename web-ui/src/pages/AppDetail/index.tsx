import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Alert,
  Badge,
  Breadcrumb,
  Button,
  Card,
  Col,
  Divider,
  Dropdown,
  Form,
  Input,
  InputNumber,
  Modal,
  Row,
  Select,
  Space,
  Statistic,
  Switch,
  Tabs,
  Tooltip,
  Typography,
  message,
} from 'antd';
import type { MenuProps } from 'antd';
import {
  ArrowLeftOutlined,
  DeleteOutlined,
  EditOutlined,
  ReloadOutlined,
  MoreOutlined,
  RocketOutlined,
  SettingOutlined,
  UserOutlined,
  DatabaseOutlined,
  HistoryOutlined,
  ThunderboltOutlined,
  QuestionCircleOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { api } from '@/api/client';
import type { AppInfo, UpdateDeploymentRequest, UpdateEndpointConfigRequest, SpecInfo, PVCInfo } from '@/types';
import OverviewTab from './tabs/OverviewTab';
import WorkersTab from './tabs/WorkersTab';
import TasksTab from './tabs/TasksTab';
import ScalingHistoryTab from './tabs/ScalingHistoryTab';

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;

const GLOBAL_TASK_TIMEOUT = 3600;

const AUTOSCALER_FIELD_TIPS = {
  priority: 'Higher numbers get resources first. Default 50.',
  scaleUpThreshold: 'Pending tasks required before adding replicas (>=1).',
  scaleDownIdleTime: 'Idle seconds before replicas are removed. Larger = fewer flaps.',
  scaleUpCooldown: 'Minimum seconds between two scale-up actions.',
  scaleDownCooldown: 'Minimum seconds between two scale-down actions.',
  highLoadThreshold: 'Queue length treated as "high load" for dynamic priority.',
  priorityBoost: 'How many priority points to add when high load is detected.',
} as const;

type AutoScalerTipKey = keyof typeof AUTOSCALER_FIELD_TIPS;

const renderFieldLabel = (label: string, key: AutoScalerTipKey) => (
  <span>
    {label}
    <Tooltip title={AUTOSCALER_FIELD_TIPS[key]}>
      <QuestionCircleOutlined style={{ marginLeft: 4, color: '#999' }} />
    </Tooltip>
  </span>
);

const AppDetailPage = () => {
  const { endpoint } = useParams<{ endpoint: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState('overview');
  const [deleteModalVisible, setDeleteModalVisible] = useState(false);
  const [updateModalVisible, setUpdateModalVisible] = useState(false);
  const [editConfigModalVisible, setEditConfigModalVisible] = useState(false);
  const [selectedUpdateSpec, setSelectedUpdateSpec] = useState<string>('');
  const [jumpToWorker, setJumpToWorker] = useState<string | null>(null);

  const [updateForm] = Form.useForm();
  const [editConfigForm] = Form.useForm();

  // Fetch app details
  const { data: appInfo, isLoading, refetch } = useQuery<AppInfo>({
    queryKey: ['app', endpoint],
    queryFn: async () => {
      if (!endpoint) throw new Error('Endpoint is required');
      const response = await api.apps.get(endpoint);
      const app = response.data;
      // Compute imageUpdateAvailable
      return {
        ...app,
        imageUpdateAvailable: !!app.latestImage && app.latestImage !== app.image,
      };
    },
    enabled: !!endpoint,
    refetchInterval: 10000,
  });

  // Fetch workers for accurate statistics
  const { data: workers } = useQuery({
    queryKey: ['workers', endpoint],
    queryFn: async () => {
      if (!endpoint) return [];
      const response = await api.apps.workers(endpoint);
      return response.data;
    },
    enabled: !!endpoint,
    refetchInterval: 5000,
  });

  // Fetch specs for update form
  const { data: specs } = useQuery<SpecInfo[]>({
    queryKey: ['specs'],
    queryFn: async () => {
      const response = await api.specs.list();
      return response.data;
    },
  });

  // Fetch PVCs
  const { data: pvcs } = useQuery<PVCInfo[]>({
    queryKey: ['pvcs'],
    queryFn: async () => {
      const response = await api.k8s.listPVCs();
      return response.data;
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: async () => {
      if (!endpoint) throw new Error('Endpoint is required');
      await api.apps.delete(endpoint);
    },
    onSuccess: () => {
      message.success(`Endpoint ${endpoint} deleted successfully`);
      queryClient.invalidateQueries({ queryKey: ['apps'] });
      navigate('/apps');
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to delete endpoint');
    },
  });

  // Update deployment mutation
  const updateMutation = useMutation({
    mutationFn: async (data: { name: string; request: UpdateDeploymentRequest }) => {
      const response = await api.apps.update(data.name, data.request);
      return response.data;
    },
    onSuccess: () => {
      message.success('Deployment updated successfully');
      setUpdateModalVisible(false);
      updateForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['apps'] });
      queryClient.invalidateQueries({ queryKey: ['app', endpoint] });
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to update deployment');
    },
  });

  // Update config mutation
  const updateConfigMutation = useMutation({
    mutationFn: async (data: { name: string; config: UpdateEndpointConfigRequest }) => {
      const response = await api.apps.updateMetadata(data.name, data.config);
      return response.data;
    },
    onSuccess: () => {
      message.success('Configuration updated successfully');
      setEditConfigModalVisible(false);
      editConfigForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['apps'] });
      queryClient.invalidateQueries({ queryKey: ['app', endpoint] });
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to update configuration');
    },
  });

  // Trigger autoscaler mutation
  const triggerAutoscalerMutation = useMutation({
    mutationFn: async (endpointName: string) => {
      await api.autoscaler.trigger(endpointName);
    },
    onSuccess: () => {
      message.success(`AutoScaler triggered for ${endpoint}`);
      queryClient.invalidateQueries({ queryKey: ['apps'] });
      queryClient.invalidateQueries({ queryKey: ['app', endpoint] });
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to trigger autoscaler');
    },
  });

  // Check image update mutation
  const checkImageMutation = useMutation({
    mutationFn: async () => {
      if (!endpoint) throw new Error('Endpoint is required');
      const response = await api.apps.checkImage(endpoint);
      return response.data;
    },
    onSuccess: (data) => {
      if (data.updateAvailable) {
        message.info({
          content: `New image version available: ${data.newDigest?.substring(0, 19)}...`,
          duration: 5,
        });
      } else {
        message.success('Image is up to date');
      }
      queryClient.invalidateQueries({ queryKey: ['apps'] });
      queryClient.invalidateQueries({ queryKey: ['app', endpoint] });
      refetch();
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to check image update');
    },
  });

  // Scroll to top when page loads
  useEffect(() => {
    window.scrollTo({ top: 0, behavior: 'instant' });
  }, [endpoint]);

  const handleDelete = () => {
    setDeleteModalVisible(true);
  };

  const confirmDelete = () => {
    deleteMutation.mutate();
    setDeleteModalVisible(false);
  };

  const handleEditConfig = () => {
    if (!appInfo) return;

    editConfigForm.setFieldsValue({
      displayName: appInfo.displayName || appInfo.name,
      description: appInfo.description || '',
      minReplicas: appInfo.minReplicas ?? 0,
      maxReplicas: appInfo.maxReplicas ?? 1,
      priority: appInfo.priority ?? 50,
      taskTimeout: appInfo.taskTimeout ?? 0,
      maxPendingTasks: appInfo.maxPendingTasks ?? 1,
      imagePrefix: appInfo.imagePrefix || '',
      autoscalerEnabled: appInfo.autoscalerEnabled || '',
      scaleUpThreshold: appInfo.scaleUpThreshold ?? 1,
      scaleDownIdleTime: appInfo.scaleDownIdleTime ?? 300,
      scaleUpCooldown: appInfo.scaleUpCooldown ?? 30,
      scaleDownCooldown: appInfo.scaleDownCooldown ?? 60,
      enableDynamicPrio: appInfo.enableDynamicPrio ?? true,
      highLoadThreshold: appInfo.highLoadThreshold ?? 10,
      priorityBoost: appInfo.priorityBoost ?? 20,
    });
    setEditConfigModalVisible(true);
  };

  const handleSaveConfig = () => {
    editConfigForm.validateFields().then((values) => {
      if (!endpoint) return;

      const config: UpdateEndpointConfigRequest = {
        displayName: values.displayName,
        description: values.description,
        taskTimeout: values.taskTimeout,
        maxPendingTasks: values.maxPendingTasks,
        imagePrefix: values.imagePrefix !== undefined ? values.imagePrefix : undefined,
        minReplicas: values.minReplicas,
        maxReplicas: values.maxReplicas,
        priority: values.priority,
        scaleUpThreshold: values.scaleUpThreshold,
        scaleDownIdleTime: values.scaleDownIdleTime,
        scaleUpCooldown: values.scaleUpCooldown,
        scaleDownCooldown: values.scaleDownCooldown,
        enableDynamicPrio: values.enableDynamicPrio,
        highLoadThreshold: values.highLoadThreshold,
        priorityBoost: values.priorityBoost,
        autoscalerEnabled: values.autoscalerEnabled,
      };

      updateConfigMutation.mutate({ name: endpoint, config });
    });
  };

  const handleUpdateDeployment = async () => {
    if (!appInfo) return;

    setSelectedUpdateSpec(appInfo.specName || '');

    // Convert env object to array format
    let envVars: Array<{ key: string; value: string }> = [];
    if (appInfo.env && Object.keys(appInfo.env).length > 0) {
      envVars = Object.entries(appInfo.env).map(([key, value]) => ({ key, value }));
    } else {
      try {
        const response = await api.config.getDefaultEnv();
        if (response.data && Object.keys(response.data).length > 0) {
          envVars = Object.entries(response.data).map(([key, value]) => ({ key, value }));
        }
      } catch (error) {
        console.error('Failed to fetch default env:', error);
      }
    }

    updateForm.setFieldsValue({
      specName: appInfo.specName || '',
      image: appInfo.image || '',
      replicas: appInfo.replicas || 1,
      shmSize: appInfo.shmSize || '',
      volumeMounts: appInfo.volumeMounts || [],
      enablePtrace: appInfo.enablePtrace || false,
      envVars: envVars,
    });
    setUpdateModalVisible(true);
  };

  const handleUpdate = (values: any) => {
    if (!endpoint) return;

    const request: UpdateDeploymentRequest = { endpoint };
    if (values.specName) request.specName = values.specName;
    if (values.image) request.image = values.image;
    if (values.replicas !== undefined && values.replicas !== null) {
      request.replicas = values.replicas;
    }
    if (values.volumeMounts && values.volumeMounts.length > 0) {
      request.volumeMounts = values.volumeMounts
        .filter((vm: any) => vm && vm.pvcName && vm.mountPath)
        .map((vm: any) => ({ pvcName: vm.pvcName, mountPath: vm.mountPath }));
    }
    if (values.shmSize) request.shmSize = values.shmSize;
    if (typeof values.enablePtrace === 'boolean') request.enablePtrace = values.enablePtrace;
    if (values.envVars && Array.isArray(values.envVars)) {
      const env: Record<string, string> = {};
      values.envVars.forEach((item: { key: string; value: string }) => {
        if (item.key && item.value) {
          env[item.key] = item.value;
        }
      });
      if (Object.keys(env).length > 0) {
        request.env = env;
      }
    }

    updateMutation.mutate({ name: endpoint, request });
  };

  const handleTriggerAutoscaler = () => {
    if (!endpoint) return;
    triggerAutoscalerMutation.mutate(endpoint);
  };

  // Handle jump to worker from tasks tab
  const handleJumpToWorker = (workerId: string) => {
    setJumpToWorker(workerId);
    setActiveTab('workers');
  };

  // More actions menu
  const moreActions: MenuProps['items'] = [
    {
      key: 'trigger',
      icon: <ThunderboltOutlined />,
      label: 'Trigger AutoScaler',
      onClick: handleTriggerAutoscaler,
    },
    {
      type: 'divider',
    },
    {
      key: 'delete',
      icon: <DeleteOutlined />,
      label: 'Delete',
      danger: true,
      onClick: handleDelete,
    },
  ];

  // Get status badge
  const getStatusBadge = (status?: string) => {
    const statusMap: Record<string, { status: 'success' | 'processing' | 'error' | 'default'; text: string }> = {
      Running: { status: 'success', text: 'Running' },
      Pending: { status: 'processing', text: 'Pending' },
      Failed: { status: 'error', text: 'Failed' },
    };
    const config = statusMap[status || ''] || { status: 'default' as const, text: status || 'Unknown' };
    return <Badge status={config.status} text={config.text} />;
  };

  if (!endpoint) {
    return <Alert message="Error" description="Endpoint parameter is missing" type="error" showIcon />;
  }

  return (
    <div style={{ background: '#f0f2f5', minHeight: 'calc(100vh - 112px)', margin: '-24px -16px', padding: '24px' }}>
      {/* Breadcrumb Navigation */}
      <div style={{ marginBottom: 16 }}>
        <Breadcrumb
          items={[
            {
              title: (
                <Button type="link" icon={<ArrowLeftOutlined />} onClick={() => navigate('/apps')} style={{ padding: 0 }}>
                  Applications
                </Button>
              ),
            },
            { title: endpoint },
          ]}
        />
      </div>

      {/* Image Update Alert */}
      {appInfo?.imageUpdateAvailable && (
        <Alert
          message="New Docker Image Version Available"
          description={
            <Space direction="vertical" size="small">
              <Text>A new version of the Docker image is available for this endpoint.</Text>
              {appInfo.latestImage && (
                <Text code style={{ fontSize: 11 }}>
                  Latest: {appInfo.latestImage}
                </Text>
              )}
              {appInfo.imageLastChecked && (
                <Text type="secondary">
                  Last checked: {new Date(appInfo.imageLastChecked).toLocaleString()}
                </Text>
              )}
            </Space>
          }
          type="warning"
          showIcon
          closable
          style={{ marginBottom: 16 }}
          action={
            <Button
              size="small"
              type="primary"
              icon={<RocketOutlined />}
              onClick={handleUpdateDeployment}
            >
              Update Now
            </Button>
          }
        />
      )}

      {/* Header Section */}
      <Card loading={isLoading} style={{ marginBottom: 16 }} bodyStyle={{ padding: '24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          {/* Left: Title and Status */}
          <div style={{ flex: 1 }}>
            <Space direction="vertical" size="small">
              <Space>
                <Title level={3} style={{ margin: 0 }}>
                  {endpoint}
                </Title>
                {getStatusBadge(appInfo?.status)}
              </Space>
              {appInfo?.displayName && <Text type="secondary">{appInfo.displayName}</Text>}
              {appInfo?.description && <Text type="secondary">{appInfo.description}</Text>}
            </Space>

            {/* Statistics Row */}
            <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
              <Col xs={12} sm={8} md={6} lg={4}>
                <Card size="small" style={{ height: '100%' }}>
                  <Statistic
                    title={<Text type="secondary">Workers</Text>}
                    value={workers?.filter((w: any) => w.status?.toLowerCase() === 'busy').length || 0}
                    suffix={<Text type="secondary">/ {workers?.length || 0}</Text>}
                    prefix={<UserOutlined style={{ color: '#1890ff' }} />}
                    valueStyle={{ color: workers?.filter((w: any) => w.status?.toLowerCase() === 'busy').length ? '#1890ff' : undefined }}
                  />
                </Card>
              </Col>
              <Col xs={12} sm={8} md={6} lg={4}>
                <Card size="small" style={{ height: '100%' }}>
                  <Statistic
                    title={<Text type="secondary">Replicas</Text>}
                    value={appInfo?.readyReplicas || 0}
                    suffix={<Text type="secondary">/ {appInfo?.replicas || 0}</Text>}
                    prefix={<DatabaseOutlined style={{ color: '#52c41a' }} />}
                    valueStyle={{ color: appInfo?.readyReplicas ? '#52c41a' : undefined }}
                  />
                </Card>
              </Col>
              <Col xs={12} sm={8} md={6} lg={4}>
                <Card size="small" style={{ height: '100%' }}>
                  <Statistic
                    title={<Text type="secondary">Running</Text>}
                    value={appInfo?.runningTasks || 0}
                    prefix={<ThunderboltOutlined style={{ color: '#1890ff' }} />}
                    valueStyle={{ color: appInfo?.runningTasks ? '#1890ff' : undefined }}
                  />
                </Card>
              </Col>
              <Col xs={12} sm={8} md={6} lg={4}>
                <Card size="small" style={{ height: '100%' }}>
                  <Statistic
                    title={<Text type="secondary">Pending</Text>}
                    value={appInfo?.pendingTasks || 0}
                    prefix={<ThunderboltOutlined style={{ color: '#faad14' }} />}
                    valueStyle={{ color: appInfo?.pendingTasks ? '#faad14' : undefined }}
                  />
                </Card>
              </Col>
              <Col xs={12} sm={8} md={6} lg={4}>
                <Card size="small" style={{ height: '100%' }}>
                  <Statistic
                    title={<Text type="secondary">Total</Text>}
                    value={appInfo?.totalTasks || 0}
                    prefix={<HistoryOutlined />}
                  />
                </Card>
              </Col>
            </Row>
          </div>

          {/* Right: Action Buttons */}
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
              Refresh
            </Button>
            <Button type="primary" icon={<RocketOutlined />} onClick={handleUpdateDeployment}>
              Update
            </Button>
            <Button icon={<EditOutlined />} onClick={handleEditConfig}>
              Configure
            </Button>
            <Dropdown menu={{ items: moreActions }} placement="bottomRight">
              <Button icon={<MoreOutlined />} loading={triggerAutoscalerMutation.isPending} />
            </Dropdown>
          </Space>
        </div>
      </Card>

      {/* Tabs Section */}
      <Card bodyStyle={{ padding: 0 }}>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            {
              key: 'overview',
              label: (
                <span>
                  <DatabaseOutlined />
                  Overview
                </span>
              ),
              children: (
                <OverviewTab
                  endpoint={endpoint}
                  appInfo={appInfo}
                  onCheckImage={() => checkImageMutation.mutate()}
                  checkingImage={checkImageMutation.isPending}
                />
              ),
            },
            {
              key: 'workers',
              label: (
                <span>
                  <UserOutlined />
                  Workers
                </span>
              ),
              children: <WorkersTab endpoint={endpoint} jumpToWorker={jumpToWorker} onClearJumpToWorker={() => setJumpToWorker(null)} />,
            },
            {
              key: 'tasks',
              label: (
                <span>
                  <DatabaseOutlined />
                  Tasks
                </span>
              ),
              children: <TasksTab endpoint={endpoint} onJumpToWorker={handleJumpToWorker} workers={workers || []} />,
            },
            {
              key: 'scaling',
              label: (
                <span>
                  <HistoryOutlined />
                  Scaling History
                </span>
              ),
              children: <ScalingHistoryTab endpoint={endpoint} />,
            },
          ]}
          style={{ padding: '0 24px' }}
        />
      </Card>

      {/* Delete Confirmation Modal */}
      <Modal
        title="Confirm Delete"
        open={deleteModalVisible}
        onOk={confirmDelete}
        onCancel={() => setDeleteModalVisible(false)}
        okText="Delete"
        okButtonProps={{ danger: true, loading: deleteMutation.isPending }}
      >
        <p>
          Are you sure you want to delete endpoint <strong>{endpoint}</strong>?
        </p>
        <p>This action cannot be undone.</p>
      </Modal>

      {/* Update Deployment Modal */}
      <Modal
        title="Update Deployment"
        open={updateModalVisible}
        onOk={() => updateForm.submit()}
        onCancel={() => {
          setUpdateModalVisible(false);
          setSelectedUpdateSpec('');
          updateForm.resetFields();
        }}
        confirmLoading={updateMutation.isPending}
        width={720}
      >
        <Form form={updateForm} layout="vertical" onFinish={handleUpdate}>
          <Form.Item label="Spec" name="specName">
            <Select
              placeholder="Select spec"
              allowClear
              showSearch
              onChange={(value) => {
                setSelectedUpdateSpec(value || '');
                const spec = specs?.find((s) => s.name === value);
                if (spec?.resources?.shmSize) {
                  updateForm.setFieldValue('shmSize', spec.resources.shmSize);
                }
              }}
              options={specs?.map((spec) => ({
                label: `${spec.name} (${spec.category})`,
                value: spec.name,
              }))}
            />
          </Form.Item>
          <Form.Item label="Image" name="image">
            <Input placeholder="registry/image:tag" />
          </Form.Item>
          <Form.Item label="Replicas" name="replicas">
            <InputNumber min={0} max={100} style={{ width: '100%' }} />
          </Form.Item>

          <Divider orientation="left" style={{ marginTop: 8, marginBottom: 16 }}>
            <span>
              <DatabaseOutlined /> Volume Mounts (Optional)
            </span>
          </Divider>
          <Paragraph type="secondary" style={{ fontSize: 12, marginBottom: 16 }}>
            Configure PVC volume mounts. Leave empty to keep existing mounts unchanged.
          </Paragraph>

          <Form.List name="volumeMounts">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Card key={key} size="small" style={{ marginBottom: 12 }}>
                    <Row gutter={12}>
                      <Col span={11}>
                        <Form.Item
                          {...restField}
                          name={[name, 'pvcName']}
                          label="PVC Name"
                          rules={[{ required: true, message: 'Select PVC' }]}
                        >
                          <Select
                            placeholder="Select PVC"
                            showSearch
                            optionLabelProp="label"
                          >
                            {pvcs?.map((pvc) => (
                              <Select.Option
                                key={pvc.name}
                                value={pvc.name}
                                label={pvc.name}
                              >
                                <div>{pvc.name}</div>
                                <div style={{ fontSize: 11, color: '#999' }}>
                                  {pvc.capacity} | {pvc.status}
                                </div>
                              </Select.Option>
                            ))}
                          </Select>
                        </Form.Item>
                      </Col>
                      <Col span={11}>
                        <Form.Item
                          {...restField}
                          name={[name, 'mountPath']}
                          label="Mount Path"
                          rules={[
                            { required: true, message: 'Enter mount path' },
                            { pattern: /^\/.*/, message: 'Must start with /' },
                          ]}
                        >
                          <Input placeholder="/data" />
                        </Form.Item>
                      </Col>
                      <Col span={2} style={{ display: 'flex', alignItems: 'center', paddingTop: 30 }}>
                        <Button danger size="small" icon={<DeleteOutlined />} onClick={() => remove(name)} />
                      </Col>
                    </Row>
                  </Card>
                ))}
                <Form.Item>
                  <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />} size="small">
                    Add Volume Mount
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>

          <Divider style={{ marginTop: 8, marginBottom: 16 }}>
            <DatabaseOutlined /> Shared Memory
          </Divider>
          <Paragraph type="secondary" style={{ fontSize: 12, marginBottom: 16 }}>
            Configure shared memory size for /dev/shm. Leave empty to keep existing setting.
          </Paragraph>

          <Form.Item
            name="shmSize"
            label="Shared Memory Size"
            tooltip="Size for /dev/shm (e.g., 1Gi, 512Mi). Empty to keep current, or specify new value."
          >
            <Input placeholder="e.g., 1Gi, 512Mi" />
          </Form.Item>

          {specs?.find((s) => s.name === selectedUpdateSpec)?.resourceType === 'fixed' && (
            <>
              <Divider style={{ marginTop: 8, marginBottom: 16 }}>
                <DatabaseOutlined /> Debugging (ptrace)
              </Divider>
              <Paragraph type="secondary" style={{ fontSize: 12, marginBottom: 16 }}>
                Enable SYS_PTRACE capability for debugging tools (gdb, strace, etc.). Only available for fixed resource pools.
              </Paragraph>

              <Form.Item
                name="enablePtrace"
                label="Enable Debugging (ptrace)"
                tooltip="Enable SYS_PTRACE capability for debugging. Only works for fixed resource pool specs."
                valuePropName="checked"
              >
                <Switch checkedChildren="Enabled" unCheckedChildren="Disabled" />
              </Form.Item>
            </>
          )}

          <Divider style={{ marginTop: 8, marginBottom: 16 }}>
            <SettingOutlined /> Environment Variables
          </Divider>
          <Paragraph type="secondary" style={{ fontSize: 12, marginBottom: 16 }}>
            Configure custom environment variables for your application. Leave empty to keep existing variables unchanged.
          </Paragraph>

          <Form.List name="envVars">
            {(fields, { add, remove }) => (
              <>
                {fields.map((field) => (
                  <Row gutter={16} key={field.key} style={{ marginBottom: 8 }}>
                    <Col span={10}>
                      <Form.Item
                        {...field}
                        name={[field.name, 'key']}
                        rules={[{ required: true, message: 'Required' }]}
                        style={{ marginBottom: 0 }}
                      >
                        <Input placeholder="KEY" />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        {...field}
                        name={[field.name, 'value']}
                        rules={[{ required: true, message: 'Required' }]}
                        style={{ marginBottom: 0 }}
                      >
                        <Input placeholder="value" />
                      </Form.Item>
                    </Col>
                    <Col span={2}>
                      <Button type="text" danger icon={<DeleteOutlined />} onClick={() => remove(field.name)} />
                    </Col>
                  </Row>
                ))}
                <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />} size="small" style={{ marginTop: 8 }}>
                  Add Environment Variable
                </Button>
              </>
            )}
          </Form.List>
        </Form>
      </Modal>

      {/* Edit Config Modal */}
      <Modal
        title="Edit Metadata & AutoScaler Config"
        open={editConfigModalVisible}
        onOk={handleSaveConfig}
        onCancel={() => {
          setEditConfigModalVisible(false);
          editConfigForm.resetFields();
        }}
        confirmLoading={updateConfigMutation.isPending}
        width={600}
      >
        <Form form={editConfigForm} layout="vertical">
          <Divider>Basic Information</Divider>
          <Form.Item label="Display Name" name="displayName">
            <Input placeholder="Friendly display name" />
          </Form.Item>
          <Form.Item label="Description" name="description">
            <TextArea rows={2} placeholder="Application description" />
          </Form.Item>
          <Form.Item
            label="Task Timeout (seconds)"
            name="taskTimeout"
            help={`0 = use global default (${GLOBAL_TASK_TIMEOUT}s)`}
          >
            <InputNumber min={0} placeholder={`0 = use global default`} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item
            label="Max Pending Tasks"
            name="maxPendingTasks"
            help="Maximum allowed pending tasks before warning clients to not submit new tasks"
            tooltip="When pending tasks reach this threshold, the /check endpoint will return can_submit=false"
          >
            <InputNumber min={1} max={1000} placeholder="1" style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item
            label="Image Prefix"
            name="imagePrefix"
            help="Optional prefix for automatic image update detection"
            tooltip="E.g., 'wavespeed/model-deploy:wan_i2v-default-' for images like 'wavespeed/model-deploy:wan_i2v-default-202511051642'"
          >
            <Input placeholder="e.g., wavespeed/model-deploy:wan_i2v-default-" />
          </Form.Item>

          <Divider>AutoScaler Configuration</Divider>
          <Form.Item
            label="AutoScaler Override"
            name="autoscalerEnabled"
            help="Control autoscaling for this endpoint independently of global setting"
            tooltip="Default = follow global setting, Force Off = always disabled, Force On = always enabled"
          >
            <Select
              allowClear
              placeholder="Default (follow global)"
              options={[
                { value: '', label: 'Default (follow global)' },
                { value: 'disabled', label: 'Force Off' },
                { value: 'enabled', label: 'Force On' },
              ]}
            />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label="Min Replicas" name="minReplicas" help="Set to 0 to allow scale-to-zero when idle">
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="Max Replicas" name="maxReplicas">
                <InputNumber min={1} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item
            label={renderFieldLabel('Priority', 'priority')}
            name="priority"
            help="0 = lowest priority (best-effort). Higher values reserve resources sooner."
          >
            <InputNumber min={0} max={100} style={{ width: '100%' }} />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label={renderFieldLabel('Scale Up Threshold', 'scaleUpThreshold')}
                name="scaleUpThreshold"
                help="0 behaves like 1 (scale as soon as a task waits)."
              >
                <InputNumber min={1} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label={renderFieldLabel('Scale Down Idle Time (s)', 'scaleDownIdleTime')}
                name="scaleDownIdleTime"
                help="0 = immediate scale down when idle."
              >
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label={renderFieldLabel('Scale Up Cooldown (s)', 'scaleUpCooldown')}
                name="scaleUpCooldown"
                help="0 = no cooldown between scale ups."
              >
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label={renderFieldLabel('Scale Down Cooldown (s)', 'scaleDownCooldown')}
                name="scaleDownCooldown"
                help="0 = no cooldown between scale downs."
              >
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item
            label="Enable Dynamic Priority"
            name="enableDynamicPrio"
            valuePropName="checked"
            tooltip="Automatically boost priority when queue is long"
          >
            <Switch />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label={renderFieldLabel('High Load Threshold', 'highLoadThreshold')}
                name="highLoadThreshold"
                tooltip="Queue length to trigger priority boost"
                help="0 = always considered high load."
              >
                <InputNumber min={1} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label={renderFieldLabel('Priority Boost', 'priorityBoost')}
                name="priorityBoost"
                tooltip="Priority increase when high load threshold met"
                help="0 = no priority bump even under high load."
              >
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  );
};

export default AppDetailPage;
