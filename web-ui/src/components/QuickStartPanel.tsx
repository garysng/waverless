import { useState } from 'react';
import { Card, Button, Form, Input, Space, Alert, Typography, Tabs, message, Row, Col, Tag, Collapse } from 'antd';
import { PlayCircleOutlined, CopyOutlined, CodeOutlined, ThunderboltOutlined, SearchOutlined } from '@ant-design/icons';
import { useMutation } from '@tanstack/react-query';
import axios from 'axios';
import type { Task } from '@/types';

const { TextArea } = Input;
const { Text } = Typography;

interface QuickStartPanelProps {
  endpoint: string;
}

// Get API backend URL from environment variable or fall back to window.location.origin
const getApiBackendUrl = () => {
  return import.meta.env.VITE_API_BACKEND_URL || window.location.origin;
};

const QuickStartPanel = ({ endpoint }: QuickStartPanelProps) => {
  const [form] = Form.useForm();
  const [statusForm] = Form.useForm();
  const [responseData, setResponseData] = useState<Task | null>(null);
  const [statusData, setStatusData] = useState<Task | null>(null);
  const [requestMethod, setRequestMethod] = useState<'run' | 'runsync' | 'status'>('run');

  const apiBackendUrl = getApiBackendUrl();

  // Submit task mutation
  const submitTaskMutation = useMutation({
    mutationFn: async (data: { endpoint: string; input: any; method: 'run' | 'runsync' }) => {
      const url = `/v1/${data.endpoint}/${data.method}`;
      const response = await axios.post(url, { input: data.input });
      return response.data;
    },
    onSuccess: (data) => {
      setResponseData(data);
      message.success(requestMethod === 'run' ? 'Task submitted successfully' : 'Task completed');
      // If it's async submission, auto-fill the task ID in status form
      if (requestMethod === 'run' && data.id) {
        statusForm.setFieldValue('taskId', data.id);
      }
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to submit task');
    },
  });

  // Query task status mutation
  const queryStatusMutation = useMutation({
    mutationFn: async (taskId: string) => {
      const response = await axios.get(`/v1/status/${taskId}`);
      return response.data;
    },
    onSuccess: (data) => {
      setStatusData(data);
      message.success('Status retrieved successfully');
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to retrieve status');
    },
  });

  const handleSubmit = (values: any) => {
    let inputData: any;
    try {
      inputData = values.input ? JSON.parse(values.input) : {};
    } catch (error) {
      message.error('Invalid JSON input');
      return;
    }

    submitTaskMutation.mutate({
      endpoint,
      input: inputData,
      method: requestMethod as 'run' | 'runsync',
    });
  };

  const handleStatusQuery = (values: any) => {
    if (!values.taskId || values.taskId.trim() === '') {
      message.error('Please enter a task ID');
      return;
    }
    queryStatusMutation.mutate(values.taskId.trim());
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    message.success('Copied to clipboard');
  };

  const getCurlCommand = () => {
    if (requestMethod === 'status') {
      const taskId = statusForm.getFieldValue('taskId') || 'YOUR_TASK_ID';
      return `curl -X GET ${apiBackendUrl}/v1/status/${taskId}`;
    }

    const inputValue = form.getFieldValue('input') || '{"prompt": "Your prompt"}';
    let formattedInput: string;
    try {
      const parsed = JSON.parse(inputValue);
      formattedInput = JSON.stringify(parsed);
    } catch {
      formattedInput = '{"prompt": "Your prompt"}';
    }

    return `curl -X POST ${apiBackendUrl}/v1/${endpoint}/${requestMethod} \\
  -H "Content-Type: application/json" \\
  -d '{"input": ${formattedInput}}'`;
  };

  const getPythonExample = () => {
    if (requestMethod === 'status') {
      return `import requests

task_id = "YOUR_TASK_ID"
url = f"${apiBackendUrl}/v1/status/{task_id}"

response = requests.get(url)
result = response.json()
print(result)`;
    }

    return `import requests

url = "${apiBackendUrl}/v1/${endpoint}/${requestMethod}"
payload = {
    "input": {
        "prompt": "Your prompt"
    }
}

response = requests.post(url, json=payload)
result = response.json()
print(result)`;
  };

  const getJavaScriptExample = () => {
    if (requestMethod === 'status') {
      return `const taskId = "YOUR_TASK_ID";
const response = await fetch(\`${apiBackendUrl}/v1/status/\${taskId}\`);

const result = await response.json();
console.log(result);`;
    }

    return `const response = await fetch('${apiBackendUrl}/v1/${endpoint}/${requestMethod}', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    input: {
      prompt: "Your prompt"
    }
  })
});

const result = await response.json();
console.log(result);`;
  };

  const renderTestForm = () => {
    if (requestMethod === 'status') {
      return (
        <>
          <Alert
            message="Query task status by task ID"
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
          />

          <Form form={statusForm} layout="vertical" onFinish={handleStatusQuery}>
            <Form.Item
              label="Task ID"
              name="taskId"
              rules={[{ required: true, message: 'Please enter task ID' }]}
            >
              <Input
                placeholder="Enter task ID"
                style={{ fontFamily: 'monospace' }}
              />
            </Form.Item>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                icon={<SearchOutlined />}
                loading={queryStatusMutation.isPending}
                block
              >
                Query Status
              </Button>
            </Form.Item>
          </Form>

          {statusData && (
            <Card
              size="small"
              title={
                <Space>
                  <Text>Status Result</Text>
                  <Tag color={
                    statusData.status === 'COMPLETED' ? 'success' :
                    statusData.status === 'FAILED' ? 'error' :
                    statusData.status === 'IN_PROGRESS' ? 'processing' : 'default'
                  }>
                    {statusData.status}
                  </Tag>
                </Space>
              }
              extra={
                <Button
                  size="small"
                  icon={<CopyOutlined />}
                  onClick={() => copyToClipboard(JSON.stringify(statusData, null, 2))}
                >
                  Copy
                </Button>
              }
            >
              <pre style={{
                margin: 0,
                fontSize: 11,
                backgroundColor: '#f5f5f5',
                padding: 12,
                borderRadius: 4,
                maxHeight: 200,
                overflow: 'auto'
              }}>
                {JSON.stringify(statusData, null, 2)}
              </pre>
            </Card>
          )}
        </>
      );
    }

    return (
      <>
        <Alert
          message="Submit a test task to verify your endpoint"
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />

        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item
            label="Request Body (JSON)"
            name="input"
            initialValue={'{\n  "prompt": "Your prompt here"\n}'}
            rules={[
              {
                validator: (_, value) => {
                  if (!value) return Promise.resolve();
                  try {
                    JSON.parse(value);
                    return Promise.resolve();
                  } catch {
                    return Promise.reject('Invalid JSON format');
                  }
                },
              },
            ]}
          >
            <TextArea
              rows={6}
              placeholder='{"prompt": "Your prompt"}'
              style={{ fontFamily: 'monospace', fontSize: 12 }}
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              icon={<PlayCircleOutlined />}
              loading={submitTaskMutation.isPending}
              block
            >
              Run Test
            </Button>
          </Form.Item>
        </Form>

        {responseData && (
          <Card
            size="small"
            title={
              <Space>
                <Text>Response</Text>
                <Tag color={responseData.status === 'COMPLETED' ? 'success' : 'processing'}>
                  {responseData.status}
                </Tag>
              </Space>
            }
            extra={
              <Button
                size="small"
                icon={<CopyOutlined />}
                onClick={() => copyToClipboard(JSON.stringify(responseData, null, 2))}
              >
                Copy
              </Button>
            }
          >
            <pre style={{
              margin: 0,
              fontSize: 11,
              backgroundColor: '#f5f5f5',
              padding: 12,
              borderRadius: 4,
              maxHeight: 200,
              overflow: 'auto'
            }}>
              {JSON.stringify(responseData, null, 2)}
            </pre>
          </Card>
        )}
      </>
    );
  };

  return (
    <Collapse
      defaultActiveKey={[]}
      expandIconPosition="end"
      style={{ marginBottom: 24 }}
      items={[
        {
          key: '1',
          label: (
            <Space>
              <ThunderboltOutlined style={{ color: '#1890ff' }} />
              <span style={{ fontWeight: 500 }}>Quick Start</span>
              <Text type="secondary" style={{ fontSize: 12 }}>
                Test your endpoint with a sample request
              </Text>
            </Space>
          ),
          children: (
            <Row gutter={16}>
              {/* Left: Test Form */}
              <Col xs={24} lg={12}>
                <Form.Item label="Method" style={{ marginBottom: 16 }}>
                  <Space>
                    <Button
                      size="small"
                      type={requestMethod === 'run' ? 'primary' : 'default'}
                      onClick={() => setRequestMethod('run')}
                    >
                      /run (async)
                    </Button>
                    <Button
                      size="small"
                      type={requestMethod === 'runsync' ? 'primary' : 'default'}
                      onClick={() => setRequestMethod('runsync')}
                    >
                      /runsync (sync)
                    </Button>
                    <Button
                      size="small"
                      type={requestMethod === 'status' ? 'primary' : 'default'}
                      onClick={() => setRequestMethod('status')}
                    >
                      /status (query)
                    </Button>
                  </Space>
                </Form.Item>

                {renderTestForm()}
              </Col>

              {/* Right: Code Examples */}
              <Col xs={24} lg={12}>
                <Tabs
                  size="small"
                  items={[
                    {
                      key: 'curl',
                      label: (
                        <span>
                          <CodeOutlined />
                          cURL
                        </span>
                      ),
                      children: (
                        <div style={{ position: 'relative' }}>
                          <Button
                            size="small"
                            icon={<CopyOutlined />}
                            onClick={() => copyToClipboard(getCurlCommand())}
                            style={{ position: 'absolute', right: 0, top: 0, zIndex: 1 }}
                          >
                            Copy
                          </Button>
                          <pre style={{
                            margin: 0,
                            fontSize: 11,
                            backgroundColor: '#f5f5f5',
                            padding: 12,
                            paddingTop: 32,
                            borderRadius: 4,
                            overflow: 'auto',
                            minHeight: 150
                          }}>
                            {getCurlCommand()}
                          </pre>
                        </div>
                      ),
                    },
                    {
                      key: 'python',
                      label: (
                        <span>
                          <CodeOutlined />
                          Python
                        </span>
                      ),
                      children: (
                        <div style={{ position: 'relative' }}>
                          <Button
                            size="small"
                            icon={<CopyOutlined />}
                            onClick={() => copyToClipboard(getPythonExample())}
                            style={{ position: 'absolute', right: 0, top: 0, zIndex: 1 }}
                          >
                            Copy
                          </Button>
                          <pre style={{
                            margin: 0,
                            fontSize: 11,
                            backgroundColor: '#f5f5f5',
                            padding: 12,
                            paddingTop: 32,
                            borderRadius: 4,
                            overflow: 'auto',
                            minHeight: 150
                          }}>
                            {getPythonExample()}
                          </pre>
                        </div>
                      ),
                    },
                    {
                      key: 'javascript',
                      label: (
                        <span>
                          <CodeOutlined />
                          JavaScript
                        </span>
                      ),
                      children: (
                        <div style={{ position: 'relative' }}>
                          <Button
                            size="small"
                            icon={<CopyOutlined />}
                            onClick={() => copyToClipboard(getJavaScriptExample())}
                            style={{ position: 'absolute', right: 0, top: 0, zIndex: 1 }}
                          >
                            Copy
                          </Button>
                          <pre style={{
                            margin: 0,
                            fontSize: 11,
                            backgroundColor: '#f5f5f5',
                            padding: 12,
                            paddingTop: 32,
                            borderRadius: 4,
                            overflow: 'auto',
                            minHeight: 150
                          }}>
                            {getJavaScriptExample()}
                          </pre>
                        </div>
                      ),
                    },
                  ]}
                />

                <Alert
                  message="API Endpoints"
                  description={
                    <Space direction="vertical" size="small" style={{ width: '100%' }}>
                      <Text style={{ fontSize: 12 }}>
                        <strong>POST</strong> <Text code>{apiBackendUrl}/v1/{endpoint}/run</Text>
                      </Text>
                      <Text style={{ fontSize: 12 }}>
                        <strong>POST</strong> <Text code>{apiBackendUrl}/v1/{endpoint}/runsync</Text>
                      </Text>
                      <Text style={{ fontSize: 12 }}>
                        <strong>GET</strong> <Text code>{apiBackendUrl}/v1/status/:task_id</Text>
                      </Text>
                    </Space>
                  }
                  type="info"
                  showIcon
                  style={{ marginTop: 16 }}
                />
              </Col>
            </Row>
          ),
        },
      ]}
    />
  );
};

export default QuickStartPanel;
