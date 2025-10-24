# 智能发布控制台 - 前端

基于 React + TypeScript + Ant Design 的智能发布控制台前端应用。

## 技术栈

- **框架**: React 18
- **语言**: TypeScript
- **UI 库**: Ant Design 5
- **路由**: React Router 6
- **构建工具**: Vite
- **HTTP 客户端**: Axios

## 项目结构

```
frontend/
├── public/              # 静态资源
├── src/
│   ├── pages/          # 页面组件
│   │   ├── project/    # 项目管理
│   │   ├── release/    # 发布管理
│   │   ├── monitoring/ # 监控面板
│   │   └── config/     # 配置管理
│   ├── components/     # 公共组件
│   ├── services/       # API 服务
│   ├── types/          # TypeScript 类型定义
│   ├── utils/          # 工具函数
│   ├── App.tsx         # 应用根组件
│   └── main.tsx        # 应用入口
├── package.json
├── tsconfig.json
└── vite.config.ts
```

## 功能模块

### 1. 项目管理
- 项目列表展示
- 创建/编辑/删除项目
- 项目配置管理

### 2. 发布管理
- 发布任务创建
- 发布流程监控
- 发布历史查询
- 一键回滚

### 3. 监控面板
- 实时性能指标
- 请求成功率
- 响应时间监控
- 错误率统计

### 4. 配置管理
- 应用配置管理
- 多环境配置
- 配置版本控制

## 开发指南

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run dev
```

访问 http://localhost:3000

### 构建生产版本

```bash
npm run build
```

### 代码检查

```bash
npm run lint
```

## API 接口

前端通过 `/api/v1` 代理访问后端 API 服务 (默认 http://localhost:8080)。

主要接口:
- `/api/v1/projects` - 项目管理
- `/api/v1/applications` - 应用管理
- `/api/v1/releases` - 发布管理
- `/api/v1/monitoring` - 监控数据
- `/api/v1/approvals` - 审批流程

## 开发计划

- [ ] 集成图表库 (ECharts/Recharts)
- [ ] 添加用户认证
- [ ] 实现 WebSocket 实时推送
- [ ] 添加国际化支持
- [ ] 完善单元测试
