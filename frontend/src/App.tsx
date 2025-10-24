import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { Layout, Menu } from 'antd';
import {
  ProjectOutlined,
  RocketOutlined,
  DashboardOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import ProjectManagement from '@/pages/project/ProjectManagement';
import ReleaseManagement from '@/pages/release/ReleaseManagement';
import MonitoringDashboard from '@/pages/monitoring/MonitoringDashboard';
import ConfigManagement from '@/pages/config/ConfigManagement';

const { Header, Content, Sider } = Layout;

const App: React.FC = () => {
  return (
    <Router>
      <Layout style={{ minHeight: '100vh' }}>
        <Header style={{ color: 'white', fontSize: 20, fontWeight: 'bold' }}>
          智能发布控制台
        </Header>
        <Layout>
          <Sider width={200} style={{ background: '#fff' }}>
            <Menu
              mode="inline"
              defaultSelectedKeys={['1']}
              style={{ height: '100%', borderRight: 0 }}
            >
              <Menu.Item key="1" icon={<ProjectOutlined />}>
                <Link to="/">项目管理</Link>
              </Menu.Item>
              <Menu.Item key="2" icon={<RocketOutlined />}>
                <Link to="/releases">发布管理</Link>
              </Menu.Item>
              <Menu.Item key="3" icon={<DashboardOutlined />}>
                <Link to="/monitoring">监控面板</Link>
              </Menu.Item>
              <Menu.Item key="4" icon={<SettingOutlined />}>
                <Link to="/config">配置管理</Link>
              </Menu.Item>
            </Menu>
          </Sider>
          <Layout style={{ padding: '24px' }}>
            <Content
              style={{
                padding: 24,
                margin: 0,
                minHeight: 280,
                background: '#fff',
              }}
            >
              <Routes>
                <Route path="/" element={<ProjectManagement />} />
                <Route path="/releases" element={<ReleaseManagement />} />
                <Route path="/monitoring" element={<MonitoringDashboard />} />
                <Route path="/config" element={<ConfigManagement />} />
              </Routes>
            </Content>
          </Layout>
        </Layout>
      </Layout>
    </Router>
  );
};

export default App;
