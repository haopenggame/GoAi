package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-spring/ai/core"
)

func main() {
	fmt.Println("=== Go Spring AI - Skill 技能系统示例 ===")
	fmt.Println()

	exampleBasicRegistry()
	fmt.Println()
	exampleSkillExecution()
	fmt.Println()
	exampleSkillChain()
	fmt.Println()
	exampleEffectSystem()
	fmt.Println()
	exampleTriggerEngine()
	fmt.Println()
	exampleSkillUpgrade()
	fmt.Println()
	exampleRPGCombatSystem()
}

// 示例1：基础技能注册
func exampleBasicRegistry() {
	fmt.Println("--- 示例1：技能注册中心 ---")

	registry := core.NewSkillRegistry()

	skills := []struct {
		def      core.SkillDefinition
		executor core.SkillExecutor
	}{
		{
			def: core.SkillDefinition{
				ID:          "fireball",
				Name:        "火球术",
				Category:    "元素魔法",
				Level:       core.LevelNovice,
				Description: "发射一个火焰球，对目标造成火焰伤害",
			},
			executor: func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
				return &core.SkillResult{
					Success: true,
					Output: map[string]interface{}{
						"damage": 50,
						"type":   "fire",
					},
					Effects: []core.SkillEffect{
						{Name: "燃烧", Value: 5, Duration: 3 * time.Second},
					},
				}, nil
			},
		},
		{
			def: core.SkillDefinition{
				ID:          "ice-shield",
				Name:        "冰霜护盾",
				Category:    "防御魔法",
				Level:       core.LevelApprentice,
				Description: "召唤冰霜护盾，减少受到的物理伤害",
			},
			executor: func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
				return &core.SkillResult{
					Success: true,
					Output: map[string]interface{}{
						"defense_bonus": 30,
						"duration":      "10s",
					},
					Effects: []core.SkillEffect{
						{Name: "防御加成", Value: 30, Duration: 10 * time.Second},
					},
				}, nil
			},
		},
		{
			def: core.SkillDefinition{
				ID:          "heal",
				Name:        "治愈术",
				Category:    "治疗魔法",
				Level:       core.LevelNovice,
				Description: "恢复目标的生命值",
			},
			executor: func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
				return &core.SkillResult{
					Success: true,
					Output: map[string]interface{}{
						"heal_amount": 80,
					},
				}, nil
			},
		},
	}

	for _, s := range skills {
		err := registry.Register(s.def, s.executor)
		if err != nil {
			log.Fatal(err)
		}
	}

	allSkills := registry.List()
	fmt.Printf("已注册 %d 个技能:\n", len(allSkills))
	for _, s := range allSkills {
		fmt.Printf("  [%s] %s (Lv.%d) - %s\n", s.Category, s.Name, s.Level, s.Description)
	}
}

// 示例2：技能执行
func exampleSkillExecution() {
	fmt.Println("--- 示例2：技能执行 ---")

	registry := core.NewSkillRegistry()

	fireballDef := core.SkillDefinition{
		ID:          "fireball",
		Name:        "火球术",
		Category:    "元素魔法",
		Level:       core.LevelAdept,
		Description: "高级火球术，造成大量火焰伤害",
		Triggers: []core.SkillTrigger{
			{Type: "combat", Condition: "enemy_in_range", Threshold: 0, Cooldown: 5 * time.Second},
		},
	}

	fireballExec := func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
		power, _ := skillCtx.Parameters["power"].(float64)
		if power == 0 {
			power = 1.0
		}

		baseDamage := 50.0
		damage := int(float64(baseDamage) * power)

		result := &core.SkillResult{
			Success: true,
			Output: map[string]interface{}{
				"damage":   damage,
				"type":     "fire",
				"crit_hit": power >= 1.5,
				"target":   skillCtx.Target,
			},
			Effects: []core.SkillEffect{
				{
					Name:        "燃烧",
					Value:       float64(damage) * 0.1,
					Duration:    5 * time.Second,
					Stackable:   true,
					Description: "持续火焰伤害",
				},
			},
		}

		return result, nil
	}

	err := registry.Register(fireballDef, fireballExec)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	skillCtx := &core.SkillContext{
		Invoker:    "player-001",
		Target:     "goblin-01",
		Parameters: map[string]interface{}{"power": 1.8},
	}

	result, err := registry.Execute(ctx, "fireball", skillCtx)
	if err != nil {
		log.Fatalf("技能执行失败: %v", err)
	}

	fmt.Printf("技能执行结果:\n")
	fmt.Printf("  成功: %v\n", result.Success)
	fmt.Printf("  执行耗时: %v\n", result.Duration.Round(time.Millisecond))
	fmt.Printf("  伤害: %d\n", result.Output["damage"])
	fmt.Printf("  暴击: %v\n", result.Output["crit_hit"])
	fmt.Printf("  触发效果:\n")
	for _, effect := range result.Effects {
		fmt.Printf("    - %s: %.1f (持续%v)\n", effect.Name, effect.Value, effect.Duration)
	}

	exp, _ := registry.GetExperience("fireball")
	fmt.Printf("\n当前经验值: %d\n", exp)

	skill, _ := registry.Get("fireball")
	fmt.Printf("技能状态: %s\n", skill.State)
}

