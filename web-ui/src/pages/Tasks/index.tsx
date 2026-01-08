import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Input, Select, Modal, message, Popconfirm, Tooltip, Timeline } from 'antd';
import {
  SearchOutlined,
  StopOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  SyncOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import { api } from '@/api/client';

const TasksPage = () => {
  const [searchInput, setSearchInput] = useState('');
  const [statusInput, setStatusInput] = useState('all');
  const [endpointInput, setEndpointInput] = useState('all');
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [endpointFilter, setEndpointFilter] = useState('all');
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [selectedTask, setSelectedTask] = useState<any>(null);
  const [detailVisible, setDetailVisible] = useState(false);

  const { data: endpoints } = useQuery({
    queryKey: ['endpoints'],
    queryFn: async () => {
      const res = await api.endpoints.list();
      return Array.isArray(res.data) ? res.data : (res.data?.endpoints || []);
    },
  });

  const { data: summary } = useQuery({
    queryKey: ['tasks-summary'],
    queryFn: async () => (await api.get('/api/v1/statistics/overview')).data,
    refetchInterval: 10000,
  });

  const { data: tasksData, isLoading, refetch } = useQuery({
    queryKey: ['tasks', statusFilter, endpointFilter, search, page, pageSize],
    queryFn: async () => {
      const params: any = { limit: pageSize, offset: page * pageSize };
      if (statusFilter !== 'all') params.status = statusFilter;
      if (endpointFilter !== 'all') params.endpoint = endpointFilter;
      if (search) params.task_id = search;
      return (await api.tasks.list(params)).data;
    },
  });

  const handleSearch = () => {
    setSearch(searchInput);
    setStatusFilter(statusInput);
    setEndpointFilter(endpointInput);
    setPage(0);
  };

  const handleReset = () => {
    setSearchInput('');
    setStatusInput('all');
    setEndpointInput('all');
    setSearch('');
    setStatusFilter('all');
    setEndpointFilter('all');
    setPage(0);
  };

  const cancelMutation = useMutation({
    mutationFn: (taskId: string) => api.tasks.cancel(taskId),
    onSuccess: () => { message.success('Task cancelled'); refetch(); },
    onError: (e: any) => message.error(e.response?.data?.error || 'Failed'),
  });

  const tasks = tasksData?.tasks || [];
  const total = tasksData?.total || 0;

  const getStatusClass = (s: string) => s === 'COMPLETED' ? 'success' : s === 'FAILED' ? 'failed' : s === 'IN_PROGRESS' ? 'running' : 'pending';
  const getStatusIcon = (s: string) => {
    switch (s) {
      case 'COMPLETED': return <CheckCircleOutlined style={{ color: '#48bb78' }} />;
      case 'IN_PROGRESS': return <SyncOutlined spin style={{ color: '#1da1f2' }} />;
      case 'PENDING': return <ClockCircleOutlined style={{ color: '#f59e0b' }} />;
      case 'FAILED': return <CloseCircleOutlined style={{ color: '#f56565' }} />;
      default: return null;
    }
  };

  return (
    <>
      {/* Stats */}
      <div className="stats-grid mb-4" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
        <div className="stat-card">
          <div className="stat-icon green"><CheckCircleOutlined /></div>
          <div className="stat-content">
            <div className="stat-label">Completed</div>
            <div className="stat-value">{summary?.completed || 0}</div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon blue"><SyncOutlined /></div>
          <div className="stat-content">
            <div className="stat-label">In Progress</div>
            <div className="stat-value">{summary?.in_progress || 0}</div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon orange"><ClockCircleOutlined /></div>
          <div className="stat-content">
            <div className="stat-label">Pending</div>
            <div className="stat-value">{summary?.pending || 0}</div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon" style={{ background: 'rgba(245,101,101,0.1)', color: '#f56565' }}>
            <CloseCircleOutlined />
          </div>
          <div className="stat-content">
            <div className="stat-label">Failed</div>
            <div className="stat-value">{summary?.failed || 0}</div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex justify-between items-center mb-4">
        <div className="filters">
          <Input placeholder="Search task ID..." prefix={<SearchOutlined />} value={searchInput} onChange={e => setSearchInput(e.target.value)} onPressEnter={handleSearch} style={{ width: 240 }} />
          <Select value={statusInput} onChange={setStatusInput} style={{ width: 140 }} options={[
            { value: 'all', label: 'All Status' },
            { value: 'PENDING', label: 'Pending' },
            { value: 'IN_PROGRESS', label: 'In Progress' },
            { value: 'COMPLETED', label: 'Completed' },
            { value: 'FAILED', label: 'Failed' },
          ]} />
          <Select value={endpointInput} onChange={setEndpointInput} style={{ width: 180 }} options={[
            { value: 'all', label: 'All Endpoints' },
            ...(endpoints?.map((ep: any) => ({ value: ep.name, label: ep.name })) || []),
          ]} />
          <button className="btn btn-blue" onClick={handleSearch}><SearchOutlined /> Search</button>
          <button className="btn btn-outline" onClick={handleReset}>Reset</button>
        </div>
        <span style={{ color: '#6b7280', fontSize: 13 }}>Total: {total}</span>
      </div>

      {/* Table */}
      <div className="card">
        <div className="table-container">
          <table>
            <thead>
              <tr>
                <th>Task ID</th>
                <th>Endpoint</th>
                <th>Status</th>
                <th>Worker</th>
                <th>Created</th>
                <th>Exec Time</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr><td colSpan={7}><div className="loading"><div className="spinner"></div></div></td></tr>
              ) : tasks.length === 0 ? (
                <tr><td colSpan={7} style={{ textAlign: 'center', color: '#6b7280', padding: 40 }}>No tasks found</td></tr>
              ) : (
                tasks.map((task: any) => (
                  <tr key={task.id}>
                    <td><Tooltip title={task.id}><span style={{ fontFamily: 'monospace', fontSize: 11 }}>{task.id.substring(0, 20)}...</span></Tooltip></td>
                    <td>{task.endpoint}</td>
                    <td><span className={`tag ${getStatusClass(task.status)}`}>{getStatusIcon(task.status)} {task.status}</span></td>
                    <td>{task.workerId ? <Tooltip title={task.workerId}><span style={{ fontSize: 12, color: '#6b7280' }}>{task.workerId.substring(0, 12)}...</span></Tooltip> : '-'}</td>
                    <td style={{ fontSize: 12 }}>{task.createdAt ? new Date(task.createdAt).toLocaleString() : '-'}</td>
                    <td style={{ fontSize: 12 }}>{task.executionTime ? `${(task.executionTime/1000).toFixed(2)}s` : '-'}</td>
                    <td>
                      <div className="flex gap-2">
                        <button className="btn btn-sm btn-outline" onClick={() => { setSelectedTask(task); setDetailVisible(true); }}><EyeOutlined /></button>
                        {(task.status === 'PENDING' || task.status === 'IN_PROGRESS') && (
                          <Popconfirm title="Cancel this task?" onConfirm={() => cancelMutation.mutate(task.id)}>
                            <button className="btn btn-sm btn-outline" style={{ color: '#f56565' }}><StopOutlined /></button>
                          </Popconfirm>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        {/* Pagination */}
        <div className="flex justify-between items-center" style={{ padding: '12px 16px', borderTop: '1px solid #f3f4f6' }}>
          <div className="flex gap-2 items-center">
            <span style={{ fontSize: 13, color: '#6b7280' }}>Show</span>
            <Select value={pageSize} onChange={v => { setPageSize(v); setPage(0); }} size="small" style={{ width: 70 }} options={[
              { value: 10, label: '10' },
              { value: 20, label: '20' },
              { value: 50, label: '50' },
              { value: 100, label: '100' },
            ]} />
            <span style={{ fontSize: 13, color: '#6b7280' }}>/ page</span>
          </div>
          <div className="flex gap-2 items-center">
            <span style={{ fontSize: 13, color: '#6b7280' }}>Page {page + 1} of {Math.ceil(total / pageSize) || 1}</span>
            <button className="btn btn-sm btn-outline" disabled={page === 0} onClick={() => setPage(p => p - 1)}>Prev</button>
            <button className="btn btn-sm btn-outline" disabled={(page + 1) * pageSize >= total} onClick={() => setPage(p => p + 1)}>Next</button>
          </div>
        </div>
      </div>

      {/* Detail Modal */}
      <Modal title="Task Details" open={detailVisible} onCancel={() => setDetailVisible(false)} footer={null} width={900}>
        {selectedTask && <TaskDetailContent task={selectedTask} getStatusClass={getStatusClass} />}
      </Modal>
    </>
  );
};

// Task Detail Content with Timeline and Execution History
const TaskDetailContent = ({ task, getStatusClass }: { task: any; getStatusClass: (s: string) => string }) => {
  const { data: fullTask } = useQuery({
    queryKey: ['task-detail', task.id],
    queryFn: async () => (await api.tasks.get(task.id)).data,
  });
  const { data: timeline } = useQuery({
    queryKey: ['task-timeline', task.id],
    queryFn: async () => (await api.tasks.getTimeline(task.id)).data,
  });
  const { data: execHistory } = useQuery({
    queryKey: ['task-exec-history', task.id],
    queryFn: async () => (await api.tasks.getExecutionHistory(task.id)).data,
  });

  const t = fullTask || task;

  return (
    <div>
      <div className="form-row">
        <div className="form-group"><label className="form-label">Task ID</label><div style={{ fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-all' }}>{t.id}</div></div>
        <div className="form-group"><label className="form-label">Status</label><span className={`tag ${getStatusClass(t.status)}`}>{t.status}</span></div>
      </div>
      <div className="form-row">
        <div className="form-group"><label className="form-label">Endpoint</label><div>{t.endpoint}</div></div>
        <div className="form-group"><label className="form-label">Worker</label><div style={{ fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-all' }}>{t.worker_id || t.workerId || '-'}</div></div>
      </div>
      <div className="form-row">
        <div className="form-group"><label className="form-label">Created</label><div>{(t.created_at || t.createdAt) ? new Date(t.created_at || t.createdAt).toLocaleString() : '-'}</div></div>
        <div className="form-group"><label className="form-label">Execution Time</label><div>{(t.execution_duration_ms || t.executionTime) ? `${((t.execution_duration_ms || t.executionTime)/1000).toFixed(2)}s` : '-'}</div></div>
      </div>

      {/* Input */}
      <div style={{ marginTop: 16 }}>
        <label className="form-label">Input</label>
        <pre style={{ background: '#f9fafb', padding: 12, borderRadius: 6, fontSize: 12, overflow: 'auto', maxHeight: 200, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{t.input ? JSON.stringify(t.input, null, 2) : '-'}</pre>
      </div>

      {/* Output */}
      {t.output && (
        <div style={{ marginTop: 16 }}>
          <label className="form-label">Output</label>
          <pre style={{ background: '#f9fafb', padding: 12, borderRadius: 6, fontSize: 12, overflow: 'auto', maxHeight: 300, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{JSON.stringify(t.output, null, 2)}</pre>
        </div>
      )}

      {/* Error */}
      {t.error && (
        <div style={{ marginTop: 16 }}>
          <label className="form-label">Error</label>
          <pre style={{ background: '#fef2f2', padding: 12, borderRadius: 6, fontSize: 12, color: '#f56565', whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{t.error}</pre>
        </div>
      )}

      {/* Timeline */}
      {timeline?.timeline && timeline.timeline.length > 0 && (
        <div style={{ marginTop: 20 }}>
          <label className="form-label">Timeline</label>
          <Timeline style={{ marginTop: 12 }} items={timeline.timeline.map((e: any) => ({
            color: e.to_status === 'COMPLETED' ? 'green' : e.to_status === 'FAILED' ? 'red' : e.to_status === 'IN_PROGRESS' ? 'blue' : 'gray',
            children: (
              <div>
                <div><strong>{e.event_type}</strong>{e.from_status && e.to_status && <span style={{ color: '#6b7280' }}> ({e.from_status} → {e.to_status})</span>}</div>
                <div style={{ fontSize: 12, color: '#6b7280' }}>{new Date(e.event_time).toLocaleString()}</div>
                {e.worker_id && <div style={{ fontSize: 11 }}><code>Worker: {e.worker_id.substring(0, 20)}...</code></div>}
                {e.worker_pod_name && <div style={{ fontSize: 11 }}><code>Pod: {e.worker_pod_name}</code></div>}
                {e.error_message && <div style={{ fontSize: 12, color: '#f56565' }}>{e.error_message}</div>}
              </div>
            ),
          }))} />
        </div>
      )}

      {/* Execution History */}
      {execHistory?.history && execHistory.history.length > 0 && (
        <div style={{ marginTop: 16 }}>
          <label className="form-label">Execution History</label>
          <div style={{ background: '#f9fafb', padding: 12, borderRadius: 6, marginTop: 8 }}>
            <table style={{ width: '100%', fontSize: 12 }}>
              <thead><tr style={{ color: '#6b7280' }}><th style={{ textAlign: 'left', padding: 4 }}>Worker</th><th style={{ textAlign: 'left', padding: 4 }}>Start</th><th style={{ textAlign: 'left', padding: 4 }}>End</th><th style={{ textAlign: 'left', padding: 4 }}>Duration</th></tr></thead>
              <tbody>
                {execHistory.history.map((h: any, i: number) => (
                  <tr key={i}>
                    <td style={{ padding: 4, fontFamily: 'monospace' }}>{h.worker_id?.substring(0, 16) || '-'}</td>
                    <td style={{ padding: 4 }}>{h.start_time ? new Date(h.start_time).toLocaleString() : '-'}</td>
                    <td style={{ padding: 4 }}>{h.end_time ? new Date(h.end_time).toLocaleString() : '-'}</td>
                    <td style={{ padding: 4 }}>{h.duration_seconds ? `${h.duration_seconds.toFixed(2)}s` : '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
};

export default TasksPage;
