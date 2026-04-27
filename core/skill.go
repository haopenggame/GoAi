package core

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// SkillState 技能状态
type SkillState string

const (
	SkillActive   SkillState = "active"
	SkillInactive SkillState = "inactive"
	SkillLocked   SkillState = "locked"
	SkillUpgrade  SkillState = "upgrading"
)

// SkillLevel 技能等级
type SkillLevel int

const (
	LevelNovice     SkillLevel = 1
	LevelApprentice SkillLevel = 2
	LevelAdept      SkillLevel = 3
	LevelExpert     SkillLevel = 4
	LevelMaster     SkillLevel = 5
)

// SkillEffect 技能效果
type SkillEffect struct {
	Name        string
	Value       float64
	Duration    time.Duration
	Stackable   bool
	StackCount  int
	Description string
}

// SkillTrigger 技能触发条件
type SkillTrigger struct {
	Type      string
	Condition string
	Threshold float64
	Cooldown  time.Duration
	LastFired time.Time
}

// SkillDefinition 技能定义
type SkillDefinition struct {
	ID          string
	Name        string
	Category    string
	Level       SkillLevel
	State       SkillState
	Description string
	Effects     []SkillEffect
	Triggers    []SkillTrigger
	Dependencies []string
	Parameters  map[string]interface{}
	Metadata    map[string]interface{}
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SkillContext 技能执行上下文
type SkillContext struct {
	Invoker    string
	Target     string
	Parameters map[string]interface{}
	Effects    []SkillEffect
	Cancel     context.CancelFunc
}

// SkillExecutor 技能执行函数
type SkillExecutor func(ctx context.Context, skillCtx *SkillContext) (*SkillResult, error)

// SkillResult 技能执行结果
type SkillResult struct {
	Success   bool
	Output    map[string]interface{}
	Effects   []SkillEffect
	Duration  time.Duration
	Error     error
	Metadata  map[string]interface{}
}

// SkillUpgradeRule 技能升级规则
type SkillUpgradeRule struct {
	FromLevel   SkillLevel
	ToLevel     SkillLevel
	RequiredExp int
	Cost        map[string]interface{}
	Conditions  []string
}

// SkillRegistry 技能注册中心，管理技能的注册、查询和生命周期
type SkillRegistry struct {
	mu         sync.RWMutex
	skills     map[string]*SkillDefinition
	executors  map[string]SkillExecutor
	upgradeRules map[string][]SkillUpgradeRule
	experience map[string]int
	hooks      map[string][]SkillHook
}

// SkillHook 技能钩子函数
type SkillHook func(ctx context.Context, skillDef *SkillDefinition, result *SkillResult) error

// NewSkillRegistry 创建新的技能注册中心
func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills:       make(map[string]*SkillDefinition),
		executors:    make(map[string]SkillExecutor),
		upgradeRules: make(map[string][]SkillUpgradeRule),
		experience:   make(map[string]int),
		hooks:        make(map[string][]SkillHook),
	}
}

// Register 注册技能
func (r *SkillRegistry) Register(def SkillDefinition, executor SkillExecutor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.skills[def.ID]; exists {
		return fmt.Errorf("技能 %s 已注册", def.ID)
	}

	if def.CreatedAt.IsZero() {
		def.CreatedAt = time.Now()
	}
	def.UpdatedAt = time.Now()
	if def.State == "" {
		def.State = SkillActive
	}
	if def.Parameters == nil {
		def.Parameters = make(map[string]interface{})
	}
	if def.Metadata == nil {
		def.Metadata = make(map[string]interface{})
	}

	r.skills[def.ID] = &def
	r.executors[def.ID] = executor
	r.experience[def.ID] = 0
	return nil
}

// Unregister 注销技能
func (r *SkillRegistry) Unregister(skillID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.skills[skillID]; !exists {
		return fmt.Errorf("技能 %s 不存在", skillID)
	}

	delete(r.skills, skillID)
	delete(r.executors, skillID)
	delete(r.experience, skillID)
	delete(r.upgradeRules, skillID)
	return nil
}