// 示例3：技能链
func exampleSkillChain() {
	fmt.Println("--- 示例3：技能链编排 ---")

	registry := core.NewSkillRegistry()

	registry.Register(core.SkillDefinition{
		ID:   "prepare",
		Name: "蓄力准备",
	}, func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
		return &core.SkillResult{
			Success: true,
			Output:  map[string]interface{}{"charge_level": 100, "buff_active": true},
		}, nil
	})

	registry.Register(core.SkillDefinition{
		ID:   "enhance",
		Name: "魔力增强",
	}, func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
		charge, _ := skillCtx.Parameters["charge_level"].(int)
		multiplier := float64(charge) / 100.0
		if multiplier < 1 {
			multiplier = 1.0
		}
		return &core.SkillResult{
			Success: true,
			Output:  map[string]interface{}{"power_multiplier": multiplier, "mana_cost": 20},
		}, nil
	})

	registry.Register(core.SkillDefinition{
		ID:   "release",
		Name: "释放攻击",
	}, func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
		power, _ := skillCtx.Parameters["power_multiplier"].(float64)
		damage := int(100 * power)
		return &core.SkillResult{
			Success: true,
			Output:  map[string]interface{}{"final_damage": damage, "attack_type": "magic_burst"},
		}, nil
	})

	chain := core.NewSkillChain(registry).
		Add("prepare", nil).
		Add("enhance", nil).
		Add("release", nil).
		OnError(func(err error, step core.SkillChainStep) error {
			fmt.Printf("  [错误] 技能 %s 执行失败: %v\n", step.SkillID, err)
			return nil
		})

	ctx := context.Background()
	results, err := chain.Execute(ctx, map[string]interface{}{"target": "boss-001"})
	if err != nil {
		log.Fatalf("技能链执行失败: %v", err)
	}

	fmt.Printf("技能链执行完成，共 %d 步:\n", len(results))
	for i, r := range results {
		fmt.Printf("  步骤%d: 耗时=%v\n", i+1, r.Duration.Round(time.Millisecond))
		for k, v := range r.Output {
			fmt.Printf("    %s = %v\n", k, v)
		}
	}

	finalDamage := results[len(results)-1].Output["final_damage"]
	fmt.Printf("最终伤害: %v\n", finalDamage)
}

// 示例4：效果计算系统
func exampleEffectSystem() {
	fmt.Println("--- 示例4：效果计算系统 ---")

	calculator := core.NewEffectCalculator()

	calculator.Apply("player-001", "buff-str", core.SkillEffect{
		Name:      "力量强化",
		Value:     15.0,
		Duration:  30 * time.Second,
		Stackable: false,
	})
	fmt.Printf("[效果] 玩家获得 力量强化 +15 (30s)\n")

	calculator.Apply("player-001", "potion-attack", core.SkillEffect{
		Name:      "攻击药水",
		Value:     25.0,
		Duration:  60 * time.Second,
		Stackable: false,
	})
	fmt.Printf("[效果] 玩家获得 攻击药水 +25 (60s)\n")

	calculator.Apply("player-001", "skill-berserk", core.SkillEffect{
		Name:      "狂暴",
		Value:     10.0,
		Duration:  15 * time.Second,
		Stackable: true,
	})
	fmt.Printf("[效果] 玩家获得 狂暴 +10 (15s, 可叠加)\n")

	calculator.Apply("player-001", "skill-berserk", core.SkillEffect{
		Name:      "狂暴",
		Value:     10.0,
		Duration:  15 * time.Second,
		Stackable: true,
	})
	fmt.Printf("[效果] 玩家获得 狂暴 +10 (叠加后=+20)\n")

	totalEffects := calculator.Calculate("player-001")
	fmt.Printf("\n玩家当前总效果:\n")
	for name, value := range totalEffects {
		fmt.Printf("  %s: +%.1f\n", name, value)
	}

	calculator.Remove("player-001", "攻击药水")
	fmt.Printf("\n[移除] 攻击药水效果已消失\n")

	totalEffects = calculator.Calculate("player-001")
	fmt.Printf("移除后的效果:\n")
	for name, value := range totalEffects {
		fmt.Printf("  %s: +%.1f\n", name, value)
	}

	calculator.Clear("player-001")
	fmt.Printf("\n[清除] 玩家所有效果已清除\n")

	totalEffects = calculator.Calculate("player-001")
	fmt.Printf("剩余效果数: %d\n", len(totalEffects))
}

