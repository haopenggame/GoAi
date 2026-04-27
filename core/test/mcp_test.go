package test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

func TestMCPController(t *testing.T) {
	t.Run("创建控制器", func(t *testing.T) {
		mcp := core.NewMCPController(2)
		assert.NotNil(t, mcp)
	})

	t.Run("启动和停止控制器", func(t *testing.T) {
		mcp := core.NewMCPController(1)
		mcp.Start()
		defer mcp.Stop()
	})

	t.Run("提交任务", func(t *testing.T) {
		mcp := core.NewMCPController(2)

		err := mcp.TaskManager().RegisterHandler("test", func(ctx context.Context, task *core.Task) error {
			// 模拟任务执行
			time.Sleep(50 * time.Millisecond)
			return nil
		})
		assert.NoError(t, err)

		mcp.Start()
		defer mcp.Stop()

		task := &core.Task{
			ID:   "task-1",
			Name: "测试任务",
			Type: "test",
		}

		err = mcp.SubmitTask(task)
		assert.NoError(t, err)

		// 等待任务执行完成
		time.Sleep(100 * time.Millisecond)

		// 检查任务状态
		submittedTask, ok := mcp.TaskManager().GetTask("task-1")
		assert.True(t, ok)
		assert.Equal(t, core.TaskCompleted, submittedTask.State)
	})

	t.Run("并发任务", func(t *testing.T) {
		mcp := core.NewMCPController(2)

		var wg sync.WaitGroup

		err := mcp.TaskManager().RegisterHandler("parallel", func(ctx context.Context, task *core.Task) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		assert.NoError(t, err)

		mcp.Start()
		defer mcp.Stop()

		// 提交多个任务
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				task := &core.Task{
					ID:   fmt.Sprintf("task-%d", i),
					Name: fmt.Sprintf("任务%d", i),
					Type: "parallel",
				}
				_ = mcp.SubmitTask(task)
			}(i)
		}

		wg.Wait()
		// 等待所有任务完成
		time.Sleep(300 * time.Millisecond)

		// 检查任务状态
		for i := 0; i < 5; i++ {
			task, ok := mcp.TaskManager().GetTask(fmt.Sprintf("task-%d", i))
			assert.True(t, ok)
			assert.Equal(t, core.TaskCompleted, task.State)
		}
	})
}

