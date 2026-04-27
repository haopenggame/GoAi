# 智谱AI 异步接口文档

> 来源：[智谱AI开放文档](https://docs.bigmodel.cn)  
> 整理时间：2026-04-25

---

## 目录

1. [图像生成（异步）](#一图像生成异步)
2. [视频生成（异步）](#二视频生成异步)
3. [查询异步结果](#三查询异步结果)

---

## 一、图像生成（异步）

> **POST** `https://open.bigmodel.cn/api/paas/v4/async/images/generations`

### 认证

| 参数 | 位置 | 必填 | 说明 |
|------|------|------|------|
| `Authorization` | header | ✅ 是 | 使用 `Bearer <token>` 格式 |

### 请求体参数（Body · application/json）

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| `model` | enum | ✅ 是 | — | 模型编码，仅支持 `glm-image` |
| `prompt` | string | ✅ 是 | — | 所需图像的文本描述 |
| `quality` | enum | 否 | `hd` | 图像质量。`hd`：精细细节更丰富，整体一致性更高，耗时约 20 秒 |
| `size` | string | 否 | `1280x1280` | 图片尺寸，推荐值见下表。自定义时长宽需在 1024px～2048px 范围内，最大像素不超过 2²²px，且长宽均需为 32 的整数倍 |
| `watermark_enabled` | boolean | 否 | `true` | 是否添加水印。`false` 需签署免责声明（路径：个人中心 → 安全管理 → 去水印管理） |
| `user_id` | string | 否 | — | 终端用户唯一 ID，长度 6～128 个字符 |

#### size 推荐枚举值

| 值 | 方向 |
|----|------|
| `1280x1280` | 正方形（默认） |
| `1568x1056` | 横向 |
| `1056x1568` | 纵向 |
| `1472x1088` | 横向 |
| `1088x1472` | 纵向 |
| `1728x960`  | 横向宽屏 |
| `960x1728`  | 纵向长屏 |

### 响应（200 OK）

```json
{
  "model": "<string>",
  "id": "<string>",
  "request_id": "<string>",
  "task_status": "<string>"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `model` | string | 本次调用使用的模型名称 |
| `id` | string | 任务 ID，查询结果时使用 |
| `request_id` | string | 任务编号（客户端提交或平台生成） |
| `task_status` | string | 处理状态：`PROCESSING`（处理中）/ `SUCCESS`（成功）/ `FAIL`（失败）|

> ⚠️ 最终图片结果需通过 **查询异步结果** 接口获取。

### cURL 示例

```bash
curl --request POST \
  --url https://open.bigmodel.cn/api/paas/v4/async/images/generations \
  --header 'Authorization: Bearer <token>' \
  --header 'Content-Type: application/json' \
  --data '{
    "model": "glm-image",
    "prompt": "一只可爱的小猫咪，坐在阳光明媚的窗台上，背景是蓝天白云.",
    "size": "1280x1280"
  }'
```

---

## 二、视频生成（异步）

> **POST** `https://open.bigmodel.cn/api/paas/v4/videos/generations`

### 认证

| 参数 | 位置 | 必填 | 说明 |
|------|------|------|------|
| `Authorization` | header | ✅ 是 | 使用 `Bearer <token>` 格式 |

### 请求体参数（Body · application/json）

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| `model` | enum | ✅ 是 | — | 模型编码，仅支持 `cogvideox-3` |
| `prompt` | string | ✅ 是* | — | 视频文本描述，最多 512 个字符。`image_url` 与 `prompt` 不能同时为空 |
| `quality` | enum | 否 | `speed` | 生成模式：`speed`（速度优先）/ `quality`（质量优先） |
| `with_audio` | boolean | 否 | `false` | 是否生成 AI 音效 |
| `watermark_enabled` | boolean | 否 | `true` | 是否添加水印（`false` 需签署免责声明） |
| `image_url` | string / array | 否 | — | 图片 URL 或 Base64，支持 `.png`、`.jpeg`、`.jpg`，大小 ≤ 5M；支持首尾帧（传入两张图片） |
| `size` | enum | 否 | 短边 1080 | 视频分辨率，最高支持 4K，见下表 |
| `fps` | integer | 否 | `30` | 帧率，可选 `30` 或 `60` |
| `duration` | integer | 否 | `5` | 视频时长（秒），可选 `5` 或 `10` |
| `request_id` | string | 否 | — | 客户端提供的唯一标识符 |
| `user_id` | string | 否 | — | 终端用户 ID，长度 6～128 个字符 |

#### size 枚举值

| 值 | 说明 |
|----|------|
| `1280x720` | 720p 横向 |
| `720x1280` | 720p 纵向 |
| `1024x1024` | 1:1 正方形 |
| `1920x1080` | 1080p 横向 |
| `1080x1920` | 1080p 纵向 |
| `2048x1080` | 2K 横向 |
| `3840x2160` | 4K 超高清 |

### 响应（200 OK）

```json
{
  "model": "<string>",
  "id": "<string>",
  "request_id": "<string>",
  "task_status": "<string>"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `model` | string | 本次调用使用的模型名称 |
| `id` | string | 任务 ID，查询结果时使用 |
| `request_id` | string | 任务编号 |
| `task_status` | string | 处理状态：`PROCESSING` / `SUCCESS` / `FAIL` |

> ⚠️ 最终视频结果需通过 **查询异步结果** 接口获取。

### cURL 示例

```bash
curl --request POST \
  --url https://open.bigmodel.cn/api/paas/v4/videos/generations \
  --header 'Authorization: Bearer <token>' \
  --header 'Content-Type: application/json' \
  --data '{
    "model": "cogvideox-3",
    "prompt": "A cat is playing with a ball.",
    "quality": "quality",
    "with_audio": true,
    "size": "1920x1080",
    "fps": 30
  }'
```

---

## 三、查询异步结果

> **GET** `https://open.bigmodel.cn/api/paas/v4/async-result/{id}`  
> 用于查询对话补全、图像生成、视频生成等异步请求的处理结果和状态。

### 认证

| 参数 | 位置 | 必填 | 说明 |
|------|------|------|------|
| `Authorization` | header | ✅ 是 | 使用 `Bearer <token>` 格式 |

### 路径参数（Path Parameters）

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `id` | string | ✅ 是 | 任务 ID（由图像生成/视频生成/对话补全异步接口返回的 `id` 字段） |

### cURL 示例

```bash
curl --request GET \
  --url https://open.bigmodel.cn/api/paas/v4/async-result/{id} \
  --header 'Authorization: Bearer <token>'
```

### 响应（200 OK）

#### 对话补全 / 通用响应结构

```json
{
  "id": "<string>",
  "request_id": "<string>",
  "created": 123,
  "model": "<string>",
  "task_status": "<string>",
  "choices": [...],
  "usage": {...},
  "video_result": [...],
  "web_search": [...],
  "content_filter": [...]
}
```

#### 顶层字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 任务 ID |
| `request_id` | string | 请求 ID |
| `created` | integer | 请求创建时间（Unix 时间戳，单位：秒） |
| `model` | string | 模型名称 |
| `choices` | object[] | 模型响应列表（对话补全场景使用） |
| `usage` | object | Token 使用统计 |
| `video_result` | object[] | 视频生成结果 |
| `web_search` | object[] | 网页搜索相关信息（使用 WebSearchToolSchema 时返回） |
| `content_filter` | object[] | 内容安全相关信息 |

#### `choices` 子字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `choices[].index` | integer | 结果索引 |
| `choices[].finish_reason` | string | 推理终止原因：`stop`（自然结束/触发 stop 词）/ `tool_calls`（命中函数）/ `length`（达到 token 限制）/ `sensitive`（安全审核拦截）/ `network_error`（推理异常）/ `model_context_window_exceeded`（超出上下文窗口） |
| `choices[].message.role` | string | 对话角色，默认 `assistant` |
| `choices[].message.content` | string / object[] | 对话文本内容；调用函数时为 `null` |
| `choices[].message.reasoning_content` | string | 思维链内容，仅 `glm-4.5` 系列、`glm-4.1v-thinking` 系列返回 |
| `choices[].message.audio` | object | 音频内容，仅 `glm-4-voice` 模型返回 |
| `choices[].message.tool_calls` | object[] | 应被调用的函数名称和参数 |

#### `usage` 子字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `usage.prompt_tokens` | number | 用户输入的 Token 数量 |
| `usage.completion_tokens` | number | 输出的 Token 数量 |
| `usage.total_tokens` | integer | Token 总数（`glm-4-voice` 模型：1 秒音频 = 12.5 Token，向上取整） |
| `usage.prompt_tokens_details.cached_tokens` | number | 命中缓存的 Token 数量 |

#### `video_result` 子字段（视频生成场景）

| 字段 | 类型 | 说明 |
|------|------|------|
| `video_result[].url` | string | 视频链接 |
| `video_result[].cover_image_url` | string | 视频封面图片链接 |

#### `web_search` 子字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `web_search[].icon` | string | 来源网站图标 |
| `web_search[].title` | string | 搜索结果标题 |
| `web_search[].link` | string | 搜索结果网页链接 |
| `web_search[].media` | string | 媒体来源名称 |
| `web_search[].publish_date` | string | 网页发布时间 |
| `web_search[].content` | string | 引用的文本内容 |
| `web_search[].refer` | string | 角标序号 |

#### `content_filter` 子字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `content_filter[].role` | string | 安全生效环节：`assistant`（模型推理）/ `user`（用户输入）/ `history`（历史上下文） |
| `content_filter[].level` | integer | 严重程度 0～3（0 最严重，3 轻微） |

---

## 典型使用流程

```
1. 调用「图像生成(异步)」或「视频生成(异步)」接口
       ↓ 返回 task_status: PROCESSING + id
2. 轮询「查询异步结果」接口（传入 id）
       ↓ task_status: PROCESSING → 继续等待
       ↓ task_status: SUCCESS    → 从 video_result / choices 取结果
       ↓ task_status: FAIL       → 处理失败，查看错误信息
```

---

*文档来源：[智谱AI开放文档](https://docs.bigmodel.cn) · 整理时间：2026-04-25*
