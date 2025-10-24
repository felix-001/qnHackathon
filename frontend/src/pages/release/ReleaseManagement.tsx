import React, { useState, useEffect } from 'react';
import { Table, Button, Modal, Form, Input, Select, message, Tag, Space } from 'antd';
import { PlusOutlined, RollbackOutlined, EyeOutlined } from '@ant-design/icons';
import { releaseAPI } from '@/services/api';
import type { Release } from '@/types';

const ReleaseManagement: React.FC = () => {
  const [releases, setReleases] = useState<Release[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();

  useEffect(() => {
    loadReleases();
  }, []);

  const loadReleases = async () => {
    setLoading(true);
    try {
      const response = await releaseAPI.list();
      setReleases(response.data);
    } catch (error) {
      message.error('加载发布列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    form.resetFields();
    setModalVisible(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      await releaseAPI.create(values);
      message.success('创建发布任务成功');
      setModalVisible(false);
      loadReleases();
    } catch (error) {
      message.error('创建失败');
    }
  };

  const handleRollback = (release: Release) => {
    Modal.confirm({
      title: '确认回滚',
      content: `确定要回滚发布 ${release.version} 吗?`,
      onOk: async () => {
        try {
          await releaseAPI.rollback(release.id, 'previous', '手动回滚');
          message.success('回滚成功');
          loadReleases();
        } catch (error) {
          message.error('回滚失败');
        }
      },
    });
  };

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      pending_approval: { color: 'orange', text: '待审批' },
      deploying: { color: 'blue', text: '部署中' },
      completed: { color: 'green', text: '已完成' },
      failed: { color: 'red', text: '失败' },
      rolled_back: { color: 'default', text: '已回滚' },
    };
    const { color, text } = statusMap[status] || { color: 'default', text: status };
    return <Tag color={color}>{text}</Tag>;
  };

  const columns = [
    { title: '版本', dataIndex: 'version', key: 'version' },
    { title: '环境', dataIndex: 'environment', key: 'environment' },
    { title: '策略', dataIndex: 'strategy', key: 'strategy' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => getStatusTag(status),
    },
    { title: '发布人', dataIndex: 'scheduler', key: 'scheduler' },
    { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt' },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Release) => (
        <Space>
          <Button icon={<EyeOutlined />} size="small">详情</Button>
          {record.status === 'completed' && (
            <Button 
              icon={<RollbackOutlined />} 
              size="small" 
              danger 
              onClick={() => handleRollback(record)}
            >
              回滚
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          创建发布
        </Button>
      </div>
      <Table
        columns={columns}
        dataSource={releases}
        rowKey="id"
        loading={loading}
      />
      <Modal
        title="创建发布任务"
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="projectId" label="项目" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="applicationId" label="应用" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="version" label="版本" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="environment" label="环境" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="dev">开发环境</Select.Option>
              <Select.Option value="test">测试环境</Select.Option>
              <Select.Option value="staging">预发布环境</Select.Option>
              <Select.Option value="production">生产环境</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="strategy" label="发布策略" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="blue-green">蓝绿部署</Select.Option>
              <Select.Option value="canary">金丝雀发布</Select.Option>
              <Select.Option value="rolling-update">滚动更新</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ReleaseManagement;
