package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-spring/ai/core"
)

func main() {
	fmt.Println("=== Go Spring AI - MCP 主控程序示例 ===")
	fmt.Println()

	exampleBasicController()
	fmt.Println()
	exampleTaskManager()
	fmt.Println()
	exampleResourceManager()
	fmt.Println()
	examplePipeline()
	fmt.Println()
	exampleMiddleware()
	fmt.Println()
	exampleFullSystem()
}

// 示例1：基础控制器使用
func exampleBasicController() {
	fmt.Println("--- 示例1：基础控制器创建与启停 ---")

	mcp := core.NewMCPController(2)
	mcp.Start()

	status := mcp.Status()
	fmt.Printf("控制器运行状态: %v\n", status.Running)

	mcp.Stop()
	fmt.Printf("控制器已停止\n")
}

// 示例2：任务管理器
func exampleTaskManager() {
	fmt.Println("--- 示例2：任务管理器 ---")

	mcp := core.NewMCPController(3)

	err := mcp.TaskManager().RegisterHandler("data-process", func(ctx context.Context, task *core.Task) error {
		data, ok := task.Input["data"].(string)
		if !ok {
			return fmt.Errorf("无效的输入数据")
		}
		task.Output = map[string]interface{}{
			"result": "处理完成: " + data,
			"length": len(data),
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	err = mcp.TaskManager().RegisterHandler("file-download", func(ctx context.Context, task *core.Task) error {
		url, ok := task.Input["url"].(string)
		if !ok {
			return fmt.Errorf("缺少URL参数")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}

		task.Output = map[string]interface{}{
			"file": url,
			"size": 1024,
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	mcp.Start()
	defer mcp.Stop()

	task1 := &core.Task{
		ID:       "task-001",
		Name:     "数据处理任务",
		Type:     "data-process",
		Priority: 1,
		Input:    map[string]interface{}{"data": "Hello MCP"},
	}

	err = mcp.SubmitTask(task1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("任务 %s 已提交\n", task1.ID)

	time.Sleep(200 * time.Millisecond)

	t1, ok := mcp.TaskManager().GetTask("task-001")
	if ok {
		fmt.Printf("任务状态: %s, 输出: %v\n", t1.State, t1.Output)
	}

	stats := mcp.TaskManager().Stats()
	fmt.Printf("任务统计: 总计=%d, 完成=%d, 失败=%d\n", stats.Total, stats.Completed, stats.Failed)
}

// 示例3：资源管理器
func exampleResourceManager() {
	fmt.Println("--- 示例3：资源管理器 ---")

	rm := core.NewResourceManager()

	rm.Register("cpu", 8, map[string]interface{}{"type": "intel-core-i7"})
	rm.Register("memory", 16384, map[string]interface{}{"type": "ddr4"})
	rm.Register("gpu", 1, map[string]interface{}{"type": "nvidia-rtx4090"})
	rm.Register("disk-io", 100, map[string]interface{}{"unit": "iops"})

	fmt.Printf("已注册资源:\n")
	for _, r := range rm.ListResources() {
		fmt.Printf("  %-10s 容量=%-8d 可用=%d\n", r.Name, r.Capacity, r.Available)
	}

	processes := []struct {
		id  string
		cpu int
		mem int
		gpu int
	}{
		{"worker-001", 2, 2048, 0},
		{"worker-002", 4, 4096, 1},
		{"worker-003", 1, 1024, 0},
	}

	for _, p := range processes {
		if err := rm.Allocate("cpu", p.id, p.cpu); err != nil {
			fmt.Printf("分配 CPU 失败 (%s): %v\n", p.id, err)
		}
		if err := rm.Allocate("memory", p.id, p.mem); err != nil {
			fmt.Printf("分配内存失败 (%s): %v\n", p.id, err)
		}
		if p.gpu > 0 {
			if err := rm.Allocate("gpu", p.id, p.gpu); err != nil {
				fmt.Printf("分配 GPU 失败 (%s): %v\n", p.id, err)
			}
		}
	}

	fmt.Printf("\n资源分配后:\n")
	for _, r := range rm.ListResources() {
		fmt.Printf("  %-10s 使用=%-4d/%-4d 剩余=%d\n", r.Name, r.Used, r.Capacity, r.Available)
	}

	rm.Release("memory", "worker-002", 2048)
	state, _ := rm.GetState("memory")
	fmt.Printf("\n释放 worker-002 部分内存后: 内存使用=%d/%d\n", state.Used, state.Capacity)
}

// 示例4：流水线编排
func examplePipeline() {
	fmt.Println("--- 示例4：流水线编排 ---")

	pipeline := core.NewPipeline("data-analysis-pipeline")

	pipeline.AddStep("validate", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		rawData, ok := input["raw_data"].(string)
		if !ok || rawData == "" {
			return nil, fmt.Errorf("原始数据不能为空")
		}
		input["validated"] = true
		input["data_length"] = len(rawData)
		fmt.Printf("  [验证] 数据长度: %d\n", len(rawData))
		return input, nil
	})

	pipeline.AddStep("transform", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		input["transformed"] = fmt.Sprintf("TRANSFORMED:%s", input["raw_data"])
		fmt.Printf("  [转换] 数据已转换\n")
		return input, nil
	}).WithTimeout(5 * time.Second)

	pipeline.AddStep("analyze", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		length, _ := input["data_length"].(int)
		input["analysis"] = map[string]interface{}{
			"char_count": length,
			"status":     "completed",
		}
		fmt.Printf("  [分析] 字符数: %d\n", length)
		return input, nil
	})

	pipeline.AddStep("store", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		input["stored"] = true
		input["storage_id"] = "doc-001"
		fmt.Printf("  [存储] 已保存 (ID: doc-001)\n")
		return input, nil
	})

	ctx := context.Background()
	result, err := pipeline.Execute(ctx, map[string]interface{}{
		"raw_data": "Go Spring AI MCP Pipeline Demo",
	})
	if err != nil {
		log.Fatalf("流水线执行失败: %v", err)
	}

	fmt.Printf("\n流水线执行结果:\n")
	for k, v := range result {
		fmt.Printf("  %s: %v\n", k, v)
	}
}

// 示例5：中间件机制
func exampleMiddleware() {
	fmt.Println("--- 示例5：中间件机制 ---")

	mcp := core.NewMCPController(2)

	var logEntries []string

	mcp.Use(func(ctx context.Context, task *core.Task) error {
		entry := fmt.Sprintf("[日志] 任务提交: %s (%s)", task.Name, task.Type)
		logEntries = append(logEntries, entry)
		return nil
	})

	mcp.Use(func(ctx context.Context, task *core.Task) error {
		if task.Priority < 5 {
			task.Priority = 5
		}
		return nil
	})

	mcp.Use(func(ctx context.Context, task *core.Task) error {
		if task.Metadata == nil {
			task.Metadata = make(map[string]interface{})
		}
		task.Metadata["submitted_at"] = time.Now().Format(time.RFC3339)
		return nil
	})

	err := mcp.TaskManager().RegisterHandler("middleware-demo", func(ctx context.Context, task *core.Task) error {
		task.Output = map[string]interface{}{"processed": true}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	mcp.Start()
	defer mcp.Stop()

	lowPriorityTask := &core.Task{
		ID:       "low-priority-task",
		Name:     "低优先级任务",
		Type:     "middleware-demo",
		Priority: 1,
	}

	err = mcp.SubmitTask(lowPriorityTask)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(200 * time.Millisecond)

	completedTask, ok := mcp.TaskManager().GetTask("low-priority-task")
	if ok {
		fmt.Printf("原始优先级: 1, 调整后优先级: %d\n", completedTask.Priority)
		fmt.Printf("提交时间: %v\n", completedTask.Metadata["submitted_at"])
	}

	fmt.Printf("\n中间件日志:\n")
	for _, entry := range logEntries {
		fmt.Printf("  %s\n", entry)
	}
}

// 示例6：完整系统演示
func exampleFullSystem() {
	fmt.Println("--- 示例6：完整系统演示 ---")

	mcp := core.NewMCPController(4)

	mcp.ResourceManager().Register("compute", 4, map[string]interface{}{"type": "cpu-core"})
	mcp.ResourceManager().Register("storage", 10240, map[string]interface{}{"type": "mb"})

	processingPipeline := core.NewPipeline("document-processing")
	processingPipeline.AddStep("parse", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		content, _ := input["content"].(string)
		input["parsed"] = content
		input["word_count"] = len(content)
		return input, nil
	})
	processingPipeline.AddStep("embed", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		count, _ := input["word_count"].(int)
		input["embedding"] = make([]float32, count%128+64)
		input["vector_dim"] = len(input["embedding"].([]float32))
		return input, nil
	})
	processingPipeline.AddStep("index", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		dim, _ := input["vector_dim"].(int)
		input["indexed"] = true
		input["index_id"] = fmt.Sprintf("idx-%d", dim)
		return input, nil
	})
	mcp.RegisterPipeline(processingPipeline)

	err := mcp.TaskManager().RegisterHandler("process-document", func(ctx context.Context, task *core.Task) error {
		content, ok := task.Input["content"].(string)
		if !ok {
			return fmt.Errorf("缺少文档内容")
		}

		consumerID := task.ID

		if err := mcp.ResourceManager().Allocate("compute", consumerID, 1); err != nil {
			return err
		}
		defer mcp.ResourceManager().Release("compute", consumerID, 1)

		if err := mcp.ResourceManager().Allocate("storage", consumerID, 100); err != nil {
			return err
		}
		defer mcp.ResourceManager().Release("storage", consumerID, 100)

		ctx = context.Background()
		result, err := mcp.ExecutePipeline(ctx, "document-processing", map[string]interface{}{
			"content": content,
		})
		if err != nil {
			return err
		}

		task.Output = result
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	var authMiddleware = func(ctx context.Context, task *core.Task) error {
		token, ok := task.Metadata["auth_token"].(string)
		if !ok || token == "" {
			return fmt.Errorf("未授权：缺少认证令牌")
		}
		return nil
	}

	var metricsMiddleware = func(ctx context.Context, task *core.Task) error {
		if task.Metadata == nil {
			task.Metadata = make(map[string]interface{})
		}
		task.Metadata["metrics_start_time"] = time.Now().UnixNano()
		return nil
	}

	mcp.Use(authMiddleware)
	mcp.Use(metricsMiddleware)

	mcp.Start()
	defer mcp.Stop()

	documents := []string{
		"Go语言是一门开源编程语言，专为构建简单、可靠和高效的软件而设计。",
		"Spring AI Alibaba 是一个 AI 应用开发框架，提供统一的 API 接口。",
		"MCP（Model Context Protocol）是一种用于构建 AI 应用的协议标准。",
	}

	for i, doc := range documents {
		task := &core.Task{
			ID:       fmt.Sprintf("doc-%03d", i+1),
			Name:     fmt.Sprintf("文档处理 #%d", i+1),
			Type:     "process-document",
			Priority: 3,
			Input:    map[string]interface{}{"content": doc},
			Metadata: map[string]interface{}{"auth_token": "valid-token-123"},
		}

		err = mcp.SubmitTask(task)
		if err != nil {
			fmt.Printf("提交失败: %v\n", err)
			continue
		}
		fmt.Printf("已提交文档任务: %s\n", task.ID)
	}

	time.Sleep(500 * time.Millisecond)

	status := mcp.Status()
	fmt.Printf("\n=== 系统状态报告 ===\n")
	fmt.Printf("运行中: %v\n", status.Running)
	fmt.Printf("任务统计: 总计=%d, 完成=%d, 运行中=%d, 失败=%d\n",
		status.TaskStats.Total, status.TaskStats.Completed,
		status.TaskStats.Running, status.TaskStats.Failed)
	fmt.Printf("资源状态:\n")
	for _, r := range status.Resources {
		fmt.Printf("  %-10s 使用=%-4d/%-4d\n", r.Name, r.Used, r.Capacity)
	}
	fmt.Printf("注册流水线: %d 个\n", status.PipelineCount)

	fmt.Printf("\n各任务结果:\n")
	for i := range documents {
		taskID := fmt.Sprintf("doc-%03d", i+1)
		task, ok := mcp.TaskManager().GetTask(taskID)
		if ok && task.State == core.TaskCompleted {
			output, _ := task.Output["index_id"].(string)
			wordCount, _ := task.Output["word_count"].(int)
			fmt.Printf("  %s: 状态=%s, 词数=%d, 索引ID=%s\n",
				taskID, task.State, wordCount, output)
		} else if ok {
			fmt.Printf("  %s: 状态=%s, 错误=%v\n", taskID, task.State, task.Error)
		}
	}
}