// Get 获取技能定义
func (r *SkillRegistry) Get(skillID string) (*SkillDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	def, ok := r.skills[skillID]
	if !ok {
		return nil, false
	}
	copy := *def
	return &copy, true
}

// List 列出所有技能
func (r *SkillRegistry) List() []SkillDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]SkillDefinition, 0, len(r.skills))
	for _, def := range r.skills {
		result = append(result, *def)
	}
	return result
}

// ListByCategory 按类别列出技能
func (r *SkillRegistry) ListByCategory(category string) []SkillDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []SkillDefinition
	for _, def := range r.skills {
		if def.Category == category {
			result = append(result, *def)
		}
	}
	return result
}

// ListByLevel 按等级列出技能
func (r *SkillRegistry) ListByLevel(level SkillLevel) []SkillDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []SkillDefinition
	for _, def := range r.skills {
		if def.Level == level {
			result = append(result, *def)
		}
	}
	return result
}

// Execute 执行技能
func (r *SkillRegistry) Execute(ctx context.Context, skillID string, skillCtx *SkillContext) (*SkillResult, error) {
	r.mu.RLock()
	def, defExists := r.skills[skillID]
	executor, execExists := r.executors[skillID]
	r.mu.RUnlock()

	if !defExists {
		return nil, fmt.Errorf("技能 %s 不存在", skillID)
	}
	if !execExists {
		return nil, fmt.Errorf("技能 %s 没有注册执行器", skillID)
	}
	if def.State != SkillActive {
		return nil, fmt.Errorf("技能 %s 当前状态为 %s，无法执行", skillID, def.State)
	}

	for _, depID := range def.Dependencies {
		r.mu.RLock()
		dep, depExists := r.skills[depID]
		r.mu.RUnlock()
		if !depExists || dep.State != SkillActive {
			return nil, fmt.Errorf("技能 %s 的依赖 %s 未满足", skillID, depID)
		}
	}

	startTime := time.Now()
	result, err := executor(ctx, skillCtx)
	if result != nil {
		result.Duration = time.Since(startTime)
	}

	if err != nil {
		if result == nil {
			result = &SkillResult{Success: false, Error: err}
		}
	}

	r.executeHooks(ctx, skillID, "after", def, result)

	r.mu.Lock()
	r.experience[skillID] += 1
	r.mu.Unlock()

	return result, err
}

// AddHook 添加技能钩子
func (r *SkillRegistry) AddHook(skillID string, hookType string, hook SkillHook) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := skillID + ":" + hookType
	r.hooks[key] = append(r.hooks[key], hook)
	return nil
}

func (r *SkillRegistry) executeHooks(ctx context.Context, skillID string, hookType string, def *SkillDefinition, result *SkillResult) {
	key := skillID + ":" + hookType
	r.mu.RLock()
	hooks, ok := r.hooks[key]
	r.mu.RUnlock()

	if !ok {
		return
	}
	for _, hook := range hooks {
		_ = hook(ctx, def, result)
	}
}

// SetUpgradeRules 设置技能升级规则
func (r *SkillRegistry) SetUpgradeRules(skillID string, rules []SkillUpgradeRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.skills[skillID]; !exists {
		return fmt.Errorf("技能 %s 不存在", skillID)
	}
	r.upgradeRules[skillID] = rules
	return nil
}

// Upgrade 升级技能
func (r *SkillRegistry) Upgrade(skillID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	def, exists := r.skills[skillID]
	if !exists {
		return fmt.Errorf("技能 %s 不存在", skillID)
	}

	rules, hasRules := r.upgradeRules[skillID]
	if !hasRules {
		return fmt.Errorf("技能 %s 没有升级规则", skillID)
	}

	var applicableRule *SkillUpgradeRule
	for i := range rules {
		if rules[i].FromLevel == def.Level {
			applicableRule = &rules[i]
			break
		}
	}

	if applicableRule == nil {
		return fmt.Errorf("技能 %s 当前等级 %d 没有可用的升级规则", skillID, def.Level)
	}

	exp := r.experience[skillID]
	if exp < applicableRule.RequiredExp {
		return fmt.Errorf("技能 %s 经验不足：需要 %d，当前 %d", skillID, applicableRule.RequiredExp, exp)
	}

	def.State = SkillUpgrade
	def.Level = applicableRule.ToLevel
	def.UpdatedAt = time.Now()
	def.State = SkillActive

	return nil
}

