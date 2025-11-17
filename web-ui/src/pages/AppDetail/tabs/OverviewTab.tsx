import { Descriptions, Tag, Typography } from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  GlobalOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons';
import type { AppInfo } from '@/types';
import { Tooltip } from 'antd';

const { Text } = Typography;

interface OverviewTabProps {
  endpoint: string;
  appInfo?: AppInfo;
}

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

const GLOBAL_TASK_TIMEOUT = 3600;

const renderFieldLabel = (label: string, key: AutoScalerTipKey) => (
  <span>
    {label}
    <Tooltip title={AUTOSCALER_FIELD_TIPS[key]}>
      <QuestionCircleOutlined style={{ marginLeft: 4, color: '#999' }} />
    </Tooltip>
  </span>
);

const renderZeroHint = (condition: boolean, text: string) =>
  condition ? (
    <Text type="secondary" style={{ marginLeft: 8 }}>
      {text}
    </Text>
  ) : null;

const formatTaskTimeout = (value?: number) =>
  value && value > 0 ? `${value}` : `0 (uses global default ${GLOBAL_TASK_TIMEOUT}s)`;

const OverviewTab = ({ appInfo }: OverviewTabProps) => {
  if (!appInfo) {
    return <div style={{ padding: 24 }}>Loading...</div>;
  }

  return (
    <div style={{ padding: 24 }}>
      {/* Basic Information */}
      <Descriptions
        title="Basic Information"
        bordered
        column={2}
        size="middle"
        style={{ marginBottom: 24 }}
      >
        <Descriptions.Item label="Endpoint Name" span={2}>
          {appInfo.name}
        </Descriptions.Item>
        <Descriptions.Item label="Display Name" span={2}>
          {appInfo.displayName || appInfo.name}
        </Descriptions.Item>
        <Descriptions.Item label="Description" span={2}>
          {appInfo.description || '-'}
        </Descriptions.Item>
        <Descriptions.Item label="Namespace">
          {appInfo.namespace || '-'}
        </Descriptions.Item>
        <Descriptions.Item label="Type">
          {appInfo.type || '-'}
        </Descriptions.Item>
        <Descriptions.Item label="Spec">
          {appInfo.specName || '-'}
        </Descriptions.Item>
        <Descriptions.Item label="Task Timeout">
          {formatTaskTimeout(appInfo.taskTimeout)}
        </Descriptions.Item>
        <Descriptions.Item label="Image" span={2}>
          <Text code style={{ fontSize: 12 }}>
            {appInfo.image || '-'}
          </Text>
        </Descriptions.Item>
        <Descriptions.Item label="Created At">
          {appInfo.createdAt ? new Date(appInfo.createdAt).toLocaleString() : '-'}
        </Descriptions.Item>
        <Descriptions.Item label="Updated At">
          {appInfo.updatedAt ? new Date(appInfo.updatedAt).toLocaleString() : '-'}
        </Descriptions.Item>
        <Descriptions.Item label="Last Scale Time">
          {appInfo.lastScaleTime ? new Date(appInfo.lastScaleTime).toLocaleString() : '-'}
        </Descriptions.Item>
        <Descriptions.Item label="Last Task Time">
          {appInfo.lastTaskTime ? new Date(appInfo.lastTaskTime).toLocaleString() : '-'}
        </Descriptions.Item>
      </Descriptions>

      {/* Resource Configuration */}
      <Descriptions
        title="Resource Configuration"
        bordered
        column={2}
        size="middle"
        style={{ marginBottom: 24 }}
      >
        <Descriptions.Item label="Replicas">
          {appInfo.readyReplicas || 0} / {appInfo.replicas || 0}
          {appInfo.minReplicas === 0 && (
            <Text type="secondary" style={{ marginLeft: 8 }}>
              min 0 → can fully scale down when idle
            </Text>
          )}
        </Descriptions.Item>
        <Descriptions.Item label="Available Replicas">
          {appInfo.availableReplicas || 0}
        </Descriptions.Item>
        <Descriptions.Item label="Shared Memory Size">
          {appInfo.shmSize || '-'}
        </Descriptions.Item>
        <Descriptions.Item label="Ptrace Debugging">
          {appInfo.enablePtrace ? (
            <Tag color="green">Enabled</Tag>
          ) : (
            <Tag>Disabled</Tag>
          )}
        </Descriptions.Item>
      </Descriptions>

      {/* AutoScaler Configuration */}
      <Descriptions
        title="AutoScaler Configuration"
        bordered
        column={2}
        size="middle"
        style={{ marginBottom: 24 }}
      >
        <Descriptions.Item label="AutoScaler Override" span={2}>
          {!appInfo.autoscalerEnabled || appInfo.autoscalerEnabled === '' ? (
            <Tag color="default" icon={<GlobalOutlined />}>
              Default (follow global)
            </Tag>
          ) : appInfo.autoscalerEnabled === 'disabled' ? (
            <Tag color="red" icon={<CloseCircleOutlined />}>
              Force Off
            </Tag>
          ) : appInfo.autoscalerEnabled === 'enabled' ? (
            <Tag color="green" icon={<CheckCircleOutlined />}>
              Force On
            </Tag>
          ) : (
            <Tag>Unknown</Tag>
          )}
        </Descriptions.Item>
        <Descriptions.Item label={renderFieldLabel('Priority', 'priority')}>
          {appInfo.priority ?? 50}
          {renderZeroHint(appInfo.priority === 0, '0 = lowest priority (best-effort)')}
        </Descriptions.Item>
        <Descriptions.Item label="Min Replicas">
          {appInfo.minReplicas ?? 0}
          {renderZeroHint(appInfo.minReplicas === 0, 'scale-to-zero enabled')}
        </Descriptions.Item>
        <Descriptions.Item label="Max Replicas">
          {appInfo.maxReplicas ?? 'Not Set'}
          {appInfo.maxReplicas === 0 && renderZeroHint(true, 'maxReplicas=0 will disable autoscaling!')}
        </Descriptions.Item>
        <Descriptions.Item label={renderFieldLabel('Scale Up Threshold', 'scaleUpThreshold')}>
          {appInfo.scaleUpThreshold ?? 1}
          {renderZeroHint(appInfo.scaleUpThreshold === 0, '0 behaves like 1 (scale on first task)')}
        </Descriptions.Item>
        <Descriptions.Item label={renderFieldLabel('Scale Down Idle Time', 'scaleDownIdleTime')}>
          {appInfo.scaleDownIdleTime ?? 300}s
          {renderZeroHint(appInfo.scaleDownIdleTime === 0, '0 = scale down immediately after idle')}
        </Descriptions.Item>
        <Descriptions.Item label={renderFieldLabel('Scale Up Cooldown', 'scaleUpCooldown')}>
          {typeof appInfo.scaleUpCooldown === 'number' ? `${appInfo.scaleUpCooldown}s` : '-'}
          {renderZeroHint(appInfo.scaleUpCooldown === 0, '0 = no cooldown between scale ups')}
        </Descriptions.Item>
        <Descriptions.Item label={renderFieldLabel('Scale Down Cooldown', 'scaleDownCooldown')}>
          {typeof appInfo.scaleDownCooldown === 'number' ? `${appInfo.scaleDownCooldown}s` : '-'}
          {renderZeroHint(appInfo.scaleDownCooldown === 0, '0 = no cooldown between scale downs')}
        </Descriptions.Item>
        <Descriptions.Item label="Dynamic Priority">
          {appInfo.enableDynamicPrio ? (
            <Tag color="blue">Enabled</Tag>
          ) : (
            <Tag>Disabled</Tag>
          )}
        </Descriptions.Item>
        {appInfo.enableDynamicPrio && (
          <>
            <Descriptions.Item label={renderFieldLabel('High Load Threshold', 'highLoadThreshold')}>
              {appInfo.highLoadThreshold ?? '-'}
              {renderZeroHint(appInfo.highLoadThreshold === 0, '0 = always treated as high load')}
            </Descriptions.Item>
            <Descriptions.Item label={renderFieldLabel('Priority Boost', 'priorityBoost')}>
              {appInfo.priorityBoost ?? '-'}
            </Descriptions.Item>
          </>
        )}
      </Descriptions>

      {/* Task Statistics */}
      <Descriptions
        title="Task Statistics"
        bordered
        column={2}
        size="middle"
      >
        <Descriptions.Item label="Total Tasks">
          {appInfo.totalTasks || 0}
        </Descriptions.Item>
        <Descriptions.Item label="Completed Tasks">
          <Tag color="success">{appInfo.completedTasks || 0}</Tag>
        </Descriptions.Item>
        <Descriptions.Item label="Failed Tasks">
          <Tag color="error">{appInfo.failedTasks || 0}</Tag>
        </Descriptions.Item>
        <Descriptions.Item label="Pending Tasks">
          <Tag color="warning">{appInfo.pendingTasks || 0}</Tag>
        </Descriptions.Item>
        <Descriptions.Item label="Running Tasks">
          <Tag color="processing">{appInfo.runningTasks || 0}</Tag>
        </Descriptions.Item>
        <Descriptions.Item label="Worker Count">
          {appInfo.activeWorkerCount || 0} / {appInfo.workerCount || 0}
        </Descriptions.Item>
      </Descriptions>

      {/* Volume Mounts */}
      {appInfo.volumeMounts && appInfo.volumeMounts.length > 0 && (
        <Descriptions
          title="Volume Mounts"
          bordered
          column={1}
          size="middle"
          style={{ marginTop: 24 }}
        >
          {appInfo.volumeMounts.map((mount, index) => (
            <Descriptions.Item key={index} label={`Mount ${index + 1}`}>
              PVC: <Text code>{mount.pvcName}</Text> → Path: <Text code>{mount.mountPath}</Text>
            </Descriptions.Item>
          ))}
        </Descriptions>
      )}

      {/* Environment Variables */}
      {appInfo.env && Object.keys(appInfo.env).length > 0 && (
        <Descriptions
          title="Environment Variables"
          bordered
          column={1}
          size="middle"
          style={{ marginTop: 24 }}
        >
          {Object.entries(appInfo.env).map(([key, value]) => (
            <Descriptions.Item key={key} label={key}>
              <Text code>{value}</Text>
            </Descriptions.Item>
          ))}
        </Descriptions>
      )}
    </div>
  );
};

export default OverviewTab;