// 示例5：触发引擎
func exampleTriggerEngine() {
	fmt.Println("--- 示例5：技能触发引擎 ---")

	registry := core.NewSkillRegistry()
	engine := core.NewSkillTriggerEngine(registry)

	registry.Register(core.SkillDefinition{
		ID:   "counter-attack",
		Name: "反击",
	}, func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
		damage, _ := skillCtx.Parameters["incoming_damage"].(float64)
		counterDamage := int(damage * 0.5)
		return &core.SkillResult{
			Success: true,
			Output:  map[string]interface{}{"counter_damage": counterDamage},
		}, nil
	})

	registry.Register(core.SkillDefinition{
		ID:   "evasion",
		Name: "闪避",
	}, func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
		return &core.SkillResult{
			Success: true,
			Output:  map[string]interface{}{"evaded": true},
		}, nil
	})

	registry.Register(core.SkillDefinition{
		ID:   "rage",
		Name: "怒气爆发",
	}, func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
		return &core.SkillResult{
			Success: true,
			Output:  map[string]interface{}{"rage_mode": true, "attack_boost": 50},
		}, nil
	})

	engine.Bind("take_damage", "counter-attack", core.SkillTrigger{
		Type:      "take_damage",
		Threshold: 30,
		Cooldown:  3 * time.Second,
	})

	engine.Bind("take_damage", "evasion", core.SkillTrigger{
		Type:      "take_damage",
		Threshold: 10,
		Cooldown:  2 * time.Second,
	})

	engine.Bind("low_health", "rage", core.SkillTrigger{
		Type:      "low_health",
		Threshold: 20,
		Cooldown:  10 * time.Second,
	})

	ctx := context.Background()

	fmt.Printf("场景1：受到轻微伤害 (伤害=8)\n")
	results := engine.Fire(ctx, "take_damage", 8.0, map[string]interface{}{"incoming_damage": 8})
	fmt.Printf("  触发技能数: %d (未达到阈值)\n\n", len(results))

	fmt.Printf("场景2：受到中等伤害 (伤害=35)\n")
	results = engine.Fire(ctx, "take_damage", 35.0, map[string]interface{}{"incoming_damage": 35})
	fmt.Printf("  触发技能数: %d\n", len(results))
	for _, r := range results {
		fmt.Printf("  - 反击伤害: %v\n", r.Output["counter_damage"])
	}

	fmt.Printf("\n场景3：生命值过低 (HP=15%%)\n")
	results = engine.Fire(ctx, "low_health", 15.0, nil)
	fmt.Printf("  触发技能数: %d\n", len(results))
	for _, r := range results {
		fmt.Printf("  - 怒气模式: %v, 攻击提升: %v\n", r.Output["rage_mode"], r.Output["attack_boost"])
	}

	fmt.Printf("\n场景4：再次触发反击 (冷却中)\n")
	results = engine.Fire(ctx, "take_damage", 40.0, map[string]interface{}{"incoming_damage": 40})
	fmt.Printf("  触发技能数: %d (处于冷却期)\n", len(results))
}

