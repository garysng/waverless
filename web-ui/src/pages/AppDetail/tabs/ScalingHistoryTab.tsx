import { useQuery } from '@tanstack/react-query';
import { Empty, Space, Tag, Timeline, Typography } from 'antd';
import { api } from '@/api/client';

const { Text } = Typography;

interface ScalingEvent {
  id: string;
  endpoint: string;
  timestamp: string;
  action: string;
  fromReplicas: number;
  toReplicas: number;
  reason: string;
  queueLength: number;
  priority?: number;
}

interface ScalingHistoryTabProps {
  endpoint: string;
}

const ScalingHistoryTab = ({ endpoint }: ScalingHistoryTabProps) => {
  const { data: history, isLoading } = useQuery<ScalingEvent[]>({
    queryKey: ['autoscaler-history', endpoint],
    queryFn: async () => {
      const response = await api.autoscaler.getHistory(endpoint, 50);
      return response.data;
    },
    enabled: !!endpoint,
    refetchInterval: 20000,
    placeholderData: (previousData) => previousData,
  });

  if (isLoading) {
    return (
      <div style={{ padding: 24 }}>
        <Text type="secondary">Loading scaling events...</Text>
      </div>
    );
  }

  if (!history || history.length === 0) {
    return (
      <div style={{ padding: 24 }}>
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="No scaling events yet" />
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      <Timeline
        items={history.map((event: ScalingEvent) => ({
          color: event.action === 'scale_up' ? 'green' : event.action === 'scale_down' ? 'blue' : 'orange',
          children: (
            <Space direction="vertical" size={0}>
              <Space>
                <Tag color={event.action === 'scale_up' ? 'green' : event.action === 'scale_down' ? 'blue' : 'orange'}>
                  {event.action.replace('_', ' ').toUpperCase()}
                </Tag>
                <Text>
                  {event.fromReplicas} → {event.toReplicas} replicas
                </Text>
              </Space>
              <Text type="secondary">
                {event.reason} (Queue: {event.queueLength}
                {typeof event.priority === 'number' ? ` | Priority: ${event.priority}` : ''})
              </Text>
              <Text type="secondary" style={{ fontSize: 12 }}>
                {new Date(event.timestamp).toLocaleString()}
              </Text>
            </Space>
          ),
        }))}
      />
    </div>
  );
};

export default ScalingHistoryTab;
