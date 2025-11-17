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

const TasksTab = ({ endpoint, onJumpToWorker, workers = [] }: TasksTabProps) => {
  const [taskDetail, setTaskDetail] = useState<Task | null>(null);
  const [taskDrawerVisible, setTaskDrawerVisible] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string | undefined>(undefined);
  const [taskIdSearch, setTaskIdSearch] = useState<string>('');

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

  // Filter tasks by search term (frontend filter as backup)
  const tasks = (tasksResponse?.tasks || []).filter((task) => {
    if (taskIdSearch && !task.id.toLowerCase().includes(taskIdSearch.toLowerCase())) {
      return false;
    }
    return true;
  });

  // Fetch task events when task is selected
  const { data: _eventsData } = useQuery({
    queryKey: ['task-events', taskDetail?.id],
    queryFn: async () => {
      if (!taskDetail?.id) return null;
      const response = await api.tasks.getEvents(taskDetail.id);
      return response.data;
    },
    enabled: !!taskDetail?.id && taskDrawerVisible,
  });

  // Fetch task timeline
  const { data: timelineData } = useQuery({
    queryKey: ['task-timeline', taskDetail?.id],
    queryFn: async () => {
      if (!taskDetail?.id) return null;
      const response = await api.tasks.getTimeline(taskDetail.id);
      return response.data;
    },
    enabled: !!taskDetail?.id && taskDrawerVisible,
  });

  // Fetch task execution history
  const { data: executionHistory } = useQuery({
    queryKey: ['task-execution-history', taskDetail?.id],
    queryFn: async () => {
      if (!taskDetail?.id) return null;
      const response = await api.tasks.getExecutionHistory(taskDetail.id);
      return response.data;
    },
    enabled: !!taskDetail?.id && taskDrawerVisible,
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
            setTaskDetail(task);
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
        title={`Task Details: ${taskDetail?.id?.substring(0, 16) || '-'}...`}
        open={taskDrawerVisible}
        width="60%"
        onClose={() => {
          setTaskDrawerVisible(false);
          setTaskDetail(null);
        }}
      >
        {taskDetail && (
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            {/* Basic Info */}
            <Descriptions title="Basic Information" bordered column={1} size="small">
              <Descriptions.Item label="Task ID">
                <Text code>{taskDetail.id}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Status">
                <Tag color={getStatusColor(taskDetail.status)} icon={getStatusIcon(taskDetail.status)}>
                  {taskDetail.status}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Endpoint">{taskDetail.endpoint || '-'}</Descriptions.Item>
              <Descriptions.Item label="Worker ID">
                {taskDetail.workerId ? <Text code>{taskDetail.workerId}</Text> : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Created At">
                {taskDetail.createdAt ? new Date(taskDetail.createdAt).toLocaleString() : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Execution Time">
                {taskDetail.executionTime ? `${taskDetail.executionTime}ms` : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Delay Time">
                {taskDetail.delayTime ? `${taskDetail.delayTime}ms` : '-'}
              </Descriptions.Item>
            </Descriptions>

            {/* Input */}
            {taskDetail.input && (
              <div>
                <Text strong>Input:</Text>
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
                  {JSON.stringify(taskDetail.input, null, 2)}
                </pre>
              </div>
            )}

            {/* Output */}
            {taskDetail.output && (
              <div>
                <Text strong>Output:</Text>
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
                  {JSON.stringify(taskDetail.output, null, 2)}
                </pre>
              </div>
            )}

            {/* Error */}
            {taskDetail.error && (
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
                  }}
                >
                  {taskDetail.error}
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
