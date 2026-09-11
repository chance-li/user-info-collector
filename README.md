# 用户信息采集系统

一个简单的演示项目：访客在前台提交姓名和手机号，管理员在后台查看全部记录。

- 后端：Golang REST API + SQLite
- 前台：Vue 3 + Vite（用户提交页、管理列表页各一个应用）
- **管理页没有登录保护**，仅用于本地演示，请勿直接暴露到公网

## 目录结构

```text
backend/          Go API（SQLite 持久化）
frontend-user/   用户提交表单
frontend-admin/  管理员查看列表
```

## 环境要求

- Go 1.22+
- Node.js 18+（含 npm）

## 安装

```bash
# 后端依赖
cd backend
go mod tidy

# 用户端
cd ../frontend-user
npm install

# 管理端
cd ../frontend-admin
npm install
```

## 启动

请打开 **三个终端**，分别启动后端和两个前端。

### 1. 启动 Go API

```bash
cd backend
go run .
```

默认监听 `http://localhost:8080`。

- 数据库文件：`backend/data.db`（自动创建）
- 健康检查：`GET http://localhost:8080/api/health`
- 已开启 CORS，Vite 开发服务器可直接跨域调用 API

可选环境变量：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `ADDR` | `:8080` | 监听地址 |
| `DB_PATH` | `data.db` | SQLite 文件路径 |

### 2. 启动用户提交页

```bash
cd frontend-user
npm run dev
```

浏览器打开 [http://localhost:5173](http://localhost:5173)。

### 3. 启动管理列表页

```bash
cd frontend-admin
npm run dev
```

浏览器打开 [http://localhost:5174](http://localhost:5174)。

开发环境下，两个 Vite 应用会把 `/api` 代理到 `http://127.0.0.1:8080`。也可以设置 `VITE_API_BASE=http://localhost:8080` 让前端直连后端（后端已允许 CORS）。

## 接口说明

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/submissions` | 创建一条记录 |
| `GET` | `/api/submissions` | 列出全部记录（按时间倒序） |

创建请求体：

```json
{ "name": "张三", "phone": "13800138000" }
```

校验规则：

- 姓名：去空白后不能为空
- 手机号：中国大陆手机号，格式 `1` + `[3-9]` + 9 位数字（共 11 位）

## 端到端自测

1. 确认三个服务都已启动。
2. 打开用户页，填写姓名 `张三`、手机号 `13800138000`，点击「提交」，应看到成功提示。
3. 故意提交空姓名或错误手机号（如 `123`），应看到中文错误提示。
4. 打开管理页，应能看到刚才的记录（姓名、手机号、提交时间）。点击「刷新」可重新拉取。
5. 也可用 curl 验证 API：

```bash
curl -s -X POST http://localhost:8080/api/submissions \
  -H 'Content-Type: application/json' \
  -d '{"name":"李四","phone":"13900139000"}'

curl -s http://localhost:8080/api/submissions
```
