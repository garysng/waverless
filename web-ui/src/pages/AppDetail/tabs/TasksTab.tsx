import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Button,
  Col,
  Descriptions,
  Drawer,
  Empty,
  Input,
  Row,
  Select,
  Space,
  Statistic,
  Table,
  Tag,
  Timeline,
  Tooltip,
  Typography,
} from 'antd';
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  LoadingOutlined,
  SyncOutlined,
} from '@ant-design/icons';
import { api } from '@/api/client';
import type { Task, TaskListResponse, TaskEvent, ExecutionRecord, WorkerWithPodInfo } from '@/types';

const { Text } = Typography;

interface TasksTabProps {
  endpoint: string;
  onJumpToWorker?: (workerId: string) => void;
  workers?: WorkerWithPodInfo[];
}

// Helper function to check if JSON data is too large
const isDataTooLarge = (data: any, maxSizeKB = 50): boolean => {
  if (!data) return false;
  try {
    const jsonString = JSON.stringify(data);
    const sizeKB = new Blob([jsonString]).size / 1024;
    return sizeKB > maxSizeKB;
  } catch {
    return false;
  }
};

// Helper function to format data size
const formatDataSize = (data: any): string => {
  if (!data) return '0 KB';
  try {
    const jsonString = JSON.stringify(data);
    const sizeKB = new Blob([jsonString]).size / 1024;
    if (sizeKB < 1) {
      return `${(sizeKB * 1024).toFixed(0)} B`;
    } else if (sizeKB < 1024) {
      return `${sizeKB.toFixed(2)} KB`;
    } else {
      return `${(sizeKB / 1024).toFixed(2)} MB`;
    }
  } catch {
    return 'Unknown';
  }
};