// GetExperience 获取技能经验值
func (r *SkillRegistry) GetExperience(skillID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, exists := r.skills[skillID]; !exists {
		return 0, fmt.Errorf("技能 %s 不存在", skillID)
	}
	return r.experience[skillID], nil
}

// SetState 设置技能状态
func (r *SkillRegistry) SetState(skillID string, state SkillState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	def, exists := r.skills[skillID]
	if !exists {
		return fmt.Errorf("技能 %s 不存在", skillID)
	}
	def.State = state
	def.UpdatedAt = time.Now()
	return nil
}

// SkillChain 技能链，用于按顺序执行多个技能
type SkillChain struct {
	registry   *SkillRegistry
	steps      []SkillChainStep
	errorHandler func(err error, step SkillChainStep) error
}

// SkillChainStep 技能链步骤
type SkillChainStep struct {
	SkillID    string
	Parameters map[string]interface{}
	SkipOnError bool
}

// NewSkillChain 创建新的技能链
func NewSkillChain(registry *SkillRegistry) *SkillChain {
	return &SkillChain{
		registry: registry,
		steps:    make([]SkillChainStep, 0),
	}
}

// Add 添加技能链步骤
func (c *SkillChain) Add(skillID string, parameters map[string]interface{}) *SkillChain {
	c.steps = append(c.steps, SkillChainStep{
		SkillID:    skillID,
		Parameters: parameters,
	})
	return c
}

// SkipOnError 设置步骤出错时跳过
func (c *SkillChain) SkipOnError() *SkillChain {
	if len(c.steps) > 0 {
		c.steps[len(c.steps)-1].SkipOnError = true
	}
	return c
}

// OnError 设置错误处理器
func (c *SkillChain) OnError(handler func(err error, step SkillChainStep) error) *SkillChain {
	c.errorHandler = handler
	return c
}

// Execute 执行技能链
func (c *SkillChain) Execute(ctx context.Context, initialInput map[string]interface{}) ([]SkillResult, error) {
	results := make([]SkillResult, 0, len(c.steps))
	data := initialInput
	if data == nil {
		data = make(map[string]interface{})
	}

	for _, step := range c.steps {
		skillCtx := &SkillContext{
			Parameters: make(map[string]interface{}),
		}
		for k, v := range step.Parameters {
			skillCtx.Parameters[k] = v
		}
		for k, v := range data {
			skillCtx.Parameters[k] = v
		}

		result, err := c.registry.Execute(ctx, step.SkillID, skillCtx)
		if err != nil {
			if step.SkipOnError {
				continue
			}
			if c.errorHandler != nil {
				if handlerErr := c.errorHandler(err, step); handlerErr != nil {
					return results, handlerErr
				}
				continue
			}
			return results, fmt.Errorf("技能链步骤 %s 执行失败: %w", step.SkillID, err)
		}

		if result != nil && result.Output != nil {
			for k, v := range result.Output {
				data[k] = v
			}
			results = append(results, *result)
		}
	}

	return results, nil
}

// EffectCalculator 效果计算器，处理技能效果的叠加、衰减和计算
type EffectCalculator struct {
	mu      sync.RWMutex
	effects map[string][]ActiveEffect
}

// ActiveEffect 活跃效果
type ActiveEffect struct {
	SourceSkill string
	Effect      SkillEffect
	AppliedAt   time.Time
	ExpiresAt   time.Time
}

// NewEffectCalculator 创建新的效果计算器
func NewEffectCalculator() *EffectCalculator {
	return &EffectCalculator{
		effects: make(map[string][]ActiveEffect),
	}
}

