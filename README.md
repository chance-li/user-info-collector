# 用户信息采集系统

一个简单的演示项目：访客在前台提交姓名和手机号，管理员在后台查看全部记录。

- 后端：Golang REST API + SQLite
- 前台：Vue 3 + Vite（用户提交页、管理列表页各一个应用）
- **管理页没有登录保护**，仅用于演示。本地和线上都不要当作真实后台使用，请勿直接暴露到公网处理敏感数据。

## 目录结构

```text
backend/          Go API（SQLite 持久化，含 Dockerfile）
frontend-user/   用户提交表单（Vercel）
frontend-admin/  管理员查看列表（Vercel）
render.yaml      Render Blueprint（可选）
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
| `PORT` | （未设置） | 若设置则监听 `:{PORT}`（Render 会注入此变量，优先级高于 `ADDR`） |
| `ADDR` | `:8080` | 监听地址（仅在未设置 `PORT` 时生效） |
| `DB_PATH` | `data.db` | SQLite 文件路径 |
| `CORS_ORIGINS` | 未设置（允许 `*`） | 逗号分隔的允许来源，例如两个 Vercel 地址。未设置时保持演示用宽松 CORS |

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

开发环境下，两个 Vite 应用会把 `/api` 代理到 `http://127.0.0.1:8080`。也可以设置 `VITE_API_BASE=http://localhost:8080` 让前端直连后端。线上构建必须设置 `VITE_API_BASE` 为 Render 地址（见下方部署说明）。

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

## 线上部署（Render + Vercel）

以下步骤使用平台免费域名（`*.onrender.com`、`*.vercel.app`），暂不绑定自定义域名。

**警告：管理页没有任何登录/鉴权，仅供演示。** 部署到公网后任何人打开管理地址都能看到全部提交记录。不要用于真实用户数据。

仓库内已包含 `backend/Dockerfile`、根目录 `render.yaml`，以及两个前端的 `vercel.json`。下面只说明如何自己创建服务，并不代表仓库已经连上了线上实例。

### 1. 合并并推送到 `main`

将代码合并进 GitHub 仓库的 `main`（或你准备连接的分支）并推送，Render / Vercel 才能拉到最新提交。

### 2. 在 Render 部署 Go 后端

1. 打开 [Render](https://render.com)，用 GitHub 登录，**New → Web Service**，选择本仓库。
2. 设置：
   - **Root Directory**：`backend`
   - **Runtime**：Docker（使用仓库里的 `backend/Dockerfile`）
   - **Health Check Path**：`/api/health`
3. 也可改用 Blueprint：在 Render 选择 **New → Blueprint**，指向仓库根目录的 `render.yaml`。
4. Render 会设置 `PORT`，容器会监听 `:{PORT}`，无需再配 `ADDR`。
5. 部署完成后记下免费域名，形如 `https://user-info-api.onrender.com`（**不要末尾斜杠**）。免费实例可能冷启动较慢，第一次请求 `/api/health` 可能要等几十秒。

**SQLite 与磁盘（重要）：**

Render 免费 Web Service 的本地磁盘是**临时的**，重新部署或实例休眠后，`data.db` 可能丢失。

- **推荐（数据要留下来）**：给服务挂一块 Render **Disk**，例如：
  - Mount Path：`/data`
  - 环境变量：`DB_PATH=/data/data.db`
  - Disk 通常需要付费实例；免费套餐请以 Render 当前说明为准。
- **仅演示**：不挂 Disk，使用默认 `DB_PATH=data.db`。数据可能在 redeploy 后清空，这是预期行为。

### 3. 在 Vercel 部署两个前端

从**同一个 GitHub 仓库**创建两个 Vercel 项目（Framework 选 Vite），分别设置 Root Directory 与构建时环境变量：

| 项目用途 | Root Directory | 环境变量 |
| --- | --- | --- |
| 用户提交页 | `frontend-user` | `VITE_API_BASE=https://<render-service>.onrender.com` |
| 管理列表页 | `frontend-admin` | 同上（**不要末尾斜杠**） |

`VITE_API_BASE` 是 **Vite 构建时** 写入前端的，改完必须重新 Deploy 才会生效。本地开发仍可用 Vite 代理，不必设该变量。

部署完成后记下两个免费域名，例如：

- `https://frontend-user-xxx.vercel.app`
- `https://frontend-admin-xxx.vercel.app`

### 4. 回写 CORS 并重新部署后端

在 Render 服务的 Environment 中设置：

```text
CORS_ORIGINS=https://frontend-user-xxx.vercel.app,https://frontend-admin-xxx.vercel.app
```

不要末尾斜杠。保存后 **Redeploy** 后端，使新环境变量生效。未设置 `CORS_ORIGINS` 时后端仍允许 `*`（方便本地演示）；线上应改成上面两个 Vercel 地址。

### 5. 冒烟测试

1. 打开用户提交页（Vercel），填写姓名和合法手机号并提交，应显示成功。
2. 打开管理列表页（Vercel），应能看到刚提交的记录；点「刷新」可重新拉取。
3. 若浏览器控制台出现 CORS 错误，检查 `CORS_ORIGINS` 是否与 Vercel 地址完全一致（含 `https://`），以及后端是否已 Redeploy。
4. 若提交失败且 Render 刚唤醒，等实例启动后再试（免费实例会休眠）。
