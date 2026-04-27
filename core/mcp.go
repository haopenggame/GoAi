package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TaskState 任务状态
type TaskState string

const (
	TaskPending   TaskState = "pending"
	TaskRunning   TaskState = "running"
	TaskCompleted TaskState = "completed"
	TaskFailed    TaskState = "failed"
	TaskCancelled TaskState = "cancelled"
)

// Task 表示一个可调度的任务
type Task struct {
	ID          string
	Name        string
	Type        string
	Priority    int
	State       TaskState
	Input       map[string]interface{}
	Output      map[string]interface{}
	Error       error
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	Metadata    map[string]interface{}
	CancelFunc  context.CancelFunc
}

// TaskHandler 任务处理函数
type TaskHandler func(ctx context.Context, task *Task) error

// TaskManager 任务管理器，负责任务的调度、执行和生命周期管理
type TaskManager struct {
	mu       sync.RWMutex
	handlers map[string]TaskHandler
	tasks    map[string]*Task
	queue    chan *Task
	maxWorkers int
	workers    int
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewTaskManager 创建新的任务管理器
func NewTaskManager(maxWorkers int) *TaskManager {
	if maxWorkers <= 0 {
		maxWorkers = 10
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskManager{
		handlers:   make(map[string]TaskHandler),
		tasks:      make(map[string]*Task),
		queue:      make(chan *Task, 1000),
		maxWorkers: maxWorkers,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// RegisterHandler 注册任务处理器
func (tm *TaskManager) RegisterHandler(taskType string, handler TaskHandler) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if _, exists := tm.handlers[taskType]; exists {
		return fmt.Errorf("任务类型 %s 的处理器已存在", taskType)
	}
	tm.handlers[taskType] = handler
	return nil
}

// Submit 提交任务
func (tm *TaskManager) Submit(task *Task) error {
	tm.mu.Lock()
	if _, exists := tm.handlers[task.Type]; !exists {
		tm.mu.Unlock()
		return fmt.Errorf("未找到任务类型 %s 的处理器", task.Type)
	}
	task.State = TaskPending
	task.CreatedAt = time.Now()
	tm.tasks[task.ID] = task
	tm.mu.Unlock()

	tm.queue <- task
	return nil
}

// Start 启动任务管理器的工作协程
func (tm *TaskManager) Start() {
	for i := 0; i < tm.maxWorkers; i++ {
		tm.wg.Add(1)
		go tm.worker()
	}
}

// Stop 停止任务管理器
func (tm *TaskManager) Stop() {
	tm.cancel()
	close(tm.queue)
	tm.wg.Wait()
}

func (tm *TaskManager) worker() {
	defer tm.wg.Done()
	for {
		select {
		case <-tm.ctx.Done():
			return
		case task, ok := <-tm.queue:
			if !ok {
				return
			}
			tm.executeTask(task)
		}
	}
}

func (tm *TaskManager) executeTask(task *Task) {
	tm.mu.RLock()
	handler, exists := tm.handlers[task.Type]
	tm.mu.RUnlock()

	if !exists {
		task.State = TaskFailed
		task.Error = fmt.Errorf("未找到任务类型 %s 的处理器", task.Type)
		return
	}

	ctx, cancel := context.WithCancel(tm.ctx)
	task.CancelFunc = cancel
	task.State = TaskRunning
	task.StartedAt = time.Now()

	err := handler(ctx, task)

	task.CompletedAt = time.Now()
	if err != nil {
		task.State = TaskFailed
		task.Error = err
	} else {
		task.State = TaskCompleted
	}
}

// GetTask 获取任务信息
func (tm *TaskManager) GetTask(id string) (*Task, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	task, ok := tm.tasks[id]
	return task, ok
}

// CancelTask 取消任务
func (tm *TaskManager) CancelTask(id string) error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	task, ok := tm.tasks[id]
	if !ok {
		return fmt.Errorf("任务 %s 不存在", id)
	}
	if task.CancelFunc != nil {
		task.CancelFunc()
	}
	task.State = TaskCancelled
	return nil
}

// TaskStats 任务统计信息
type TaskStats struct {
	Total     int
	Pending   int
	Running   int
	Completed int
	Failed    int
	Cancelled int
}

// Stats 获取任务统计信息
func (tm *TaskManager) Stats() TaskStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var stats TaskStats
	for _, task := range tm.tasks {
		stats.Total++
		switch task.State {
		case TaskPending:
			stats.Pending++
		case TaskRunning:
			stats.Running++
		case TaskCompleted:
			stats.Completed++
		case TaskFailed:
			stats.Failed++
		case TaskCancelled:
			stats.Cancelled++
		}
	}
	return stats
}

// ResourceState 资源状态
type ResourceState struct {
	Name      string
	Capacity  int
	Used      int
	Available int
	Metadata  map[string]interface{}
}

// ResourceManager 资源管理器，负责系统资源的分配、监控和回收
type ResourceManager struct {
	mu        sync.RWMutex
	resources map[string]*ResourceState
	allocations map[string]map[string]int
}

// NewResourceManager 创建新的资源管理器
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources:   make(map[string]*ResourceState),
		allocations: make(map[string]map[string]int),
	}
}