// Apply 应用效果
func (c *EffectCalculator) Apply(target string, sourceSkill string, effect SkillEffect) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	activeEffect := ActiveEffect{
		SourceSkill: sourceSkill,
		Effect:      effect,
		AppliedAt:   now,
	}

	if effect.Duration > 0 {
		activeEffect.ExpiresAt = now.Add(effect.Duration)
	} else {
		activeEffect.ExpiresAt = now.Add(24 * time.Hour)
	}

	existing := c.effects[target]
	for i, ae := range existing {
		if ae.SourceSkill == sourceSkill && ae.Effect.Name == effect.Name {
			if effect.Stackable {
				existing[i].Effect.StackCount++
				existing[i].Effect.Value += effect.Value
				c.effects[target] = existing
				return
			}
			existing[i] = activeEffect
			c.effects[target] = existing
			return
		}
	}

	c.effects[target] = append(c.effects[target], activeEffect)
}

// Calculate 计算目标的总效果
func (c *EffectCalculator) Calculate(target string) map[string]float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	result := make(map[string]float64)

	activeEffects := c.effects[target]
	var filtered []ActiveEffect

	for _, ae := range activeEffects {
		if now.After(ae.ExpiresAt) {
			continue
		}
		filtered = append(filtered, ae)
		result[ae.Effect.Name] += ae.Effect.Value
	}

	c.effects[target] = filtered
	return result
}

// Remove 移除效果
func (c *EffectCalculator) Remove(target string, effectName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var filtered []ActiveEffect
	for _, ae := range c.effects[target] {
		if ae.Effect.Name != effectName {
			filtered = append(filtered, ae)
		}
	}
	c.effects[target] = filtered
}

// Clear 清除目标所有效果
func (c *EffectCalculator) Clear(target string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.effects, target)
}

// SkillTriggerEngine 技能触发引擎
type SkillTriggerEngine struct {
	mu       sync.RWMutex
	registry *SkillRegistry
	triggers map[string][]TriggerBinding
}

// TriggerBinding 触发器绑定
type TriggerBinding struct {
	SkillID string
	Trigger SkillTrigger
}

// NewSkillTriggerEngine 创建新的技能触发引擎
func NewSkillTriggerEngine(registry *SkillRegistry) *SkillTriggerEngine {
	return &SkillTriggerEngine{
		registry: registry,
		triggers: make(map[string][]TriggerBinding),
	}
}

// Bind 绑定触发器
func (e *SkillTriggerEngine) Bind(triggerType string, skillID string, trigger SkillTrigger) {
	e.mu.Lock()
	defer e.mu.Unlock()

	trigger.Type = triggerType
	e.triggers[triggerType] = append(e.triggers[triggerType], TriggerBinding{
		SkillID: skillID,
		Trigger: trigger,
	})
}

// Fire 触发事件
func (e *SkillTriggerEngine) Fire(ctx context.Context, triggerType string, value float64, params map[string]interface{}) []SkillResult {
	e.mu.RLock()
	bindings, ok := e.triggers[triggerType]
	e.mu.RUnlock()

	if !ok {
		return nil
	}

	now := time.Now()
	var results []SkillResult

	sortedBindings := make([]TriggerBinding, len(bindings))
	copy(sortedBindings, bindings)
	sort.Slice(sortedBindings, func(i, j int) bool {
		return sortedBindings[i].Trigger.Threshold > sortedBindings[j].Trigger.Threshold
	})

	for _, binding := range sortedBindings {
		if value < binding.Trigger.Threshold {
			continue
		}

		if !binding.Trigger.LastFired.IsZero() && now.Sub(binding.Trigger.LastFired) < binding.Trigger.Cooldown {
			continue
		}

		skillCtx := &SkillContext{
			Parameters: params,
		}

		result, err := e.registry.Execute(ctx, binding.SkillID, skillCtx)
		if err != nil {
			continue
		}

		if result != nil {
			results = append(results, *result)
		}

		binding.Trigger.LastFired = now
	}

	return results
}