// 示例6：技能升级系统
func exampleSkillUpgrade() {
	fmt.Println("--- 示例6：技能升级系统 ---")

	registry := core.NewSkillRegistry()

	swordMasteryDef := core.SkillDefinition{
		ID:          "sword-mastery",
		Name:        "剑术精通",
		Category:    "战斗技能",
		Level:       core.LevelNovice,
		Description: "提升剑类武器的攻击力和命中率",
	}

	swordMasteryExec := func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
		level := 0
		if def, ok := registry.Get("sword-mastery"); ok {
			level = int(def.Level)
		}
		baseDamage := 10 + level*5
		return &core.SkillResult{
			Success: true,
			Output: map[string]interface{}{
				"bonus_damage": baseDamage,
				"accuracy":     80 + level*4,
			},
		}, nil
	}

	err := registry.Register(swordMasteryDef, swordMasteryExec)
	if err != nil {
		log.Fatal(err)
	}

	rules := []core.SkillUpgradeRule{
		{
			FromLevel:   core.LevelNovice,
			ToLevel:     core.LevelApprentice,
			RequiredExp: 3,
			Cost:        map[string]interface{}{"gold": 100},
			Conditions:  []string{"完成新手任务"},
		},
		{
			FromLevel:   core.LevelApprentice,
			ToLevel:     core.LevelAdept,
			RequiredExp: 10,
			Cost:        map[string]interface{}{"gold": 500},
			Conditions:  []string{"击败10个敌人"},
		},
		{
			FromLevel:   core.LevelAdept,
			ToLevel:     core.LevelExpert,
			RequiredExp: 30,
			Cost:        map[string]interface{}{"gold": 2000},
			Conditions:  []string{"完成精英挑战"},
		},
		{
			FromLevel:   core.LevelExpert,
			ToLevel:     core.LevelMaster,
			RequiredExp: 100,
			Cost:        map[string]interface{}{"gold": 10000},
			Conditions:  []string{"获得大师认可"},
		},
	}
	err = registry.SetUpgradeRules("sword-mastery", rules)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	skillCtx := &core.SkillContext{}

	levelNames := map[core.SkillLevel]string{
		core.LevelNovice:     "初学者",
		core.LevelApprentice: "学徒",
		core.LevelAdept:      "熟练",
		core.LevelExpert:     "专家",
		core.LevelMaster:     "大师",
	}

	fmt.Printf("初始状态: 剑术精通 Lv.%d (%s)\n", core.LevelNovice, levelNames[core.LevelNovice])

	var exp int
	var skill *core.SkillDefinition
	for useCount := 1; useCount <= 150; useCount++ {
		_, _ = registry.Execute(ctx, "sword-mastery", skillCtx)

		skill, _ = registry.Get("sword-mastery")
		exp, _ = registry.GetExperience("sword-mastery")

		currentRules := getUpgradeRulesForLevel(rules, skill.Level)
		if currentRules != nil && exp >= currentRules.RequiredExp {
			err = registry.Upgrade("sword-mastery")
			if err == nil {
				fmt.Printf(">>> 升级! 剑术精通 -> Lv.%d (%s), 经验=%d/%d\n",
					skill.Level, levelNames[skill.Level], exp, currentRules.RequiredExp)

				result, _ := registry.Execute(ctx, "sword-mastery", skillCtx)
				if result != nil {
					fmt.Printf("    新属性: 伤害加成=%v, 命中率=%v%%\n",
						result.Output["bonus_damage"], result.Output["accuracy"])
				}
			}
		}
	}

	skill, _ = registry.Get("sword-mastery")
	exp, _ = registry.GetExperience("sword-mastery")
	fmt.Printf("\n最终状态: Lv.%d (%s), 总使用次数=%d\n", skill.Level, levelNames[skill.Level], exp)
}

func getUpgradeRulesForLevel(rules []core.SkillUpgradeRule, level core.SkillLevel) *core.SkillUpgradeRule {
	for i := range rules {
		if rules[i].FromLevel == level {
			return &rules[i]
		}
	}
	return nil
}