func TestTaskManager(t *testing.T) {
	t.Run("创建任务管理器", func(t *testing.T) {
		manager := core.NewTaskManager(2)
		assert.NotNil(t, manager)
	})

	t.Run("注册和提交任务", func(t *testing.T) {
		manager := core.NewTaskManager(1)

		var executed bool
		err := manager.RegisterHandler("test", func(ctx context.Context, task *core.Task) error {
			executed = true
			return nil
		})
		assert.NoError(t, err)

		task := &core.Task{
			ID:   "test-task",
			Name: "测试任务",
			Type: "test",
		}

		err = manager.Submit(task)
		assert.NoError(t, err)

		// 启动任务管理器
		manager.Start()
		defer manager.Stop()

		// 等待任务执行完成
		time.Sleep(100 * time.Millisecond)

		assert.True(t, executed)

		// 检查任务状态
		submittedTask, ok := manager.GetTask("test-task")
		assert.True(t, ok)
		assert.Equal(t, core.TaskCompleted, submittedTask.State)
	})

	t.Run("任务状态管理", func(t *testing.T) {
		manager := core.NewTaskManager(1)

		task := &core.Task{
			ID:   "status-task",
			Name: "状态测试任务",
			Type: "test",
		}

		// 注册任务处理器
		err := manager.RegisterHandler("test", func(ctx context.Context, task *core.Task) error {
			return nil
		})
		assert.NoError(t, err)

		// 提交任务
		err = manager.Submit(task)
		assert.NoError(t, err)

		// 启动任务管理器
		manager.Start()
		defer manager.Stop()

		// 等待任务执行完成
		time.Sleep(100 * time.Millisecond)

		// 检查任务状态
		submittedTask, ok := manager.GetTask("status-task")
		assert.True(t, ok)
		assert.Equal(t, core.TaskCompleted, submittedTask.State)
	})

	t.Run("取消任务", func(t *testing.T) {
		manager := core.NewTaskManager(1)

		// 注册一个长时间运行的任务
		err := manager.RegisterHandler("long-running", func(ctx context.Context, task *core.Task) error {
			time.Sleep(500 * time.Millisecond)
			return nil
		})
		assert.NoError(t, err)

		task := &core.Task{
			ID:   "cancel-task",
			Name: "取消测试任务",
			Type: "long-running",
		}

		err = manager.Submit(task)
		assert.NoError(t, err)

		// 启动任务管理器
		manager.Start()
		defer manager.Stop()

		// 等待一段时间后取消任务
		time.Sleep(100 * time.Millisecond)
		err = manager.CancelTask("cancel-task")
		assert.NoError(t, err)

		// 等待任务被取消
		time.Sleep(100 * time.Millisecond)

		// 检查任务状态
		cancelledTask, ok := manager.GetTask("cancel-task")
		assert.True(t, ok)
		assert.Equal(t, core.TaskCancelled, cancelledTask.State)
	})

	t.Run("任务统计", func(t *testing.T) {
		manager := core.NewTaskManager(1)

		// 注册任务处理器
		err := manager.RegisterHandler("test", func(ctx context.Context, task *core.Task) error {
			return nil
		})
		assert.NoError(t, err)

		// 提交任务
		task1 := &core.Task{
			ID:   "stat-task-1",
			Name: "统计测试任务1",
			Type: "test",
		}
		task2 := &core.Task{
			ID:   "stat-task-2",
			Name: "统计测试任务2",
			Type: "test",
		}

		err = manager.Submit(task1)
		assert.NoError(t, err)
		err = manager.Submit(task2)
		assert.NoError(t, err)

		// 启动任务管理器
		manager.Start()
		defer manager.Stop()

		// 等待任务执行完成
		time.Sleep(100 * time.Millisecond)

		// 检查任务统计
		stats := manager.Stats()
		assert.Equal(t, 2, stats.Total)
		assert.Equal(t, 2, stats.Completed)
	})
}

func TestResourceManager(t *testing.T) {
	t.Run("创建资源管理器", func(t *testing.T) {
		manager := core.NewResourceManager()
		assert.NotNil(t, manager)
	})

	t.Run("注册和分配资源", func(t *testing.T) {
		manager := core.NewResourceManager()

		// 注册资源
		err := manager.Register("cpu", 4, map[string]interface{}{"type": "intel"})
		assert.NoError(t, err)

		// 分配资源
		err = manager.Allocate("cpu", "process-1", 2)
		assert.NoError(t, err)

		// 检查资源状态
		state, ok := manager.GetState("cpu")
		assert.True(t, ok)
		assert.Equal(t, 4, state.Capacity)
		assert.Equal(t, 2, state.Used)
		assert.Equal(t, 2, state.Available)
	})

	t.Run("释放资源", func(t *testing.T) {
		manager := core.NewResourceManager()

		// 注册资源
		err := manager.Register("memory", 1024, map[string]interface{}{"type": "ddr4"})
		assert.NoError(t, err)

		// 分配资源
		err = manager.Allocate("memory", "process-1", 512)
		assert.NoError(t, err)

		// 释放资源
		err = manager.Release("memory", "process-1", 256)
		assert.NoError(t, err)

		// 检查资源状态
		state, ok := manager.GetState("memory")
		assert.True(t, ok)
		assert.Equal(t, 1024, state.Capacity)
		assert.Equal(t, 256, state.Used)
		assert.Equal(t, 768, state.Available)
	})

	t.Run("列出资源", func(t *testing.T) {
		manager := core.NewResourceManager()

		// 注册资源
		err := manager.Register("cpu", 4, nil)
		assert.NoError(t, err)
		err = manager.Register("memory", 1024, nil)
		assert.NoError(t, err)

		// 列出资源
		resources := manager.ListResources()
		assert.Len(t, resources, 2)
		assert.Equal(t, "cpu", resources[0].Name)
		assert.Equal(t, "memory", resources[1].Name)
	})
}

