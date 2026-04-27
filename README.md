# Go Spring AI

一个受 Spring AI Alibaba 启发的 Go 语言 AI 应用开发库，提供统一的 API 接口来构建 AI 驱动的应用程序。本库提供与模型无关的抽象层，支持聊天补全、文本嵌入、图像生成、音频处理、RAG（检索增强生成）、工具调用等功能。

## 特性

- **模型无关 API**：最小化代码改动即可切换不同的 AI 服务提供商
- **聊天客户端**：流畅的 API 支持流式对话交互
- **嵌入模型**：生成文本向量嵌入
- **向量存储**：支持相似度搜索的内存向量存储
- **RAG 支持**：检索增强生成，包含文档处理管道
- **工具/函数调用**：支持模型调用外部函数
- **提示词模板**：支持变量替换的动态提示词生成
- **结构化输出**：将模型输出解析为结构化类型
- **对话记忆**：对话历史管理
- **顾问模式**：拦截和修改请求/响应（日志、RAG 等）
- **多模态支持**：图像生成、视频生成、音频处理
- **Provider 注册中心**：统一的 AI 服务提供商管理和切换

## 安装

```bash
go get github.com/go-spring/ai
```

## 快速开始

### 基础聊天

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/go-spring/ai/core"
    "github.com/go-spring/ai/openai"
)

func main() {
    client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
    chatModel := openai.NewChatModel(client)
    chatClient := core.NewChatClient(chatModel)

    ctx := context.Background()
    response, err := chatClient.Prompt().
        User("法国的首都是哪里？").
        CallWithContext(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(response.Content)
}
```

### 使用智谱 AI 异步图像/视频生成

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/go-spring/ai/core"
    "github.com/go-spring/ai/zhipu"
)

func main() {
    // 创建智谱 Provider
    provider := zhipu.NewProvider(os.Getenv("ZHIPU_API_KEY"))
    
    // 创建媒体生成服务
    service := core.NewMediaGenerationService(provider)
    
    ctx := context.Background()
    
    // 生成图像
    imageTask, err := service.SubmitImageTask(ctx, "一只可爱的小猫咪", core.AsyncImageOptions{
        Size:    "1024x1024",
        Quality: "hd",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("图像任务已提交：%s\n", imageTask.ID)
    
    // 等待图像生成完成
    imageResult, err := service.WaitForTask(ctx, imageTask.ID, core.PollOptions{
        Interval:   5,
        MaxRetries: 60,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("图像生成完成：%v\n", imageResult.ImageResults)
    
    // 生成视频
    videoTask, err := service.SubmitVideoTask(ctx, "A cat is playing", core.AsyncVideoOptions{
        Model:     "cogvideox-3",
        Quality:   "quality",
        Duration:  5,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("视频任务已提交：%s\n", videoTask.ID)
    
    // 等待视频生成完成
    videoResult, err := service.WaitForTask(ctx, videoTask.ID, core.PollOptions{
        Interval:   10,
        MaxRetries: 60,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("视频生成完成：%v\n", videoResult.VideoResults)
}
```

### 使用 Provider 注册中心切换不同服务商

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/go-spring/ai/core"
    "github.com/go-spring/ai/zhipu" // 智谱 AI
    // _ "github.com/go-spring/ai/jimeng" // 即梦 AI（未来支持）
    // _ "github.com/go-spring/ai/hunyuan" // 腾讯混元（未来支持）
)