// 示例7：完整 RPG 战斗系统演示
func exampleRPGCombatSystem() {
	fmt.Println("--- 示例7：RPG 战斗系统集成 ---")

	registry := core.NewSkillRegistry()
	calculator := core.NewEffectCalculator()
	engine := core.NewSkillTriggerEngine(registry)

	registry.Register(core.SkillDefinition{
		ID:           "slash",
		Name:         "斩击",
		Category:     "物理攻击",
		Level:        core.LevelNovice,
		Dependencies: []string{},
	}, func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
		effects := calculator.Calculate(skillCtx.Invoker)
		attackBonus, _ := effects["攻击力"]
		baseDamage := 30 + int(attackBonus)
		return &core.SkillResult{
			Success: true,
			Output: map[string]interface{}{
				"damage": baseDamage,
				"type":   "physical",
			},
		}, nil
	})

	registry.Register(core.SkillDefinition{
		ID:           "power-strike",
		Name:         "强力打击",
		Category:     "物理攻击",
		Level:        core.LevelApprentice,
		Dependencies: []string{},
	}, func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
		effects := calculator.Calculate(skillCtx.Invoker)
		attackBonus, _ := effects["攻击力"]
		baseDamage := 60 + int(attackBonus)*2
		return &core.SkillResult{
			Success: true,
			Output: map[string]interface{}{
				"damage": baseDamage,
				"type":   "physical_heavy",
			},
			Effects: []core.SkillEffect{
				{Name: "眩晕", Value: 1, Duration: 2 * time.Second},
			},
		}, nil
	})

	registry.Register(core.SkillDefinition{
		ID:           "battle-cry",
		Name:         "战吼",
		Category:     "增益",
		Level:        core.LevelNovice,
		Dependencies: []string{},
	}, func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
		calculator.Apply(skillCtx.Invoker, "battle-cry", core.SkillEffect{
			Name:     "攻击力",
			Value:    20.0,
			Duration: 15 * time.Second,
		})
		return &core.SkillResult{
			Success: true,
			Output:  map[string]interface{}{"buff_applied": true, "attack_boost": 20},
		}, nil
	})

	registry.Register(core.SkillDefinition{
		ID:           "iron-skin",
		Name:         "铁皮术",
		Category:     "防御",
		Level:        core.LevelNovice,
		Dependencies: []string{},
	}, func(ctx context.Context, skillCtx *core.SkillContext) (*core.SkillResult, error) {
		calculator.Apply(skillCtx.Invoker, "iron-skin", core.SkillEffect{
			Name:     "防御力",
			Value:    30.0,
			Duration: 20 * time.Second,
		})
		return &core.SkillResult{
			Success: true,
			Output:  map[string]interface{}{"defense_boost": 30},
		}, nil
	})

	engine.Bind("combat_start", "battle-cry", core.SkillTrigger{
		Type:      "combat_start",
		Threshold: 0,
		Cooldown:  30 * time.Second,
	})

	playerID := "hero-001"
	enemyID := "dragon-001"

	ctx := context.Background()

	fmt.Printf("=== 战斗开始: %s vs %s ===\n\n", playerID, enemyID)

	engine.Fire(ctx, "combat_start", 1, map[string]interface{}{})
	fmt.Printf("[战吼] 战斗开始，释放战吼!\n")

	comboChain := core.NewSkillChain(registry).
		Add("iron-skin", map[string]interface{}{"target": playerID}).
		Add("slash", map[string]interface{}{"target": enemyID}).
		Add("slash", map[string]interface{}{"target": enemyID}).
		Add("power-strike", map[string]interface{}{"target": enemyID}).
		SkipOnError()

	results, err := comboChain.Execute(ctx, map[string]interface{}{
		"invoker": playerID,
	})
	if err != nil {
		fmt.Printf("连招执行出错: %v\n", err)
	} else {
		fmt.Printf("\n--- 连招执行结果 ---\n")
		totalDamage := 0
		for i, r := range results {
			if dmg, ok := r.Output["damage"].(int); ok {
				totalDamage += dmg
			}
			stepName := ""
			switch i {
			case 0:
				stepName = "[铁皮术]"
			case 1:
				stepName = "[斩击#1]"
			case 2:
				stepName = "[斩击#2]"
			case 3:
				stepName = "[强力打击]"
			}
			fmt.Printf("%s ", stepName)
			for k, v := range r.Output {
				fmt.Printf("%s=%v ", k, v)
			}
			fmt.Printf("(耗时=%v)\n", r.Duration.Round(time.Millisecond))
		}
		fmt.Printf("\n总伤害: %d\n", totalDamage)
	}

	fmt.Printf("\n--- 玩家当前增益效果 ---\n")
	activeEffects := calculator.Calculate(playerID)
	for name, value := range activeEffects {
		fmt.Printf("  %s: +%.1f\n", name, value)
	}

	fmt.Printf("\n--- 技能经验统计 ---\n")
	allSkills := registry.List()
	for _, s := range allSkills {
		exp, _ := registry.GetExperience(s.ID)
		fmt.Printf("  %s (Lv.%d): 使用%d次\n", s.Name, s.Level, exp)
	}
}
