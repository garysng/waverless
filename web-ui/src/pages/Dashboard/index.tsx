import { useEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import {
  CloudServerOutlined,
  TeamOutlined,
  UnorderedListOutlined,
  ThunderboltOutlined,
  ArrowUpOutlined,
} from '@ant-design/icons';
import * as echarts from 'echarts';
import { api } from '@/api/client';

const DashboardPage = () => {
  const statusChartRef = useRef<HTMLDivElement>(null);
  const endpointChartRef = useRef<HTMLDivElement>(null);

  // Fetch endpoints
  const { data: endpoints } = useQuery({
    queryKey: ['endpoints'],
    queryFn: async () => {
      const res = await api.endpoints.list();
      return Array.isArray(res.data) ? res.data : (res.data?.endpoints || []);
    },
  });

  // Fetch statistics
  const { data: stats } = useQuery({
    queryKey: ['statistics'],
    queryFn: async () => (await api.statistics.getOverview()).data,
  });

  // Fetch endpoint statistics
  const { data: endpointStats } = useQuery({
    queryKey: ['endpoint-statistics'],
    queryFn: async () => (await api.get('/api/v1/statistics/endpoints')).data,
  });

  // Calculate totals
  const totalEndpoints = endpoints?.length || 0;
  const runningEndpoints = endpoints?.filter((e: any) => e.status === 'Running').length || 0;
  const totalWorkers = endpoints?.reduce((sum: number, e: any) => sum + (e.readyReplicas || 0), 0) || 0;

  // Task Status Pie Chart
  useEffect(() => {
    if (!statusChartRef.current || !stats) return;
    const chart = echarts.init(statusChartRef.current);
    chart.setOption({
      tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
      series: [{
        type: 'pie',
        radius: ['50%', '70%'],
        center: ['50%', '50%'],
        data: [
          { value: stats.completed || 0, name: 'Completed', itemStyle: { color: '#48bb78' } },
          { value: stats.in_progress || 0, name: 'Running', itemStyle: { color: '#1da1f2' } },
          { value: stats.pending || 0, name: 'Pending', itemStyle: { color: '#f59e0b' } },
          { value: stats.failed || 0, name: 'Failed', itemStyle: { color: '#f56565' } },
        ].filter(d => d.value > 0),
        label: { show: false },
      }],
    });
    return () => chart.dispose();
  }, [stats]);

  // Endpoint Tasks Bar Chart
  useEffect(() => {
    if (!endpointChartRef.current || !endpointStats?.endpoints) return;
    const chart = echarts.init(endpointChartRef.current);
    const topEndpoints = endpointStats.endpoints.slice(0, 7);
    chart.setOption({
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
      grid: { left: 100, right: 20, top: 20, bottom: 30 },
      xAxis: { type: 'value', axisLabel: { color: '#6b7280' } },
      yAxis: { type: 'category', data: topEndpoints.map((e: any) => e.endpoint.length > 15 ? e.endpoint.substring(0, 15) + '...' : e.endpoint), axisLabel: { color: '#6b7280' } },
      series: [
        { name: 'Completed', type: 'bar', stack: 'total', data: topEndpoints.map((e: any) => e.completed), itemStyle: { color: '#48bb78' } },
        { name: 'Failed', type: 'bar', stack: 'total', data: topEndpoints.map((e: any) => e.failed), itemStyle: { color: '#f56565' } },
      ],
    });
    return () => chart.dispose();
  }, [endpointStats]);

  return (
    <>
      {/* Stats Grid */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-icon blue"><CloudServerOutlined /></div>
          <div className="stat-content">
            <div className="stat-label">Total Endpoints</div>
            <div className="stat-value">{totalEndpoints}</div>
            <div className="stat-change positive"><ArrowUpOutlined /> {runningEndpoints} running</div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon green"><TeamOutlined /></div>
          <div className="stat-content">
            <div className="stat-label">Active Workers</div>
            <div className="stat-value">{totalWorkers}</div>
            <div className="stat-change positive"><ArrowUpOutlined /> online</div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon purple"><UnorderedListOutlined /></div>
          <div className="stat-content">
            <div className="stat-label">Tasks (24h)</div>
            <div className="stat-value">{stats?.completed || 0}</div>
            <div className="stat-change positive"><ArrowUpOutlined /> completed</div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon orange"><ThunderboltOutlined /></div>
          <div className="stat-content">
            <div className="stat-label">Pending Tasks</div>
            <div className="stat-value">{stats?.pending || 0}</div>
            <div className="stat-change">{stats?.in_progress || 0} in progress</div>
          </div>
        </div>
      </div>

      {/* Charts */}
      <div className="charts-row">
        <div className="card">
          <div className="card-header">
            <h3>Tasks by Endpoint</h3>
            <span className="subtitle">Top endpoints by task count</span>
          </div>
          <div className="chart-container" ref={endpointChartRef}></div>
        </div>
        <div className="card">
          <div className="card-header"><h3>Task Status</h3></div>
          <div className="chart-container" ref={statusChartRef}></div>
        </div>
      </div>

      {/* Lists */}
      <div className="lists-row">
        <div className="card">
          <div className="card-header">
            <h3>Recent Endpoints</h3>
            <Link to="/endpoints" className="view-all">View All →</Link>
          </div>
          <div className="list-content">
            {endpoints?.slice(0, 5).map((ep: any) => (
              <Link to={`/endpoints/${ep.name}`} key={ep.name} className="list-item" style={{ textDecoration: 'none' }}>
                <div className="item-icon gpu"><ThunderboltOutlined /></div>
                <div className="item-info">
                  <div className="item-name">{ep.name}</div>
                  <div className="item-meta">{ep.specName} • {ep.readyReplicas || 0} workers</div>
                </div>
                <span className={`tag ${ep.status === 'Running' ? 'running' : ep.status === 'Pending' ? 'pending' : 'stopped'}`}>
                  {ep.status}
                </span>
              </Link>
            ))}
            {(!endpoints || endpoints.length === 0) && (
              <div className="empty-state">
                <p>No endpoints yet</p>
              </div>
            )}
          </div>
        </div>
        <div className="card">
          <div className="card-header">
            <h3>Quick Actions</h3>
          </div>
          <div className="card-body">
            <Link to="/serverless" className="btn btn-blue" style={{ width: '100%', justifyContent: 'center', marginBottom: 12 }}>
              <ThunderboltOutlined /> Create New Endpoint
            </Link>
            <Link to="/tasks" className="btn btn-outline" style={{ width: '100%', justifyContent: 'center' }}>
              <UnorderedListOutlined /> View All Tasks
            </Link>
          </div>
        </div>
      </div>
    </>
  );
};

export default DashboardPage;