func main() {
    // 获取注册中心
    registry := core.GetProviderRegistry()
    
    // 注册智谱 AI（通常在 init() 中自动注册）
    _ = registry.Register("zhipu", func(config core.ProviderConfig) (core.AsyncMediaProvider, error) {
        return zhipu.NewProvider(config.APIKey), nil
    })
    
    // 创建智谱媒体生成服务
    provider, err := registry.CreateProvider("zhipu", core.ProviderConfig{
        APIKey: "your-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    service := core.NewMediaGenerationService(provider)
    
    ctx := context.Background()
    result, err := service.GenerateImage(ctx, "美丽的风景", core.AsyncImageOptions{})
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("图像生成成功：%v\n", result.ImageResults)
}
```

### RAG 文档问答

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/go-spring/ai/core"
    "github.com/go-spring/ai/openai"
)

func main() {
    // 创建聊天模型
    chatModel := openai.NewChatModel(openai.NewClient(os.Getenv("OPENAI_API_KEY")))
    
    // 创建嵌入模型
    embeddingModel := openai.NewEmbeddingModel(openai.NewClient(os.Getenv("OPENAI_API_KEY")))
    
    // 创建向量存储
    vectorStore := core.NewMemoryVectorStore(embeddingModel)
    
    // 创建 RAG 管道
    ragPipeline := core.NewRagPipeline(chatModel, vectorStore)
    
    // 添加文档
    docs := []core.Document{
        {
            ID:      "doc1",
            Content: "Go 语言是一门静态类型、编译型语言，由 Google 开发。",
        },
        {
            ID:      "doc2",
            Content: "Spring AI Alibaba 是阿里巴巴开源的 AI 应用开发框架。",
        },
    }
    _ = vectorStore.AddDocuments(context.Background(), docs)
    
    // 执行 RAG 查询
    ctx := context.Background()
    response, err := ragPipeline.Query(ctx, "Go 语言是什么？", core.RagOptions{
        TopK: 2,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(response)
}
```

### 工具调用

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/go-spring/ai/core"
    "github.com/go-spring/ai/openai"
)

// 定义工具
func getWeather(city string) string {
    return fmt.Sprintf("%s 的天气是晴天，温度 25°C", city)
}

func main() {
    client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
    chatModel := openai.NewChatModel(client)
    
    // 注册工具
    toolRegistry := core.NewToolRegistry()
    toolRegistry.RegisterTool("get_weather", "查询天气", getWeather)
    
    // 创建带工具的聊天客户端
    chatClient := core.NewChatClient(chatModel, core.WithTools(toolRegistry.GetTools()...))
    
    ctx := context.Background()
    response, err := chatClient.Prompt().
        User("北京天气怎么样？").
        CallWithContext(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(response.Content)
}
```

## 支持的 AI 服务提供商

### 已实现

- **智谱 AI**：聊天补全、图像生成、视频生成、异步任务处理
- **OpenAI**：聊天补全、嵌入、图像生成、音频处理

### 即将支持

- **即梦 AI**：图像生成、视频生成
- **腾讯混元**：聊天补全、图像生成
- **其他主流 AI 平台**

## 核心架构

### 分层设计

```
┌─────────────────────────────────────────────┐
│           应用层（Application）              │
│  ChatClient, RagPipeline, MediaService     │
├─────────────────────────────────────────────┤
│           抽象层（Abstraction）              │
│  ChatModel, EmbeddingModel,                │
│  AsyncMediaProvider, VectorStore           │
├─────────────────────────────────────────────┤
│           实现层（Implementation）           │
│  OpenAI, Zhipu, Hunyuan, Jimeng, ...       │
└─────────────────────────────────────────────┘
```

### Provider 模式

```
┌─────────────────────────────────────────────┐
│           MediaGenerationService            │  ← 服务层：统一入口
├─────────────────────────────────────────────┤
│          AsyncMediaProvider 接口            │  ← 抽象层：跨厂商标准
│  ├── Name() → string                        │
│  ├── ImageModel() → AsyncImageModel         │
│  ├── VideoModel() → AsyncVideoModel         │
│  └── TaskQuery() → AsyncTaskQuery           │
├─────────────────────────────────────────────┤
│  zhipu.Provider  │  jimeng.Provider  │ ...  │  ← 实现层：各厂商适配
└─────────────────────────────────────────────┘
```

## 项目结构

```
GoAi/
├── core/                    # 核心抽象层和通用实现
│   ├── model.go            # 核心模型接口定义
│   ├── provider.go         # Provider 注册中心和服务层
│   ├── chatclient.go       # 聊天客户端
│   ├── memory.go           # 对话记忆
│   ├── rag.go              # RAG 管道
│   ├── vectorstore.go      # 向量存储
│   ├── tool.go             # 工具调用
│   └── test/               # 单元测试
├── openai/                  # OpenAI 实现
├── zhipu/                   # 智谱 AI 实现
│   ├── provider.go         # Provider 实现
│   ├── client.go           # API 客户端
│   ├── image.go            # 图像生成
│   ├── video.go            # 视频生成
│   └── task.go             # 异步任务处理
├── examples/                # 示例代码
│   ├── basic_chat/         # 基础聊天示例
│   ├── memory/             # 对话记忆示例
│   ├── rag/                # RAG 示例
│   ├── tools/              # 工具调用示例
│   └── zhipu_async/        # 智谱异步媒体生成示例
└── doc/                     # 文档
```

## 开发指南

### 添加新的 AI 服务提供商

1. 创建新包（如 `hunyuan/`）
2. 实现 `AsyncMediaProvider` 接口
3. 在 `init()` 中注册工厂函数
4. 编写单元测试

```go
package hunyuan

import (
    "github.com/go-spring/ai/core"
)

const ProviderName = "hunyuan"

type Provider struct {
    // ...
}

func NewProvider(apiKey string) *Provider {
    // ...
}

func (p *Provider) Name() string {
    return ProviderName
}

func (p *Provider) ImageModel() core.AsyncImageModel {
    // ...
}

func (p *Provider) VideoModel() core.AsyncVideoModel {
    // ...
}

func (p *Provider) TaskQuery() core.AsyncTaskQuery {
    // ...
}

func init() {
    _ = core.RegisterProvider(ProviderName, func(config core.ProviderConfig) (core.AsyncMediaProvider, error) {
        return NewProvider(config.APIKey), nil
    })
}
```

## 测试

运行所有测试：

```bash
go test ./...
```

运行特定包的测试：

```bash
go test ./zhipu/... -v
go test ./core/test/... -v
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 致谢

- [Spring AI Alibaba](https://github.com/alibaba/spring-ai-alibaba) - 设计灵感来源
- [LangChain](https://github.com/langchain-ai/langchain) - RAG 和工具调用设计参考
- [智谱 AI](https://open.bigmodel.cn/) - 异步 API 文档支持
