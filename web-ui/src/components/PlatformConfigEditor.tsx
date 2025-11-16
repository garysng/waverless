import { useState } from 'react';
import {
  Input,
  Button,
  Space,
  Card,
  Select,
  Collapse,
  Row,
  Col,
  Typography,
  Popconfirm,
  Tabs,
  message,
} from 'antd';
import {
  PlusOutlined,
  DeleteOutlined,
  CopyOutlined,
} from '@ant-design/icons';
import type { CollapseProps } from 'antd';
import yaml from 'js-yaml';

const { Text } = Typography;
const { Option } = Select;
const { TextArea } = Input;

export interface Toleration {
  key: string;
  operator: string;
  value?: string;
  effect: string;
}

export interface PlatformConfig {
  nodeSelector?: Record<string, string>;
  tolerations?: Toleration[];
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
}

export interface PlatformConfigs {
  [platformName: string]: PlatformConfig;
}

interface PlatformConfigEditorProps {
  value?: PlatformConfigs;
  onChange?: (value: PlatformConfigs) => void;
  mode?: 'simple' | 'advanced' | 'yaml';
}

const PlatformConfigEditor: React.FC<PlatformConfigEditorProps> = ({
  value = {},
  onChange,
  mode = 'simple',
}) => {
  const [activeMode, setActiveMode] = useState<'simple' | 'advanced' | 'yaml'>(mode);
  const [platforms, setPlatforms] = useState<PlatformConfigs>(value);

  const handlePlatformsChange = (newPlatforms: PlatformConfigs) => {
    setPlatforms(newPlatforms);
    onChange?.(newPlatforms);
  };

  // Add new platform
  const addPlatform = () => {
    const platformName = prompt('Enter platform name (e.g., generic, aliyun-ack, aws-eks):');
    if (platformName && !platforms[platformName]) {
      handlePlatformsChange({
        ...platforms,
        [platformName]: {
          nodeSelector: {},
          tolerations: [],
          labels: {},
          annotations: {},
        },
      });
    }
  };

  // Remove platform
  const removePlatform = (platformName: string) => {
    const { [platformName]: _, ...rest } = platforms;
    handlePlatformsChange(rest);
  };

  // Add toleration to a platform
  const addToleration = (platformName: string) => {
    const platform = platforms[platformName];
    handlePlatformsChange({
      ...platforms,
      [platformName]: {
        ...platform,
        tolerations: [
          ...(platform.tolerations || []),
          { key: '', operator: 'Equal', value: '', effect: 'NoSchedule' },
        ],
      },
    });
  };

  // Update toleration
  const updateToleration = (
    platformName: string,
    index: number,
    field: keyof Toleration,
    value: string
  ) => {
    const platform = platforms[platformName];
    const tolerations = [...(platform.tolerations || [])];
    tolerations[index] = { ...tolerations[index], [field]: value };
    handlePlatformsChange({
      ...platforms,
      [platformName]: { ...platform, tolerations },
    });
  };

  // Remove toleration
  const removeToleration = (platformName: string, index: number) => {
    const platform = platforms[platformName];
    const tolerations = (platform.tolerations || []).filter((_, i) => i !== index);
    handlePlatformsChange({
      ...platforms,
      [platformName]: { ...platform, tolerations },
    });
  };

  // Add key-value pair to nodeSelector/labels/annotations (inline)
  const addKeyValue = (
    platformName: string,
    field: 'nodeSelector' | 'labels' | 'annotations'
  ) => {
    const platform = platforms[platformName];
    // Add empty key-value pair
    handlePlatformsChange({
      ...platforms,
      [platformName]: {
        ...platform,
        [field]: { ...(platform[field] || {}), '': '' },
      },
    });
  };

  // Update key-value pair (supports updating both key and value)
  const updateKeyValue = (
    platformName: string,
    field: 'nodeSelector' | 'labels' | 'annotations',
    oldKey: string,
    newKey: string,
    value: string
  ) => {
    const platform = platforms[platformName];
    const fieldData = platform[field] || {};

    // Remove old key if key changed
    if (oldKey !== newKey) {
      const { [oldKey]: _, ...rest } = fieldData;
      handlePlatformsChange({
        ...platforms,
        [platformName]: {
          ...platform,
          [field]: { ...rest, [newKey]: value },
        },
      });
    } else {
      // Just update value
      handlePlatformsChange({
        ...platforms,
        [platformName]: {
          ...platform,
          [field]: { ...fieldData, [newKey]: value },
        },
      });
    }
  };

  // Remove key-value pair
  const removeKeyValue = (
    platformName: string,
    field: 'nodeSelector' | 'labels' | 'annotations',
    key: string
  ) => {
    const platform = platforms[platformName];
    const { [key]: _, ...rest } = platform[field] || {};
    handlePlatformsChange({
      ...platforms,
      [platformName]: { ...platform, [field]: rest },
    });
  };

  // Render simple mode (form-based)
  const renderSimpleMode = () => {
    return (
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        <Button icon={<PlusOutlined />} onClick={addPlatform}>
          Add Platform
        </Button>

        {Object.entries(platforms).map(([platformName, config]) => {
          const items: CollapseProps['items'] = [
            {
              key: 'tolerations',
              label: `Tolerations (${config.tolerations?.length || 0})`,
              children: (
                <Space direction="vertical" style={{ width: '100%' }}>
                  {config.tolerations?.map((tol, index) => (
                    <Card key={index} size="small" style={{ marginBottom: 8 }}>
                      <Row gutter={8}>
                        <Col span={6}>
                          <Input
                            placeholder="Key"
                            value={tol.key}
                            onChange={(e) =>
                              updateToleration(platformName, index, 'key', e.target.value)
                            }
                          />
                        </Col>
                        <Col span={5}>
                          <Select
                            style={{ width: '100%' }}
                            value={tol.operator}
                            onChange={(val) =>
                              updateToleration(platformName, index, 'operator', val)
                            }
                          >
                            <Option value="Equal">Equal</Option>
                            <Option value="Exists">Exists</Option>
                          </Select>
                        </Col>
                        <Col span={5}>
                          <Input
                            placeholder="Value"
                            value={tol.value}
                            disabled={tol.operator === 'Exists'}
                            onChange={(e) =>
                              updateToleration(platformName, index, 'value', e.target.value)
                            }
                          />
                        </Col>
                        <Col span={6}>
                          <Select
                            style={{ width: '100%' }}
                            value={tol.effect}
                            onChange={(val) =>
                              updateToleration(platformName, index, 'effect', val)
                            }
                          >
                            <Option value="NoSchedule">NoSchedule</Option>
                            <Option value="PreferNoSchedule">PreferNoSchedule</Option>
                            <Option value="NoExecute">NoExecute</Option>
                          </Select>
                        </Col>
                        <Col span={2}>
                          <Button
                            danger
                            size="small"
                            icon={<DeleteOutlined />}
                            onClick={() => removeToleration(platformName, index)}
                          />
                        </Col>
                      </Row>
                    </Card>
                  ))}
                  <Button
                    type="dashed"
                    icon={<PlusOutlined />}
                    onClick={() => addToleration(platformName)}
                    block
                  >
                    Add Toleration
                  </Button>
                </Space>
              ),
            },
            {
              key: 'nodeSelector',
              label: `Node Selector (${Object.keys(config.nodeSelector || {}).length})`,
              children: (
                <Space direction="vertical" style={{ width: '100%' }}>
                  {Object.entries(config.nodeSelector || {}).map(([key, value], index) => (
                    <Card key={index} size="small" style={{ marginBottom: 8 }}>
                      <Row gutter={8}>
                        <Col span={10}>
                          <Input
                            placeholder="Key (e.g., gpu.nvidia.com/class)"
                            value={key}
                            onChange={(e) =>
                              updateKeyValue(platformName, 'nodeSelector', key, e.target.value, value)
                            }
                          />
                        </Col>
                        <Col span={12}>
                          <Input
                            placeholder="Value (e.g., H200)"
                            value={value}
                            onChange={(e) =>
                              updateKeyValue(platformName, 'nodeSelector', key, key, e.target.value)
                            }
                          />
                        </Col>
                        <Col span={2}>
                          <Button
                            danger
                            size="small"
                            icon={<DeleteOutlined />}
                            onClick={() => removeKeyValue(platformName, 'nodeSelector', key)}
                          />
                        </Col>
                      </Row>
                    </Card>
                  ))}
                  <Button
                    type="dashed"
                    icon={<PlusOutlined />}
                    onClick={() => addKeyValue(platformName, 'nodeSelector')}
                    block
                  >
                    Add Node Selector
                  </Button>
                </Space>
              ),
            },
            {
              key: 'labels',
              label: `Labels (${Object.keys(config.labels || {}).length})`,
              children: (
                <Space direction="vertical" style={{ width: '100%' }}>
                  {Object.entries(config.labels || {}).map(([key, value], index) => (
                    <Card key={index} size="small" style={{ marginBottom: 8 }}>
                      <Row gutter={8}>
                        <Col span={10}>
                          <Input
                            placeholder="Key (e.g., app)"
                            value={key}
                            onChange={(e) =>
                              updateKeyValue(platformName, 'labels', key, e.target.value, value)
                            }
                          />
                        </Col>
                        <Col span={12}>
                          <Input
                            placeholder="Value (e.g., waverless)"
                            value={value}
                            onChange={(e) =>
                              updateKeyValue(platformName, 'labels', key, key, e.target.value)
                            }
                          />
                        </Col>
                        <Col span={2}>
                          <Button
                            danger
                            size="small"
                            icon={<DeleteOutlined />}
                            onClick={() => removeKeyValue(platformName, 'labels', key)}
                          />
                        </Col>
                      </Row>
                    </Card>
                  ))}
                  <Button
                    type="dashed"
                    icon={<PlusOutlined />}
                    onClick={() => addKeyValue(platformName, 'labels')}
                    block
                  >
                    Add Label
                  </Button>
                </Space>
              ),
            },
            {
              key: 'annotations',
              label: `Annotations (${Object.keys(config.annotations || {}).length})`,
              children: (
                <Space direction="vertical" style={{ width: '100%' }}>
                  {Object.entries(config.annotations || {}).map(([key, value], index) => (
                    <Card key={index} size="small" style={{ marginBottom: 8 }}>
                      <Row gutter={8}>
                        <Col span={10}>
                          <Input
                            placeholder="Key (e.g., k8s.aliyun.com/...)"
                            value={key}
                            onChange={(e) =>
                              updateKeyValue(platformName, 'annotations', key, e.target.value, value)
                            }
                          />
                        </Col>
                        <Col span={12}>
                          <Input
                            placeholder="Value"
                            value={value}
                            onChange={(e) =>
                              updateKeyValue(platformName, 'annotations', key, key, e.target.value)
                            }
                          />
                        </Col>
                        <Col span={2}>
                          <Button
                            danger
                            size="small"
                            icon={<DeleteOutlined />}
                            onClick={() => removeKeyValue(platformName, 'annotations', key)}
                          />
                        </Col>
                      </Row>
                    </Card>
                  ))}
                  <Button
                    type="dashed"
                    icon={<PlusOutlined />}
                    onClick={() => addKeyValue(platformName, 'annotations')}
                    block
                  >
                    Add Annotation
                  </Button>
                </Space>
              ),
            },
          ];

          return (
            <Card
              key={platformName}
              title={
                <Space>
                  <Text strong>{platformName}</Text>
                </Space>
              }
              extra={
                <Popconfirm
                  title="Remove Platform"
                  description={`Are you sure you want to remove platform "${platformName}"?`}
                  onConfirm={() => removePlatform(platformName)}
                  okText="Yes"
                  cancelText="No"
                >
                  <Button danger size="small" icon={<DeleteOutlined />}>
                    Remove Platform
                  </Button>
                </Popconfirm>
              }
            >
              <Collapse items={items} />
            </Card>
          );
        })}
      </Space>
    );
  };

  // Render advanced mode (JSON editor)
  const renderAdvancedMode = () => {
    const [jsonText, setJsonText] = useState(JSON.stringify(platforms, null, 2));
    const [error, setError] = useState<string | null>(null);

    const handleJsonChange = (text: string) => {
      setJsonText(text);
      try {
        const parsed = JSON.parse(text);
        setError(null);
        handlePlatformsChange(parsed);
      } catch (e) {
        setError((e as Error).message);
      }
    };

    return (
      <Space direction="vertical" style={{ width: '100%' }}>
        {error && (
          <Text type="danger">Invalid JSON: {error}</Text>
        )}
        <TextArea
          value={jsonText}
          onChange={(e) => handleJsonChange(e.target.value)}
          rows={20}
          style={{ fontFamily: 'monospace' }}
        />
        <Space>
          <Button
            icon={<CopyOutlined />}
            onClick={() => {
              navigator.clipboard.writeText(jsonText);
              message.success('Copied to clipboard');
            }}
          >
            Copy JSON
          </Button>
          <Button
            onClick={() => {
              try {
                const formatted = JSON.stringify(JSON.parse(jsonText), null, 2);
                setJsonText(formatted);
                message.success('Formatted successfully');
              } catch (e) {
                message.error('Invalid JSON format');
              }
            }}
          >
            Format JSON
          </Button>
        </Space>
      </Space>
    );
  };

  // Render YAML mode
  const renderYamlMode = () => {
    const [yamlText, setYamlText] = useState(yaml.dump(platforms, { indent: 2 }));
    const [error, setError] = useState<string | null>(null);

    const handleYamlChange = (text: string) => {
      setYamlText(text);
      try {
        const parsed = yaml.load(text) as PlatformConfigs;
        setError(null);
        handlePlatformsChange(parsed);
      } catch (e) {
        setError((e as Error).message);
      }
    };

    return (
      <Space direction="vertical" style={{ width: '100%' }}>
        {error && (
          <Text type="danger">Invalid YAML: {error}</Text>
        )}
        <TextArea
          value={yamlText}
          onChange={(e) => handleYamlChange(e.target.value)}
          rows={20}
          style={{ fontFamily: 'monospace' }}
          placeholder={`Example:
generic:
  nodeSelector: {}
  tolerations: []
  labels: {}
  annotations: {}
aliyun-ack:
  nodeSelector:
    gpu.nvidia.com/class: H200
  tolerations:
    - key: hardware-type/h200
      operator: Equal
      value: gpu-cpfs
      effect: NoSchedule
  labels:
    alibabacloud.com/acs: "true"
  annotations:
    k8s.aliyun.com/image-accelerate-mode: on-demand`}
        />
        <Space>
          <Button
            icon={<CopyOutlined />}
            onClick={() => {
              navigator.clipboard.writeText(yamlText);
              message.success('Copied to clipboard');
            }}
          >
            Copy YAML
          </Button>
          <Button
            onClick={() => {
              try {
                const parsed = yaml.load(yamlText);
                const formatted = yaml.dump(parsed, { indent: 2 });
                setYamlText(formatted);
                message.success('Formatted successfully');
              } catch (e) {
                message.error('Invalid YAML format');
              }
            }}
          >
            Format YAML
          </Button>
          <Button
            onClick={() => {
              try {
                const parsed = yaml.load(yamlText);
                const jsonStr = JSON.stringify(parsed, null, 2);
                navigator.clipboard.writeText(jsonStr);
                message.success('Converted to JSON and copied');
              } catch (e) {
                message.error('Invalid YAML format');
              }
            }}
          >
            Convert to JSON
          </Button>
        </Space>
      </Space>
    );
  };

  return (
    <div>
      <Tabs
        activeKey={activeMode}
        onChange={(key) => setActiveMode(key as 'simple' | 'advanced' | 'yaml')}
        items={[
          {
            key: 'simple',
            label: 'Simple Mode (Form)',
            children: renderSimpleMode(),
          },
          {
            key: 'yaml',
            label: 'YAML Mode',
            children: renderYamlMode(),
          },
          {
            key: 'advanced',
            label: 'JSON Mode',
            children: renderAdvancedMode(),
          },
        ]}
      />
    </div>
  );
};

export default PlatformConfigEditor;