// Register 注册资源
func (rm *ResourceManager) Register(name string, capacity int, metadata map[string]interface{}) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if _, exists := rm.resources[name]; exists {
		return fmt.Errorf("资源 %s 已注册", name)
	}
	rm.resources[name] = &ResourceState{
		Name:      name,
		Capacity:  capacity,
		Available: capacity,
		Metadata:  metadata,
	}
	rm.allocations[name] = make(map[string]int)
	return nil
}

// Allocate 分配资源
func (rm *ResourceManager) Allocate(resourceName string, consumerID string, amount int) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	resource, exists := rm.resources[resourceName]
	if !exists {
		return fmt.Errorf("资源 %s 不存在", resourceName)
	}

	if resource.Available < amount {
		return fmt.Errorf("资源 %s 不足：需要 %d，可用 %d", resourceName, amount, resource.Available)
	}

	resource.Available -= amount
	resource.Used += amount
	rm.allocations[resourceName][consumerID] += amount
	return nil
}

// Release 释放资源
func (rm *ResourceManager) Release(resourceName string, consumerID string, amount int) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	resource, exists := rm.resources[resourceName]
	if !exists {
		return fmt.Errorf("资源 %s 不存在", resourceName)
	}

	allocated := rm.allocations[resourceName][consumerID]
	if allocated < amount {
		amount = allocated
	}

	resource.Available += amount
	resource.Used -= amount
	rm.allocations[resourceName][consumerID] -= amount
	if rm.allocations[resourceName][consumerID] <= 0 {
		delete(rm.allocations[resourceName], consumerID)
	}
	return nil
}

// GetState 获取资源状态
func (rm *ResourceManager) GetState(resourceName string) (*ResourceState, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	state, ok := rm.resources[resourceName]
	if !ok {
		return nil, false
	}
	copy := *state
	return &copy, true
}

// ListResources 列出所有资源
func (rm *ResourceManager) ListResources() []ResourceState {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make([]ResourceState, 0, len(rm.resources))
	for _, state := range rm.resources {
		result = append(result, *state)
	}
	return result
}

// PipelineStep 流水线步骤
type PipelineStep struct {
	Name     string
	Handler  func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
	OnError  func(err error) map[string]interface{}
	Timeout  time.Duration
}

// Pipeline 流水线，用于编排多个处理步骤
type Pipeline struct {
	Name  string
	Steps []PipelineStep
}

// NewPipeline 创建新的流水线
func NewPipeline(name string) *Pipeline {
	return &Pipeline{
		Name:  name,
		Steps: make([]PipelineStep, 0),
	}
}

// AddStep 添加流水线步骤
func (p *Pipeline) AddStep(name string, handler func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)) *Pipeline {
	p.Steps = append(p.Steps, PipelineStep{
		Name:    name,
		Handler: handler,
	})
	return p
}

// WithTimeout 为步骤设置超时
func (p *Pipeline) WithTimeout(timeout time.Duration) *Pipeline {
	if len(p.Steps) > 0 {
		p.Steps[len(p.Steps)-1].Timeout = timeout
	}
	return p
}

