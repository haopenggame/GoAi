package test

import (
	"context"
	"testing"

	"github.com/go-spring/ai/core"
	"github.com/stretchr/testify/assert"
)

func TestSkillRegistry(t *testing.T) {
	t.Run("创建技能注册中心", func(t *testing.T) {
		registry := core.NewSkillRegistry()
		assert.NotNil(t, registry)
	})

	t.Run("注册技能", func(t *testing.T) {
		registry := core.NewSkillRegistry()

		skillDef := core.SkillDefinition{
			ID:          "test-skill",
			Name:        "测试技能",
			Category:    "测试",
			Level:       core.LevelNovice,
			Description: "这是一个测试技能",
		}

		executor := func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
			return &core.SkillResult{
				Success: true,
				Output:  map[string]interface{}{"result": "成功"},
			}, nil
		}

		err := registry.Register(skillDef, executor)
		assert.NoError(t, err)

		// 检查技能是否存在
		skill, exists := registry.Get("test-skill")
		assert.True(t, exists)
		assert.Equal(t, "测试技能", skill.Name)
	})

	t.Run("注销技能", func(t *testing.T) {
		registry := core.NewSkillRegistry()

		skillDef := core.SkillDefinition{
			ID:   "test-skill",
			Name: "测试技能",
		}

		executor := func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
			return &core.SkillResult{Success: true}, nil
		}

		err := registry.Register(skillDef, executor)
		assert.NoError(t, err)

		err = registry.Unregister("test-skill")
		assert.NoError(t, err)

		// 检查技能是否已注销
		_, exists := registry.Get("test-skill")
		assert.False(t, exists)
	})

	t.Run("执行技能", func(t *testing.T) {
		registry := core.NewSkillRegistry()

		skillDef := core.SkillDefinition{
			ID:   "test-skill",
			Name: "测试技能",
		}

		executor := func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
			return &core.SkillResult{
				Success: true,
				Output:  map[string]interface{}{"result": "成功"},
			}, nil
		}

		err := registry.Register(skillDef, executor)
		assert.NoError(t, err)

		ctx := context.Background()
		skillCtx := &core.SkillContext{}
		result, err := registry.Execute(ctx, "test-skill", skillCtx)
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "成功", result.Output["result"])
	})

	t.Run("技能升级", func(t *testing.T) {
		registry := core.NewSkillRegistry()

		skillDef := core.SkillDefinition{
			ID:    "test-skill",
			Name:  "测试技能",
			Level: core.LevelNovice,
		}

		executor := func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
			return &core.SkillResult{Success: true}, nil
		}

		err := registry.Register(skillDef, executor)
		assert.NoError(t, err)

		// 设置升级规则
		rules := []core.SkillUpgradeRule{
			{
				FromLevel:   core.LevelNovice,
				ToLevel:     core.LevelApprentice,
				RequiredExp: 1,
			},
		}
		err = registry.SetUpgradeRules("test-skill", rules)
		assert.NoError(t, err)

		// 执行技能以获取经验
		ctx := context.Background()
		skillCtx := &core.SkillContext{}
		_, err = registry.Execute(ctx, "test-skill", skillCtx)
		assert.NoError(t, err)

		// 升级技能
		err = registry.Upgrade("test-skill")
		assert.NoError(t, err)

		// 检查技能等级
		skill, exists := registry.Get("test-skill")
		assert.True(t, exists)
		assert.Equal(t, core.LevelApprentice, skill.Level)
	})

	t.Run("技能链", func(t *testing.T) {
		registry := core.NewSkillRegistry()

		// 注册技能1
		skills := []struct {
			id       string
			name     string
			executor core.SkillExecutor
		}{
			{
				id:   "skill1",
				name: "技能1",
				executor: func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
					return &core.SkillResult{
						Success: true,
						Output:  map[string]interface{}{"step1": "完成"},
					}, nil
				},
			},
			{
				id:   "skill2",
				name: "技能2",
				executor: func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
					return &core.SkillResult{
						Success: true,
						Output:  map[string]interface{}{"step2": "完成"},
					}, nil
				},
			},
		}

		for _, skill := range skills {
			err := registry.Register(core.SkillDefinition{
				ID:   skill.id,
				Name: skill.name,
			}, skill.executor)
			assert.NoError(t, err)
		}

		// 创建技能链
		chain := core.NewSkillChain(registry)
		chain.Add("skill1", nil).Add("skill2", nil)

		// 执行技能链
		ctx := context.Background()
		results, err := chain.Execute(ctx, nil)
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})
}

func TestEffectCalculator(t *testing.T) {
	t.Run("应用和计算效果", func(t *testing.T) {
		calculator := core.NewEffectCalculator()

		effect := core.SkillEffect{
			Name:   "攻击加成",
			Value:  10.0,
			Duration: 0,
		}

		// 应用效果
		calculator.Apply("target1", "skill1", effect)

		// 计算效果
		result := calculator.Calculate("target1")
		assert.Equal(t, 10.0, result["攻击加成"])

		// 移除效果
		calculator.Remove("target1", "攻击加成")
		result = calculator.Calculate("target1")
		assert.Zero(t, result["攻击加成"])

		// 清除所有效果
		calculator.Apply("target1", "skill1", effect)
		calculator.Clear("target1")
		result = calculator.Calculate("target1")
		assert.Zero(t, result["攻击加成"])
	})
}

func TestSkillTriggerEngine(t *testing.T) {
	t.Run("触发技能", func(t *testing.T) {
		registry := core.NewSkillRegistry()
		engine := core.NewSkillTriggerEngine(registry)

		// 注册技能
		skillDef := core.SkillDefinition{
			ID:   "test-skill",
			Name: "测试技能",
		}

		executor := func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
			return &core.SkillResult{Success: true}, nil
		}

		err := registry.Register(skillDef, executor)
		assert.NoError(t, err)

		// 绑定触发器
		trigger := core.SkillTrigger{
			Type:      "damage",
			Threshold: 50.0,
		}
		engine.Bind("damage", "test-skill", trigger)

		// 触发事件
		ctx := context.Background()
		results := engine.Fire(ctx, "damage", 60.0, nil)
		assert.Len(t, results, 1)
		assert.True(t, results[0].Success)
	})
}