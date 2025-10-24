import React, { useState } from 'react';
import { Table, Button, Modal, Form, Input, Select, message, Space, Tag } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';

interface Config {
  id: string;
  key: string;
  value: string;
  environment: string;
  description: string;
}

const ConfigManagement: React.FC = () => {
  const [configs, setConfigs] = useState<Config[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingConfig, setEditingConfig] = useState<Config | null>(null);
  const [form] = Form.useForm();

  const handleCreate = () => {
    setEditingConfig(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (config: Config) => {
    setEditingConfig(config);
    form.setFieldsValue(config);
    setModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    message.success('删除成功');
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      message.success(editingConfig ? '更新成功' : '创建成功');
      setModalVisible(false);
    } catch (error) {
      message.error('操作失败');
    }
  };

  const columns = [
    { title: '配置键', dataIndex: 'key', key: 'key' },
    { title: '配置值', dataIndex: 'value', key: 'value' },
    {
      title: '环境',
      dataIndex: 'environment',
      key: 'environment',
      render: (env: string) => <Tag>{env}</Tag>,
    },
    { title: '描述', dataIndex: 'description', key: 'description' },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Config) => (
        <Space>
          <Button icon={<EditOutlined />} size="small" onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Button icon={<DeleteOutlined />} size="small" danger onClick={() => handleDelete(record.id)}>
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          添加配置
        </Button>
      </div>
      <Table
        columns={columns}
        dataSource={configs}
        rowKey="id"
        loading={loading}
      />
      <Modal
        title={editingConfig ? '编辑配置' : '添加配置'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="key" label="配置键" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="value" label="配置值" rules={[{ required: true }]}>
            <Input.TextArea />
          </Form.Item>
          <Form.Item name="environment" label="环境" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="dev">开发环境</Select.Option>
              <Select.Option value="test">测试环境</Select.Option>
              <Select.Option value="staging">预发布环境</Select.Option>
              <Select.Option value="production">生产环境</Select.Option>
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

export default ConfigManagement;