// WithErrorHandler 为步骤设置错误处理器
func (p *Pipeline) WithErrorHandler(handler func(err error) map[string]interface{}) *Pipeline {
	if len(p.Steps) > 0 {
		p.Steps[len(p.Steps)-1].OnError = handler
	}
	return p
}

// Execute 执行流水线
func (p *Pipeline) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	data := input
	for _, step := range p.Steps {
		stepCtx := ctx
		if step.Timeout > 0 {
			var cancel context.CancelFunc
			stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
			defer cancel()
		}

		result, err := step.Handler(stepCtx, data)
		if err != nil {
			if step.OnError != nil {
				data = step.OnError(err)
				continue
			}
			return nil, fmt.Errorf("流水线步骤 %s 执行失败: %w", step.Name, err)
		}
		data = result
	}
	return data, nil
}

// MCPController 主控程序控制器，统一管理任务、资源和流水线
type MCPController struct {
	taskManager     *TaskManager
	resourceManager *ResourceManager
	pipelines       map[string]*Pipeline
	middleware      []func(ctx context.Context, task *Task) error
	mu              sync.RWMutex
	running         bool
}

// NewMCPController 创建新的主控程序控制器
func NewMCPController(maxWorkers int) *MCPController {
	return &MCPController{
		taskManager:     NewTaskManager(maxWorkers),
		resourceManager: NewResourceManager(),
		pipelines:       make(map[string]*Pipeline),
		middleware:      make([]func(ctx context.Context, task *Task) error, 0),
	}
}

// TaskManager 获取任务管理器
func (c *MCPController) TaskManager() *TaskManager {
	return c.taskManager
}

// ResourceManager 获取资源管理器
func (c *MCPController) ResourceManager() *ResourceManager {
	return c.resourceManager
}

// RegisterPipeline 注册流水线
func (c *MCPController) RegisterPipeline(pipeline *Pipeline) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.pipelines[pipeline.Name]; exists {
		return fmt.Errorf("流水线 %s 已存在", pipeline.Name)
	}
	c.pipelines[pipeline.Name] = pipeline
	return nil
}

// GetPipeline 获取流水线
func (c *MCPController) GetPipeline(name string) (*Pipeline, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	pipeline, ok := c.pipelines[name]
	return pipeline, ok
}

// Use 添加中间件
func (c *MCPController) Use(middleware func(ctx context.Context, task *Task) error) {
	c.middleware = append(c.middleware, middleware)
}

// Start 启动主控程序
func (c *MCPController) Start() {
	c.mu.Lock()
	c.running = true
	c.mu.Unlock()
	c.taskManager.Start()
}

// Stop 停止主控程序
func (c *MCPController) Stop() {
	c.mu.Lock()
	c.running = false
	c.mu.Unlock()
	c.taskManager.Stop()
}

// IsRunning 检查主控程序是否运行中
func (c *MCPController) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// SubmitTask 提交任务（经过中间件处理）
func (c *MCPController) SubmitTask(task *Task) error {
	ctx := context.Background()
	for _, mw := range c.middleware {
		if err := mw(ctx, task); err != nil {
			return fmt.Errorf("中间件处理失败: %w", err)
		}
	}
	return c.taskManager.Submit(task)
}

// ExecutePipeline 执行指定名称的流水线
func (c *MCPController) ExecutePipeline(ctx context.Context, name string, input map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	pipeline, ok := c.pipelines[name]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("流水线 %s 不存在", name)
	}
	return pipeline.Execute(ctx, input)
}

// SystemStatus 系统状态
type SystemStatus struct {
	Running      bool
	TaskStats    TaskStats
	Resources    []ResourceState
	PipelineCount int
}

// Status 获取系统状态
func (c *MCPController) Status() SystemStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return SystemStatus{
		Running:       c.running,
		TaskStats:     c.taskManager.Stats(),
		Resources:     c.resourceManager.ListResources(),
		PipelineCount: len(c.pipelines),
	}
}
