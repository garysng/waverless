import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Button,
  Card,
  Descriptions,
  Drawer,
  Empty,
  Space,
  Table,
  Tabs,
  Tag,
  Timeline,
  Tooltip,
  Typography,
} from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  CodeOutlined,
  FileTextOutlined,
  HistoryOutlined,
  LoadingOutlined,
  QuestionCircleOutlined,
  SyncOutlined,
  UnorderedListOutlined,
} from '@ant-design/icons';
import { api } from '@/api/client';
import type { WorkerWithPodInfo, Worker, PodDetail, TaskListResponse } from '@/types';
import Terminal from '@/components/Terminal';

const { Text } = Typography;

interface WorkersTabProps {
  endpoint: string;
  jumpToWorker?: string | null;
  onClearJumpToWorker?: () => void;
}

const WorkersTab = ({ endpoint, jumpToWorker, onClearJumpToWorker }: WorkersTabProps) => {
  const [workerDrawerVisible, setWorkerDrawerVisible] = useState(false);
  const [workerDrawerTab, setWorkerDrawerTab] = useState<string>('tasks');
  const [selectedWorker, setSelectedWorker] = useState<{ endpoint: string; workerId: string } | null>(null);
  const [selectedPod, setSelectedPod] = useState<{ endpoint: string; podName: string } | null>(null);

  // Fetch workers
  const { data: workers, isLoading: loadingWorkers } = useQuery({
    queryKey: ['workers', endpoint],
    queryFn: async () => {
      const response = await api.apps.workers(endpoint);
      return response.data;
    },
    enabled: !!endpoint,
    refetchInterval: 5000,
    placeholderData: (previousData) => previousData,
  });

  // Handle jump to worker from external trigger
  useEffect(() => {
    if (jumpToWorker && endpoint && workers) {
      setSelectedWorker({ endpoint, workerId: jumpToWorker });
      // Find the worker to get the correct pod name
      const worker = workers.find((w) => w.id === jumpToWorker);
      const podName = worker?.pod_name || jumpToWorker;
      setSelectedPod({ endpoint, podName });
      setWorkerDrawerTab('tasks');
      setWorkerDrawerVisible(true);
      // Clear the jump trigger
      if (onClearJumpToWorker) {
        onClearJumpToWorker();
      }
    }
  }, [jumpToWorker, endpoint, workers, onClearJumpToWorker]);

  // Fetch worker logs
  const { data: logsText, isFetching: loadingLogs, refetch: refetchLogs } = useQuery({
    queryKey: ['app-logs', endpoint, selectedWorker?.workerId],
    queryFn: async () => {
      if (!selectedWorker) return null;
      const response = await api.apps.logs(selectedWorker.endpoint, 200, selectedWorker.workerId);
      return response.data;
    },
    enabled: !!selectedWorker && workerDrawerVisible && workerDrawerTab === 'logs',
    refetchInterval: workerDrawerVisible && workerDrawerTab === 'logs' ? 3000 : false,
  });

  // Fetch Pod describe details
  const { data: podDetail, isFetching: loadingPodDetail } = useQuery<PodDetail>({
    queryKey: ['pod-describe', selectedPod?.endpoint, selectedPod?.podName],
    queryFn: async () => {
      if (!selectedPod) throw new Error('No pod selected');
      const response = await api.apps.describePod(selectedPod.endpoint, selectedPod.podName);
      return response.data;
    },
    enabled: !!selectedPod && workerDrawerVisible && workerDrawerTab === 'details',
  });

  // Fetch worker tasks
  const { data: workerTasksData, isLoading: loadingWorkerTasks } = useQuery<TaskListResponse>({
    queryKey: ['worker-tasks', endpoint, selectedWorker?.workerId],
    queryFn: async () => {
      if (!selectedWorker) throw new Error('No worker selected');
      const response = await api.tasks.list({
        endpoint,
        limit: 1000,
      });
      return response.data;
    },
    enabled: !!selectedWorker && workerDrawerVisible && workerDrawerTab === 'tasks',
    placeholderData: (previousData) => previousData,
  });

  // Filter tasks by worker ID
  const workerTasks = workerTasksData?.tasks.filter(
    (task) => task.workerId === selectedWorker?.workerId
  ) || [];

  const workerColumns = [
    {
      title: 'Worker / Pod',
      dataIndex: 'id',
      key: 'id',
      width: 250,
      render: (_: string, record: WorkerWithPodInfo) => (
        <Space direction="vertical" size={0}>
          {record.id ? (
            <Tooltip title={record.id}>
              <Text code style={{ fontSize: 12 }}>
                {record.id.substring(0, 24)}
              </Text>
            </Tooltip>
          ) : (
            <Text type="secondary" style={{ fontSize: 12 }}>
              No Worker ID
            </Text>
          )}
          {record.pod_name && record.pod_name !== record.id && (
            <Tooltip title={record.pod_name}>
              <Text type="secondary" style={{ fontSize: 11 }}>
                Pod: {record.pod_name.substring(0, 20)}...
              </Text>
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (_: string, record: WorkerWithPodInfo) => {
        // Worker status
        let workerColor = 'default';
        let workerIcon = <QuestionCircleOutlined />;
        const statusLower = record.status?.toLowerCase();
        if (statusLower === 'online') {
          workerColor = 'green';
          workerIcon = <CheckCircleOutlined />;
        } else if (statusLower === 'offline') {
          workerColor = 'red';
          workerIcon = <CloseCircleOutlined />;
        } else if (statusLower === 'busy') {
          workerColor = 'blue';
          workerIcon = <LoadingOutlined />;
        } else if (statusLower === 'draining') {
          workerColor = 'orange';
          workerIcon = <SyncOutlined spin />;
        }

        return (
          <Space direction="vertical" size={0}>
            <Tag color={workerColor} icon={workerIcon} style={{ fontSize: 11 }}>
              {record.status?.toUpperCase() || 'UNKNOWN'}
            </Tag>
            {record.podStatus && (
              <Tag
                color={
                  record.podStatus === 'Running'
                    ? 'green'
                    : record.podStatus === 'Creating' || record.podStatus === 'Pending'
                    ? 'orange'
                    : record.podStatus === 'Terminating'
                    ? 'volcano'
                    : record.podStatus === 'Failed'
                    ? 'red'
                    : 'default'
                }
                style={{ fontSize: 10 }}
              >
                Pod: {record.podStatus}
              </Tag>
            )}
          </Space>
        );
      },
    },
    {
      title: 'Current Jobs',
      dataIndex: 'current_jobs',
      key: 'current_jobs',
      width: 90,
      render: (currentJobs: number, record: Worker) => (
        <Text>
          {currentJobs || 0} / {record.concurrency || 0}
        </Text>
      ),
    },
    {
      title: 'Last Heartbeat',
      dataIndex: 'last_heartbeat',
      key: 'last_heartbeat',
      width: 110,
      render: (time: string) => {
        if (!time) return '-';
        const heartbeat = new Date(time);
        const now = new Date();
        const diff = now.getTime() - heartbeat.getTime();
        const seconds = Math.floor(diff / 1000);
        if (seconds < 60) return `${seconds}s ago`;
        if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
        return heartbeat.toLocaleString();
      },
    },
    {
      title: 'Idle Time',
      dataIndex: 'last_task_time',
      key: 'last_task_time',
      width: 120,
      render: (lastTaskTime: string, record: Worker) => {
        // If worker has current jobs, it's not idle
        if (record.current_jobs > 0) {
          return <Tag color="blue">Active</Tag>;
        }

        if (!lastTaskTime) {
          // Worker never processed tasks, show time since registration
          const registered = new Date(record.registered_at);
          const now = new Date();
          const diff = now.getTime() - registered.getTime();
          const minutes = Math.floor(diff / 60000);
          if (minutes < 1) return <Tag color="green">Just started</Tag>;
          if (minutes < 60) return <Tag color="orange">{minutes}m (new)</Tag>;
          const hours = Math.floor(minutes / 60);
          if (hours < 24) return <Tag color="volcano">{hours}h (new)</Tag>;
          return <Tag color="red">{Math.floor(hours / 24)}d (new)</Tag>;
        }

        // Calculate idle time since last task
        const lastTask = new Date(lastTaskTime);
        const now = new Date();
        const idleMs = now.getTime() - lastTask.getTime();
        const idleSeconds = Math.floor(idleMs / 1000);

        if (idleSeconds < 60) return <Tag color="green">{idleSeconds}s</Tag>;
        const idleMinutes = Math.floor(idleSeconds / 60);
        if (idleMinutes < 60) return <Tag color="orange">{idleMinutes}m</Tag>;
        const idleHours = Math.floor(idleMinutes / 60);
        if (idleHours < 24) return <Tag color="volcano">{idleHours}h</Tag>;
        const idleDays = Math.floor(idleHours / 24);
        return <Tag color="red">{idleDays}d</Tag>;
      },
    },
    {
      title: 'Version',
      dataIndex: 'version',
      key: 'version',
      width: 80,
      render: (version: string) => version || '-',
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Table
        columns={workerColumns}
        dataSource={workers || []}
        rowKey="id"
        loading={loadingWorkers}
        locale={{ emptyText: <Empty description="No workers found" /> }}
        pagination={{ pageSize: 20 }}
        onRow={(record: WorkerWithPodInfo) => ({
          onClick: () => {
            setSelectedWorker({ endpoint: record.endpoint, workerId: record.id });
            // Use pod_name if available, fallback to worker id
            const podName = record.pod_name || record.id;
            setSelectedPod({ endpoint: record.endpoint, podName });
            setWorkerDrawerTab('tasks');
            setWorkerDrawerVisible(true);
          },
          style: { cursor: 'pointer' },
        })}
      />

      {/* Worker Details Drawer */}
      <Drawer
        title={`Worker: ${selectedWorker?.workerId || '-'}`}
        open={workerDrawerVisible}
        width="80%"
        onClose={() => {
          setWorkerDrawerVisible(false);
          setSelectedWorker(null);
          setSelectedPod(null);
        }}
        destroyOnClose={workerDrawerTab !== 'terminal'}
      >
        <Tabs
          activeKey={workerDrawerTab}
          onChange={(key) => setWorkerDrawerTab(key)}
          items={[
            {
              key: 'tasks',
              label: <Space><HistoryOutlined /> Tasks</Space>,
              children: (
                <div style={{ padding: 16 }}>
                  <Table
                    columns={[
                      {
                        title: 'Task ID',
                        dataIndex: 'id',
                        key: 'id',
                        width: 200,
                        render: (id: string) => (
                          <Tooltip title={id}>
                            <Text code style={{ fontSize: 12 }}>
                              {id.substring(0, 16)}...
                            </Text>
                          </Tooltip>
                        ),
                      },
                      {
                        title: 'Status',
                        dataIndex: 'status',
                        key: 'status',
                        width: 120,
                        render: (status: string) => {
                          const statusLower = status?.toLowerCase();
                          let color = 'default';
                          let icon = <QuestionCircleOutlined />;

                          if (statusLower === 'completed') {
                            color = 'green';
                            icon = <CheckCircleOutlined />;
                          } else if (statusLower === 'failed') {
                            color = 'red';
                            icon = <CloseCircleOutlined />;
                          } else if (statusLower === 'in_progress') {
                            color = 'blue';
                            icon = <LoadingOutlined />;
                          } else if (statusLower === 'pending') {
                            color = 'orange';
                            icon = <SyncOutlined />;
                          }

                          return (
                            <Tag color={color} icon={icon}>
                              {status}
                            </Tag>
                          );
                        },
                      },
                      {
                        title: 'Created At',
                        dataIndex: 'createdAt',
                        key: 'createdAt',
                        width: 160,
                        render: (time: string) => {
                          if (!time) return '-';
                          return new Date(time).toLocaleString();
                        },
                      },
                      {
                        title: 'Execution Time',
                        dataIndex: 'executionTime',
                        key: 'executionTime',
                        width: 120,
                        render: (time: number) => {
                          if (!time) return '-';
                          if (time < 1000) return `${time}ms`;
                          return `${(time / 1000).toFixed(2)}s`;
                        },
                      },
                      {
                        title: 'Delay Time',
                        dataIndex: 'delayTime',
                        key: 'delayTime',
                        width: 120,
                        render: (time: number) => {
                          if (!time) return '-';
                          if (time < 1000) return `${time}ms`;
                          return `${(time / 1000).toFixed(2)}s`;
                        },
                      },
                    ]}
                    dataSource={workerTasks}
                    rowKey="id"
                    loading={loadingWorkerTasks}
                    locale={{ emptyText: <Empty description="No tasks found for this worker" /> }}
                    pagination={{ pageSize: 20 }}
                    size="small"
                  />
                </div>
              ),
            },
            {
              key: 'logs',
              label: <Space><FileTextOutlined /> Logs</Space>,
              children: (
                <div style={{ padding: 16 }}>
                  <Button
                    onClick={() => refetchLogs()}
                    loading={loadingLogs}
                    style={{ marginBottom: 12 }}
                    icon={<SyncOutlined />}
                  >
                    Refresh
                  </Button>
                  <pre
                    style={{
                      background: '#000',
                      color: '#fff',
                      padding: '12px',
                      borderRadius: '4px',
                      maxHeight: '70vh',
                      overflow: 'auto',
                      fontSize: '12px',
                      fontFamily: 'monospace',
                    }}
                  >
                    {logsText || 'No logs available'}
                  </pre>
                </div>
              ),
            },
            {
              key: 'terminal',
              label: <Space><CodeOutlined /> Exec</Space>,
              children: selectedWorker ? (
                <Terminal endpoint={selectedWorker.endpoint} workerId={selectedWorker.workerId} />
              ) : (
                <div style={{ padding: 16 }}>No worker selected</div>
              ),
            },
            {
              key: 'details',
              label: <Space><UnorderedListOutlined /> Details</Space>,
              children: loadingPodDetail ? (
                <div style={{ padding: 24 }}>Loading pod details...</div>
              ) : podDetail ? (
                <Tabs
                  defaultActiveKey="info"
                  items={[
                    {
                      key: 'info',
                      label: 'Info',
                      children: (
                        <Space direction="vertical" size="middle" style={{ width: '100%', padding: 16 }}>
                          {/* Pod Information */}
                          <Card title="Pod Information" size="small">
                            <Descriptions bordered column={2} size="small">
                              <Descriptions.Item label="Name" span={2}>{podDetail.name}</Descriptions.Item>
                              <Descriptions.Item label="Namespace">{podDetail.namespace}</Descriptions.Item>
                              <Descriptions.Item label="UID">{podDetail.uid}</Descriptions.Item>
                              <Descriptions.Item label="Phase">
                                <Tag color={
                                  podDetail.phase === 'Running' ? 'green' :
                                  podDetail.phase === 'Pending' ? 'orange' :
                                  podDetail.phase === 'Succeeded' ? 'blue' :
                                  podDetail.phase === 'Failed' ? 'red' : 'default'
                                }>
                                  {podDetail.phase}
                                </Tag>
                              </Descriptions.Item>
                              <Descriptions.Item label="Status">
                                <Tag>{podDetail.status}</Tag>
                                {podDetail.reason && <Text type="secondary"> ({podDetail.reason})</Text>}
                              </Descriptions.Item>
                              <Descriptions.Item label="Pod IP">{podDetail.ip || '-'}</Descriptions.Item>
                              <Descriptions.Item label="Node">{podDetail.nodeName || '-'}</Descriptions.Item>
                              <Descriptions.Item label="Created At">
                                {new Date(podDetail.createdAt).toLocaleString()}
                              </Descriptions.Item>
                              <Descriptions.Item label="Started At">
                                {podDetail.startedAt ? new Date(podDetail.startedAt).toLocaleString() : '-'}
                              </Descriptions.Item>
                              <Descriptions.Item label="Restart Count">{podDetail.restartCount}</Descriptions.Item>
                            </Descriptions>
                          </Card>

                          {/* Containers */}
                          <Card title={`Containers (${podDetail.containers?.length || 0})`} size="small">
                            {podDetail.containers && podDetail.containers.length > 0 ? (
                              <Table
                                dataSource={podDetail.containers}
                                pagination={false}
                                size="small"
                                rowKey="name"
                                columns={[
                                  {
                                    title: 'Name',
                                    dataIndex: 'name',
                                    key: 'name',
                                    render: (name: string) => <Text strong>{name}</Text>,
                                  },
                                  {
                                    title: 'Image',
                                    dataIndex: 'image',
                                    key: 'image',
                                    ellipsis: true,
                                  },
                                  {
                                    title: 'State',
                                    dataIndex: 'state',
                                    key: 'state',
                                    render: (state: string, record: any) => (
                                      <Space direction="vertical" size={0}>
                                        <Tag color={
                                          state === 'Running' ? 'green' :
                                          state === 'Waiting' ? 'orange' :
                                          state === 'Terminated' ? 'red' : 'default'
                                        }>
                                          {state}
                                        </Tag>
                                        {record.reason && <Text type="secondary" style={{ fontSize: 11 }}>{record.reason}</Text>}
                                      </Space>
                                    ),
                                  },
                                  {
                                    title: 'Ready',
                                    dataIndex: 'ready',
                                    key: 'ready',
                                    render: (ready: boolean) => ready ? (
                                      <CheckCircleOutlined style={{ color: '#52c41a' }} />
                                    ) : (
                                      <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
                                    ),
                                  },
                                  {
                                    title: 'Restarts',
                                    dataIndex: 'restartCount',
                                    key: 'restartCount',
                                  },
                                ]}
                                expandable={{
                                  expandedRowRender: (container: any) => (
                                    <Space direction="vertical" size="small" style={{ width: '100%' }}>
                                      <Descriptions bordered column={2} size="small">
                                        {container.startedAt && (
                                          <Descriptions.Item label="Started At">
                                            {new Date(container.startedAt).toLocaleString()}
                                          </Descriptions.Item>
                                        )}
                                        {container.finishedAt && (
                                          <Descriptions.Item label="Finished At">
                                            {new Date(container.finishedAt).toLocaleString()}
                                          </Descriptions.Item>
                                        )}
                                        {container.exitCode !== undefined && (
                                          <Descriptions.Item label="Exit Code">{container.exitCode}</Descriptions.Item>
                                        )}
                                        {container.message && (
                                          <Descriptions.Item label="Message" span={2}>
                                            <Text type="secondary">{container.message}</Text>
                                          </Descriptions.Item>
                                        )}
                                      </Descriptions>

                                      {/* Resources */}
                                      {container.resources && (
                                        <div>
                                          <Text strong>Resources:</Text>
                                          <Descriptions bordered column={2} size="small" style={{ marginTop: 8 }}>
                                            {container.resources.requests && (
                                              <Descriptions.Item label="Requests" span={2}>
                                                <Space wrap>
                                                  {Object.entries(container.resources.requests).map(([key, value]) => (
                                                    <Tag key={key} color="blue">
                                                      {key}: {String(value)}
                                                    </Tag>
                                                  ))}
                                                </Space>
                                              </Descriptions.Item>
                                            )}
                                            {container.resources.limits && (
                                              <Descriptions.Item label="Limits" span={2}>
                                                <Space wrap>
                                                  {Object.entries(container.resources.limits).map(([key, value]) => (
                                                    <Tag key={key} color="orange">
                                                      {key}: {String(value)}
                                                    </Tag>
                                                  ))}
                                                </Space>
                                              </Descriptions.Item>
                                            )}
                                          </Descriptions>
                                        </div>
                                      )}

                                      {/* Ports */}
                                      {container.ports && container.ports.length > 0 && (
                                        <div>
                                          <Text strong>Ports:</Text>
                                          <div style={{ marginTop: 8 }}>
                                            <Space wrap>
                                              {container.ports.map((port: any, idx: number) => (
                                                <Tag key={idx} color="green">
                                                  {port.name || 'unnamed'}: {port.containerPort}/{port.protocol}
                                                </Tag>
                                              ))}
                                            </Space>
                                          </div>
                                        </div>
                                      )}

                                      {/* Environment Variables */}
                                      {container.env && container.env.length > 0 && (
                                        <div>
                                          <Text strong>Environment Variables ({container.env.length}):</Text>
                                          <Table
                                            dataSource={container.env}
                                            pagination={false}
                                            size="small"
                                            style={{ marginTop: 8 }}
                                            columns={[
                                              {
                                                title: 'Name',
                                                dataIndex: 'name',
                                                key: 'name',
                                                width: '40%',
                                                render: (name: string) => <Text code>{name}</Text>,
                                              },
                                              {
                                                title: 'Value',
                                                dataIndex: 'value',
                                                key: 'value',
                                                ellipsis: true,
                                                render: (value: string) => <Text code>{value}</Text>,
                                              },
                                            ]}
                                          />
                                        </div>
                                      )}
                                    </Space>
                                  ),
                                }}
                              />
                            ) : (
                              <Empty description="No containers" />
                            )}
                          </Card>

                          {/* Conditions */}
                          <Card title="Conditions" size="small">
                            {podDetail.conditions && podDetail.conditions.length > 0 ? (
                              <Table
                                dataSource={podDetail.conditions}
                                pagination={false}
                                size="small"
                                rowKey="type"
                                columns={[
                                  {
                                    title: 'Type',
                                    dataIndex: 'type',
                                    key: 'type',
                                  },
                                  {
                                    title: 'Status',
                                    dataIndex: 'status',
                                    key: 'status',
                                    render: (status: string) => (
                                      <Tag color={status === 'True' ? 'green' : status === 'False' ? 'red' : 'default'}>
                                        {status}
                                      </Tag>
                                    ),
                                  },
                                  {
                                    title: 'Reason',
                                    dataIndex: 'reason',
                                    key: 'reason',
                                    render: (reason?: string) => reason || '-',
                                  },
                                  {
                                    title: 'Message',
                                    dataIndex: 'message',
                                    key: 'message',
                                    ellipsis: true,
                                    render: (message?: string) => message || '-',
                                  },
                                  {
                                    title: 'Last Transition',
                                    dataIndex: 'lastTransitionTime',
                                    key: 'lastTransitionTime',
                                    render: (time?: string) => time ? new Date(time).toLocaleString() : '-',
                                  },
                                ]}
                              />
                            ) : (
                              <Empty description="No conditions" />
                            )}
                          </Card>

                          {/* Events */}
                          <Card title={`Events (${podDetail.events?.length || 0})`} size="small">
                            {podDetail.events && podDetail.events.length > 0 ? (
                              <Timeline
                                items={podDetail.events.slice(0, 20).map((event) => ({
                                  color: event.type === 'Warning' ? 'red' : 'blue',
                                  children: (
                                    <Space direction="vertical" size={0}>
                                      <Space>
                                        <Tag color={event.type === 'Warning' ? 'red' : 'blue'}>{event.type}</Tag>
                                        <Text strong>{event.reason}</Text>
                                        {event.count > 1 && <Text type="secondary">(x{event.count})</Text>}
                                      </Space>
                                      <Text>{event.message}</Text>
                                      <Text type="secondary" style={{ fontSize: 11 }}>
                                        Last seen: {new Date(event.lastSeen).toLocaleString()}
                                      </Text>
                                    </Space>
                                  ),
                                }))}
                              />
                            ) : (
                              <Empty description="No events" />
                            )}
                          </Card>

                          {/* Tolerations */}
                          {podDetail.tolerations && podDetail.tolerations.length > 0 && (
                            <Card title={`Tolerations (${podDetail.tolerations.length})`} size="small">
                              <Table
                                dataSource={podDetail.tolerations}
                                pagination={false}
                                size="small"
                                rowKey={(record, idx) => `${record.key}-${idx}`}
                                columns={[
                                  {
                                    title: 'Key',
                                    dataIndex: 'key',
                                    key: 'key',
                                    render: (key: string) => <Text code>{key}</Text>,
                                  },
                                  {
                                    title: 'Operator',
                                    dataIndex: 'operator',
                                    key: 'operator',
                                    render: (op: string) => <Tag color="blue">{op}</Tag>,
                                  },
                                  {
                                    title: 'Value',
                                    dataIndex: 'value',
                                    key: 'value',
                                    render: (value?: string) => value ? <Text code>{value}</Text> : '-',
                                  },
                                  {
                                    title: 'Effect',
                                    dataIndex: 'effect',
                                    key: 'effect',
                                    render: (effect: string) => <Tag color="orange">{effect}</Tag>,
                                  },
                                ]}
                              />
                            </Card>
                          )}

                          {/* Volumes */}
                          {podDetail.volumes && podDetail.volumes.length > 0 && (
                            <Card title={`Volumes (${podDetail.volumes.length})`} size="small">
                              <Table
                                dataSource={podDetail.volumes}
                                pagination={false}
                                size="small"
                                rowKey="name"
                                columns={[
                                  {
                                    title: 'Name',
                                    dataIndex: 'name',
                                    key: 'name',
                                  },
                                  {
                                    title: 'Type',
                                    dataIndex: 'type',
                                    key: 'type',
                                    render: (type: string) => <Tag>{type}</Tag>,
                                  },
                                  {
                                    title: 'Source',
                                    dataIndex: 'source',
                                    key: 'source',
                                    render: (source?: Record<string, any>) =>
                                      source ? <Text code>{JSON.stringify(source)}</Text> : '-',
                                    ellipsis: true,
                                  },
                                ]}
                              />
                            </Card>
                          )}
                        </Space>
                      ),
                    },
                    {
                      key: 'yaml',
                      label: 'YAML',
                      children: (
                        <div style={{ height: 'calc(100vh - 250px)', overflow: 'auto', padding: 16 }}>
                          <pre style={{
                            backgroundColor: '#f5f5f5',
                            padding: 16,
                            borderRadius: 4,
                            fontSize: 12,
                            fontFamily: 'Monaco, Consolas, monospace',
                            whiteSpace: 'pre-wrap',
                            wordBreak: 'break-word'
                          }}>
                            {JSON.stringify(podDetail, null, 2)}
                          </pre>
                        </div>
                      ),
                    },
                  ]}
                />
              ) : (
                <div style={{ padding: 24 }}>No pod details available</div>
              ),
            },
          ]}
        />
      </Drawer>
    </div>
  );
};

export default WorkersTab;
