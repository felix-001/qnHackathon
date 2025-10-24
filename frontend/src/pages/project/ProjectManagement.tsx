import React, { useState, useEffect } from 'react';
import { Table, Button, Modal, Form, Input, Select, message, Space } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { projectAPI } from '@/services/api';
import type { Project } from '@/types';

const ProjectManagement: React.FC = () => {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    loadProjects();
  }, []);

  const loadProjects = async () => {
    setLoading(true);
    try {
      const response = await projectAPI.list();
      setProjects(response.data);
    } catch (error) {
      message.error('加载项目列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingProject(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (project: Project) => {
    setEditingProject(project);
    form.setFieldsValue(project);
    setModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await projectAPI.delete(id);
      message.success('删除成功');
      loadProjects();
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editingProject) {
        await projectAPI.update(editingProject.id, values);
        message.success('更新成功');
      } else {
        await projectAPI.create(values);
        message.success('创建成功');
      }
      setModalVisible(false);
      loadProjects();
    } catch (error) {
      message.error('操作失败');
    }
  };

  const columns = [
    { title: '项目名称', dataIndex: 'name', key: 'name' },
    { title: '项目代码', dataIndex: 'code', key: 'code' },
    { title: '负责人', dataIndex: 'owner', key: 'owner' },
    { title: '构建工具', dataIndex: 'buildTool', key: 'buildTool' },
    { title: '部署类型', dataIndex: 'deploymentType', key: 'deploymentType' },
    { title: '状态', dataIndex: 'status', key: 'status' },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Project) => (
        <Space>
          <Button icon={<EditOutlined />} onClick={() => handleEdit(record)}>编辑</Button>
          <Button icon={<DeleteOutlined />} danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          创建项目
        </Button>
      </div>
      <Table
        columns={columns}
        dataSource={projects}
        rowKey="id"
        loading={loading}
      />
      <Modal
        title={editingProject ? '编辑项目' : '创建项目'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="项目名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="code" label="项目代码" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea />
          </Form.Item>
          <Form.Item name="owner" label="负责人" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="repositoryUrl" label="仓库地址">
            <Input />
          </Form.Item>
          <Form.Item name="buildTool" label="构建工具">
            <Select>
              <Select.Option value="maven">Maven</Select.Option>
              <Select.Option value="gradle">Gradle</Select.Option>
              <Select.Option value="npm">NPM</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="deploymentType" label="部署类型">
            <Select>
              <Select.Option value="kubernetes">Kubernetes</Select.Option>
              <Select.Option value="docker">Docker</Select.Option>
              <Select.Option value="vm">VM</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ProjectManagement;