func TestPipeline(t *testing.T) {
	t.Run("创建和执行流水线", func(t *testing.T) {
		pipeline := core.NewPipeline("test-pipeline")

		// 添加步骤
		pipeline.AddStep("step1", func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
			data["step1"] = "completed"
			return data, nil
		})

		pipeline.AddStep("step2", func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
			data["step2"] = "completed"
			return data, nil
		})

		// 执行流水线
		ctx := context.Background()
		initialData := map[string]interface{}{"initial": "data"}
		result, err := pipeline.Execute(ctx, initialData)
		assert.NoError(t, err)
		assert.Equal(t, "completed", result["step1"])
		assert.Equal(t, "completed", result["step2"])
	})

	t.Run("流水线错误处理", func(t *testing.T) {
		pipeline := core.NewPipeline("error-pipeline")

		pipeline.AddStep("step1", func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
			return data, nil
		})

		pipeline.AddStep("step2", func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
			return nil, assert.AnError
		})

		// 执行流水线
		ctx := context.Background()
		initialData := map[string]interface{}{"initial": "data"}
		result, err := pipeline.Execute(ctx, initialData)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("带超时的流水线", func(t *testing.T) {
		pipeline := core.NewPipeline("timeout-pipeline")

		// 添加带超时的步骤
		pipeline.AddStep("step1", func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
			time.Sleep(100 * time.Millisecond)
			data["step1"] = "completed"
			return data, nil
		}).WithTimeout(50 * time.Millisecond)

		// 执行流水线
		ctx := context.Background()
		initialData := map[string]interface{}{"initial": "data"}
		result, err := pipeline.Execute(ctx, initialData)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestMCPMiddleware(t *testing.T) {
	t.Run("添加和执行中间件", func(t *testing.T) {
		mcp := core.NewMCPController(1)

		var middlewareCalled bool

		// 添加中间件
		mcp.Use(func(ctx context.Context, task *core.Task) error {
			middlewareCalled = true
			task.Metadata["middleware"] = "called"
			return nil
		})

		// 注册任务处理器
		err := mcp.TaskManager().RegisterHandler("middleware-test", func(ctx context.Context, task *core.Task) error {
			return nil
		})
		assert.NoError(t, err)

		mcp.Start()
		defer mcp.Stop()

		// 提交任务
		task := &core.Task{
			ID:   "middleware-task",
			Name: "中间件测试任务",
			Type: "middleware-test",
		}

		err = mcp.SubmitTask(task)
		assert.NoError(t, err)

		// 等待任务执行完成
		time.Sleep(100 * time.Millisecond)

		// 检查中间件是否被调用
		assert.True(t, middlewareCalled)

		// 检查任务是否包含中间件添加的元数据
		submittedTask, ok := mcp.TaskManager().GetTask("middleware-task")
		assert.True(t, ok)
		assert.Equal(t, "called", submittedTask.Metadata["middleware"])
	})
}

func TestMCPSystemStatus(t *testing.T) {
	t.Run("获取系统状态", func(t *testing.T) {
		mcp := core.NewMCPController(2)

		// 注册资源
		err := mcp.ResourceManager().Register("cpu", 4, nil)
		assert.NoError(t, err)

		// 注册流水线
		pipeline := core.NewPipeline("test-pipeline")
		err = mcp.RegisterPipeline(pipeline)
		assert.NoError(t, err)

		// 启动控制器
		mcp.Start()
		defer mcp.Stop()

		// 获取系统状态
		status := mcp.Status()
		assert.True(t, status.Running)
		assert.Equal(t, 0, status.TaskStats.Total)
		assert.Len(t, status.Resources, 1)
		assert.Equal(t, 1, status.PipelineCount)
	})
}