const TasksTab = ({ endpoint, onJumpToWorker, workers = [] }: TasksTabProps) => {
  const [taskDrawerVisible, setTaskDrawerVisible] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string | undefined>(undefined);
  const [taskIdSearch, setTaskIdSearch] = useState<string>('');
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);

  // Fetch tasks
  const { data: tasksResponse, isLoading: loadingTasks } = useQuery<TaskListResponse>({
    queryKey: ['tasks', endpoint, statusFilter, taskIdSearch],
    queryFn: async () => {
      const response = await api.tasks.list({
        endpoint,
        status: statusFilter,
        task_id: taskIdSearch || undefined,
        limit: 100,
      });
      return response.data;
    },
    enabled: !!endpoint,
    refetchInterval: taskIdSearch ? false : 5000, // Disable auto-refresh when searching
    placeholderData: (previousData) => previousData,
  });

  // Fetch full task details when task is selected (includes input field)
  const { data: fullTaskDetail } = useQuery({
    queryKey: ['task-detail', selectedTaskId],
    queryFn: async () => {
      if (!selectedTaskId) return null;
      const response = await api.tasks.get(selectedTaskId);
      return response.data;
    },
    enabled: !!selectedTaskId && taskDrawerVisible,
  });

  // Filter tasks by search term (frontend filter as backup)
  const tasks = (tasksResponse?.tasks || []).filter((task) => {
    if (taskIdSearch && !task.id.toLowerCase().includes(taskIdSearch.toLowerCase())) {
      return false;
    }
    return true;
  });

  // Fetch task events when task is selected
  const { data: _eventsData } = useQuery({
    queryKey: ['task-events', selectedTaskId],
    queryFn: async () => {
      if (!selectedTaskId) return null;
      const response = await api.tasks.getEvents(selectedTaskId);
      return response.data;
    },
    enabled: !!selectedTaskId && taskDrawerVisible,
  });

  // Fetch task timeline
  const { data: timelineData } = useQuery({
    queryKey: ['task-timeline', selectedTaskId],
    queryFn: async () => {
      if (!selectedTaskId) return null;
      const response = await api.tasks.getTimeline(selectedTaskId);
      return response.data;
    },
    enabled: !!selectedTaskId && taskDrawerVisible,
  });

  // Fetch task execution history
  const { data: executionHistory } = useQuery({
    queryKey: ['task-execution-history', selectedTaskId],
    queryFn: async () => {
      if (!selectedTaskId) return null;
      const response = await api.tasks.getExecutionHistory(selectedTaskId);
      return response.data;
    },
    enabled: !!selectedTaskId && taskDrawerVisible,
  });

  // Calculate statistics
  const taskStats = useMemo(() => {
    const stats = {
      total: tasks.length,
      pending: 0,
      inProgress: 0,
      completed: 0,
      failed: 0,
      cancelled: 0,
    };

    tasks.forEach((task) => {
      if (task.status === 'PENDING') stats.pending++;
      else if (task.status === 'IN_PROGRESS') stats.inProgress++;
      else if (task.status === 'COMPLETED') stats.completed++;
      else if (task.status === 'FAILED') stats.failed++;
      else if (task.status === 'CANCELLED') stats.cancelled++;
    });

    return stats;
  }, [tasks]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'COMPLETED':
        return 'success';
      case 'IN_PROGRESS':
        return 'processing';
      case 'PENDING':
        return 'default';
      case 'FAILED':
        return 'error';
      case 'CANCELLED':
        return 'warning';
      default:
        return 'default';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'COMPLETED':
        return <CheckCircleOutlined />;
      case 'IN_PROGRESS':
        return <SyncOutlined spin />;
      case 'PENDING':
        return <ClockCircleOutlined />;
      case 'FAILED':
        return <CloseCircleOutlined />;
      case 'CANCELLED':
        return <LoadingOutlined />;
      default:
        return null;
    }
  };

  const taskColumns = [
    {
      title: 'Task ID',
      dataIndex: 'id',
      key: 'id',
      render: (id: string) => (
        <Tooltip title={id}>
          <Text code style={{ fontSize: 12 }}>
            {id.substring(0, 10)}...
          </Text>
        </Tooltip>
      ),
      ellipsis: true,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const colorMap: Record<string, string> = {
          PENDING: 'orange',
          IN_PROGRESS: 'blue',
          COMPLETED: 'green',
          FAILED: 'red',
          CANCELLED: 'default',
        };
        return <Tag color={colorMap[status] || 'default'}>{status}</Tag>;
      },
    },
    {
      title: 'Worker',
      dataIndex: 'workerId',
      key: 'workerId',
      render: (workerId: string | undefined) => {
        if (!workerId) return '-';

        // Check if worker is online or busy (active)
        const worker = workers.find((w) => w.id === workerId);
        const isWorkerActive = worker && (
          worker.status?.toLowerCase() === 'online' ||
          worker.status?.toLowerCase() === 'busy' ||
          worker.podStatus === 'Running'
        );

        return (
          <Tooltip title={isWorkerActive ? `${workerId} (click to view)` : `${workerId} (offline)`}>
            <Text
              code
              style={{
                fontSize: 12,
                cursor: isWorkerActive && onJumpToWorker ? 'pointer' : 'default',
                color: isWorkerActive ? '#1890ff' : undefined,
              }}
              onClick={(e) => {
                if (isWorkerActive && onJumpToWorker) {
                  e.stopPropagation();
                  onJumpToWorker(workerId);
                }
              }}
            >
              {workerId.substring(0, 12)}...
            </Text>
          </Tooltip>
        );
      },
      ellipsis: true,
    },
    {
      title: 'Created At',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (time: string) => (time ? new Date(time).toLocaleString() : '-'),
    },
    {
      title: 'Execution Time',
      dataIndex: 'executionTime',
      key: 'executionTime',
      render: (time: number | undefined) => (time ? `${time}ms` : '-'),
    },
    {
      title: 'Details',
      key: 'details',
      render: (_: any, task: Task) => (
        <Button
          size="small"
          onClick={() => {
            setSelectedTaskId(task.id);
            setTaskDrawerVisible(true);
          }}
        >
          View
        </Button>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={12} md={4}>
          <Statistic title="Total" value={taskStats.total} />
        </Col>
        <Col xs={12} md={4}>
          <Statistic title="Pending" value={taskStats.pending} valueStyle={{ color: '#faad14' }} />
        </Col>
        <Col xs={12} md={4}>
          <Statistic title="In Progress" value={taskStats.inProgress} valueStyle={{ color: '#1890ff' }} />
        </Col>
        <Col xs={12} md={4}>
          <Statistic title="Completed" value={taskStats.completed} valueStyle={{ color: '#52c41a' }} />
        </Col>
        <Col xs={12} md={4}>
          <Statistic title="Failed" value={taskStats.failed} valueStyle={{ color: '#ff4d4f' }} />
        </Col>
        <Col xs={12} md={4}>
          <Statistic title="Cancelled" value={taskStats.cancelled} />
        </Col>
      </Row>

      {/* Filter */}
      <div style={{ marginBottom: 16 }}>
        <Space>
          <Select
            placeholder="Filter by status"
            style={{ width: 200 }}
            allowClear
            value={statusFilter}
            onChange={(value) => setStatusFilter(value)}
            options={[
              { label: 'All', value: undefined },
              { label: 'Pending', value: 'PENDING' },
              { label: 'In Progress', value: 'IN_PROGRESS' },
              { label: 'Completed', value: 'COMPLETED' },
              { label: 'Failed', value: 'FAILED' },
              { label: 'Cancelled', value: 'CANCELLED' },
            ]}
          />
          <Input
            placeholder="Search by Task ID"
            style={{ width: 250 }}
            allowClear
            value={taskIdSearch}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setTaskIdSearch(e.target.value)}
          />
        </Space>
      </div>

      {/* Tasks Table */}
      <Table
        columns={taskColumns}
        dataSource={tasks}
        rowKey="id"
        loading={loadingTasks}
        locale={{ emptyText: <Empty description="No tasks found" /> }}
        pagination={{ pageSize: 20 }}
      />

      {/* Task Detail Drawer */}
      <Drawer
        title={`Task Details: ${fullTaskDetail?.id?.substring(0, 16) || selectedTaskId?.substring(0, 16) || '-'}...`}
        open={taskDrawerVisible}
        width="60%"
        onClose={() => {
          setTaskDrawerVisible(false);
          setSelectedTaskId(null);
        }}
      >
        {fullTaskDetail && (
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            {/* Basic Info */}
            <Descriptions title="Basic Information" bordered column={1} size="small">
              <Descriptions.Item label="Task ID">
                <Text code>{fullTaskDetail.id}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Status">
                <Tag color={getStatusColor(fullTaskDetail.status)} icon={getStatusIcon(fullTaskDetail.status)}>
                  {fullTaskDetail.status}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Endpoint">{fullTaskDetail.endpoint || '-'}</Descriptions.Item>
              <Descriptions.Item label="Worker ID">
                {fullTaskDetail.workerId ? <Text code>{fullTaskDetail.workerId}</Text> : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Created At">
                {fullTaskDetail.createdAt ? new Date(fullTaskDetail.createdAt).toLocaleString() : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Execution Time">
                {fullTaskDetail.executionTime ? `${fullTaskDetail.executionTime}ms` : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Delay Time">
                {fullTaskDetail.delayTime ? `${fullTaskDetail.delayTime}ms` : '-'}
              </Descriptions.Item>
            </Descriptions>

            {/* Input */}
            {fullTaskDetail.input && (
              <div>
                <Space>
                  <Text strong>Input:</Text>
                  <Tag color="blue">{formatDataSize(fullTaskDetail.input)}</Tag>
                </Space>
                {isDataTooLarge(fullTaskDetail.input) ? (
                  <div
                    style={{
                      background: '#fff7e6',
                      padding: '12px',
                      borderRadius: '4px',
                      marginTop: '8px',
                      border: '1px solid #ffd591',
                    }}
                  >
                    <Text type="warning">
                      ⚠️ Input data is too large ({formatDataSize(fullTaskDetail.input)}) and has been omitted for performance reasons.
                    </Text>
                  </div>
                ) : (
                  <pre
                    style={{
                      background: '#f5f5f5',
                      padding: '12px',
                      borderRadius: '4px',
                      marginTop: '8px',
                      maxHeight: '200px',
                      overflow: 'auto',
                    }}
                  >
                    {JSON.stringify(fullTaskDetail.input, null, 2)}
                  </pre>
                )}
              </div>
            )}

            {/* Output */}
            {fullTaskDetail.output && (
              <div>
                <Space>
                  <Text strong>Output:</Text>
                  <Tag color="green">{formatDataSize(fullTaskDetail.output)}</Tag>
                </Space>
                {isDataTooLarge(fullTaskDetail.output) ? (
                  <div
                    style={{
                      background: '#fff7e6',
                      padding: '12px',
                      borderRadius: '4px',
                      marginTop: '8px',
                      border: '1px solid #ffd591',
                    }}
                  >
                    <Text type="warning">
                      ⚠️ Output data is too large ({formatDataSize(fullTaskDetail.output)}) and has been omitted for performance reasons.
                    </Text>
                  </div>
                ) : (
                  <pre
                    style={{
                      background: '#f5f5f5',
                      padding: '12px',
                      borderRadius: '4px',
                      marginTop: '8px',
                      maxHeight: '200px',
                      overflow: 'auto',
                    }}
                  >
                    {JSON.stringify(fullTaskDetail.output, null, 2)}
                  </pre>
                )}
              </div>
            )}

            {/* Error */}
            {fullTaskDetail.error && (
              <div>
                <Text strong type="danger">
                  Error:
                </Text>
                <pre
                  style={{
                    background: '#fff2f0',
                    padding: '12px',
                    borderRadius: '4px',
                    marginTop: '8px',
                    color: '#cf1322',
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word',
                    overflowWrap: 'break-word',
                  }}
                >
                  {fullTaskDetail.error}
                </pre>
              </div>
            )}

            {/* Timeline */}
            {timelineData && timelineData.timeline && timelineData.timeline.length > 0 && (
              <div>
                <Text strong>Timeline:</Text>
                <Timeline
                  style={{ marginTop: 16 }}
                  items={timelineData.timeline.map((event: TaskEvent) => ({
                    color:
                      event.event_type === 'STATUS_CHANGE' && event.to_status === 'COMPLETED'
                        ? 'green'
                        : event.event_type === 'STATUS_CHANGE' && event.to_status === 'FAILED'
                        ? 'red'
                        : 'blue',
                    children: (
                      <div>
                        <div>
                          <Text strong>{event.event_type}</Text>
                          {event.from_status && event.to_status && (
                            <Text type="secondary"> ({event.from_status} → {event.to_status})</Text>
                          )}
                        </div>
                        <div>
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            {new Date(event.event_time).toLocaleString()}
                          </Text>
                        </div>
                        {event.worker_id && (
                          <div>
                            <Text code style={{ fontSize: 11 }}>
                              Worker: {event.worker_id.substring(0, 16)}...
                            </Text>
                          </div>
                        )}
                        {event.error_message && (
                          <div>
                            <Text type="danger" style={{ fontSize: 12 }}>
                              {event.error_message}
                            </Text>
                          </div>
                        )}
                      </div>
                    ),
                  }))}
                />
              </div>
            )}

            {/* Execution History */}
            {executionHistory && executionHistory.history && executionHistory.history.length > 0 && (
              <div>
                <Text strong>Execution History:</Text>
                <Table
                  dataSource={executionHistory.history}
                  rowKey={(record: ExecutionRecord) => `${record.worker_id}-${record.start_time}`}
                  columns={[
                    {
                      title: 'Worker ID',
                      dataIndex: 'worker_id',
                      key: 'worker_id',
                      render: (id: string) => <Text code>{id.substring(0, 16)}...</Text>,
                    },
                    {
                      title: 'Start Time',
                      dataIndex: 'start_time',
                      key: 'start_time',
                      render: (time: string) => new Date(time).toLocaleString(),
                    },
                    {
                      title: 'End Time',
                      dataIndex: 'end_time',
                      key: 'end_time',
                      render: (time: string | undefined) => (time ? new Date(time).toLocaleString() : '-'),
                    },
                    {
                      title: 'Duration',
                      dataIndex: 'duration_seconds',
                      key: 'duration_seconds',
                      render: (duration: number | undefined) => (duration ? `${duration}s` : '-'),
                    },
                  ]}
                  pagination={false}
                  size="small"
                  style={{ marginTop: 8 }}
                />
              </div>
            )}
          </Space>
        )}
      </Drawer>
    </div>
  );
};

export default TasksTab;
