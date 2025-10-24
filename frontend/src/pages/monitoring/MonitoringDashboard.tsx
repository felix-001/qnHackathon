import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Select, message } from 'antd';
import { ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons';
import { monitoringAPI } from '@/services/api';
import type { MonitoringMetrics } from '@/types';

const MonitoringDashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<MonitoringMetrics | null>(null);
  const [selectedRelease, setSelectedRelease] = useState<string>('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (selectedRelease) {
      loadMetrics();
      const interval = setInterval(loadMetrics, 5000);
      return () => clearInterval(interval);
    }
  }, [selectedRelease]);

  const loadMetrics = async () => {
    if (!selectedRelease) return;
    setLoading(true);
    try {
      const response = await monitoringAPI.getRealtime(selectedRelease);
      setMetrics(response.data.metrics);
    } catch (error) {
      message.error('加载监控数据失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <div style={{ marginBottom: 24 }}>
        <Select
          style={{ width: 300 }}
          placeholder="选择发布任务"
          onChange={setSelectedRelease}
          value={selectedRelease}
        >
          <Select.Option value="release-1">Release v1.0.0</Select.Option>
          <Select.Option value="release-2">Release v1.1.0</Select.Option>
        </Select>
      </div>

      {metrics && (
        <Row gutter={16}>
          <Col span={6}>
            <Card>
              <Statistic
                title="请求速率 (req/s)"
                value={metrics.requestRate}
                prefix={<ArrowUpOutlined />}
                valueStyle={{ color: '#3f8600' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="错误率"
                value={metrics.errorRate * 100}
                precision={3}
                suffix="%"
                prefix={metrics.errorRate > 0.01 ? <ArrowUpOutlined /> : <ArrowDownOutlined />}
                valueStyle={{ color: metrics.errorRate > 0.01 ? '#cf1322' : '#3f8600' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="P50 延迟 (ms)"
                value={metrics.latencyP50}
                precision={0}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="P99 延迟 (ms)"
                value={metrics.latencyP99}
                precision={0}
              />
            </Card>
          </Col>
        </Row>
      )}

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={24}>
          <Card title="实时监控图表">
            <div style={{ height: 400, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              图表组件待集成 (可使用 ECharts 或 Recharts)
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default MonitoringDashboard;
