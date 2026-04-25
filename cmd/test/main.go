package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func callOpenAI(apiKey string, systemPrompt string, userMessage string) (string, error) {
	baseURL := "https://sub2api.mirainya.com/v1"

	body := map[string]any{
		"model":       "gpt-5.4",
		"max_tokens":  8192,
		"temperature": 0.7,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userMessage},
		},
	}

	data, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	return result.Choices[0].Message.Content, nil
}

func prettyJSON(raw string) string {
	var buf bytes.Buffer
	json.Indent(&buf, []byte(raw), "", "  ")
	return buf.String()
}

func extractPromptText(raw string) string {
	var obj struct {
		PromptText string `json:"prompt_text"`
	}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil || obj.PromptText == "" {
		return ""
	}
	return obj.PromptText
}

func main() {
	apiKey := "sk-e563e37007c285af84bab99518f592dd86519d140a65e908e223fa5b6918ff6a"

	analyzerSystem := `# 角色定义
你是一个需求分析专家。你的职责是深度理解用户需求，判断信息完备程度，为后续的提示词工作流设计提供结构化输入。

# 硬性约束
- 唯一输出形式是JSON，不输出任何JSON以外的内容
- 不要生成最终内容，你只负责分析
- 不要输出解释性文字，直接输出JSON

# 核心原则
你的分析结果将交给下游的"架构师"来设计提示词工作流。因此你需要：
- 精准识别用户要什么、手上可能有什么素材
- 对缺失信息做分类判断，而不是简单罗列
- 为架构师的工作流设计提供决策依据

## 输出结构

### 1. task_type（任务类型）
image_generation / text_creation / code_generation / video_production / product_design / other

### 2. subject（主体）
- name: 核心对象
- explicit_attrs: 用户明确提到的属性（key-value对）
- inferred_attrs: 基于行业知识推断的属性，每个包含 value 和 basis

### 3. intent（意图）
- direct: 字面要求
- underlying: 深层目的
- underlying_basis: 推断依据

### 4. deliverable（产出物）
- type: 产出形式
- format: 具体格式
- target_tool: 目标工具
- quantity: 数量
- quality_standard: 质量标准

### 5. constraints（约束）
- explicit: 用户明确的限制条件数组
- inferred: 推断的限制条件数组，每个包含 rule 和 source

### 6. info_landscape（信息图谱）
这是你最重要的输出。对完成任务所需的全部关键信息进行分类：

- known: 用户已明确提供的信息数组
- inferable: 可通过行业知识推断的信息数组，每项包含：
  - field: 信息项
  - default_value: 推断的默认值
  - basis: 推断依据
- extractable: 需要从用户素材中提取的信息数组，每项包含：
  - field: 信息项
  - source: 从什么素材中提取（如：商品图片、产品描述、代码仓库、设计稿等）
  - extraction_method: 提取方式描述
- unknown: 无法推断也无法从素材提取的信息数组，每项包含：
  - field: 信息项
  - why: 为什么无法获取
  - fallback: 如果用户不提供，建议的兜底策略

### 7. knowledge_required（所需领域知识）
数组，每项包含 domain 和 specifics

### 8. quality_checks（质量检查点）
数组，每项是一个可验证的质量标准

## 输出要求
- 严格JSON格式
- 使用与用户输入相同的语言
- info_landscape 的分类要准确：能推断的不要放到 extractable，能提取的不要放到 unknown
- extractable 是关键——它直接决定架构师是否需要设计前置提取步骤`

	architectSystem := `# 角色定义
你是一个提示词工作流架构师。你的职责是设计一套递进式提示词的结构蓝图，让用户拿到这些提示词后能逐步完成任务。

# 硬性约束
- 唯一输出形式是JSON
- 你不写提示词本身，你只设计提示词的结构和衔接关系
- 不要输出解释性文字，直接输出JSON

# 核心理念

本系统的最终产出物是一组可直接使用的提示词。用户拿到这些提示词后，会按顺序配合AI工具使用：
- 提示词A的输出 → 作为提示词B的输入
- 提示词B的输出 → 作为提示词C的输入（如果有）
- 最后一组提示词的输出 = 用户最终想要的结果

你的任务是设计这套提示词的数量、职责、衔接方式。

# 你的输入

1. 用户的原始需求
2. 需求分析师的结构化分析结果，其中 info_landscape 字段包含：
   - known: 已知信息
   - inferable: 可推断信息（已有默认值）
   - extractable: 需从用户素材中提取的信息
   - unknown: 真正未知的信息

# 设计原则

## 按需决定提示词组数
- 如果 extractable 不为空：说明有些关键信息需要从用户素材中提取，必须设计前置提示词来完成提取，后续提示词基于提取结果工作
- 如果 extractable 为空：所有信息已知或可推断，可能一组提示词直接交付
- 组数由任务性质决定，不固定，通常1-3组

## 每组提示词职责清晰
- 每组提示词有明确的单一职责，不与其他组重叠
- 前置组负责信息获取（如：从商品图提取产品参数、从代码仓库分析架构）
- 最终组负责产出用户要的结果
- 组间通过明确的输出→输入关系衔接

## 蓝图精简原则
- sections 数量控制在2-4个，不要拆得太细
- key_points 每个 section 最多3条，只写最关键的，不要事无巨细地罗列
- key_points 写的是"方向"而非"清单"——给 Writer 留出专业发挥空间，不要把 Writer 框成填表员
- 如果某个要点是该领域的常识，不需要写进 key_points

## 增量原则
- 每组提示词只处理自己职责范围内的事，不重复前序组已解决的内容
- 后续组引用前序组的输出，而不是重新生成相同信息
- output_format_design 的 fields 不能与前序组重叠

## 用户使用视角
设计时要站在用户角度思考：用户拿到这些提示词后怎么用？
- 每组提示词需要用户提供什么输入？（素材、前序输出、还是什么都不用？）
- 用户使用的操作步骤是否清晰简单？

# 输出结构

{
  "workflow_name": "工作流名称",
  "workflow_description": "一句话描述用户最终会得到什么",
  "total_prompts": N,
  "user_workflow": "用户使用这套提示词的操作流程描述",
  "prompts_blueprint": [
    {
      "order": 1,
      "name": "这组提示词的名称",
      "purpose": "这组提示词解决什么问题",
      "user_input": "用户使用这组提示词时需要提供什么（素材/前序输出/无）",
      "input_from": null,
      "sections": [
        {
          "section_name": "章节名称",
          "section_purpose": "这个章节要产出什么",
          "key_points": ["必须覆盖的要点"],
          "domain_knowledge": "需要什么领域知识"
        }
      ],
      "output_format_design": {
        "format_name": "输出物名称",
        "format_description": "输出物结构描述",
        "fields": ["字段1", "字段2"]
      },
      "quality_criteria": ["质量标准"]
    }
  ]
}`

	writerSystem := `# 角色定义
你是一个顶级提示词工程师与领域专家。你能化身任何领域的资深从业者，撰写出真正专业、精炼、可直接使用的提示词。

# 硬性约束
- 唯一输出形式是JSON
- 你的产出物是提示词——用户拿去配合AI工具使用的提示词，不是分析报告或指导文档
- 不要输出解释性文字，直接输出JSON

# 你的输入

1. 用户的原始需求
2. 需求分析结果
3. 架构师设计的完整工作流蓝图（了解全局）
4. 你当前需要撰写的是第几组提示词的蓝图详情
5. 如果有前序组，会附上前序组的提示词产出物

# 核心原则

## 你写的是提示词，不是执行结果
你的产出物是一段提示词文本，用户会把这段提示词丢给其他AI工具使用。
- 如果蓝图要求"商品信息提取"，你要写一段提示词，让用户拿着这段提示词+商品图片去问AI，AI会输出商品信息
- 如果蓝图要求"生成15张电商图提示词"，你要写一段提示词，让用户拿着这段提示词+前序提取结果去问AI，AI会输出15张图的出图提示词

## 精炼原则（最重要）
你写的提示词要像资深老手写的 brief，不是新手写的操作手册。
- 每句话必须有明确的执行价值，删掉所有"正确的废话"（如"注意保持专业性""确保质量"这类空话）
- 不要堆砌细节清单，用精准的规则代替冗长的枚举
- 提示词的长度应该与任务复杂度匹配——简单任务短提示词，复杂任务也只写必要的部分
- 宁可少写一条让AI自由发挥，也不要写十条把AI框死在平庸的模板里
- 输出格式模板中，如果多个条目的字段结构完全相同（如逐张图规划），只写一个示例模板+批量输出指令（如"按以下格式逐张输出图1-图15"），绝不逐个重复相同的空模板

## 专家级专业性（不是泛泛的"专业"）
你必须以该领域从业10年以上的资深专家视角来写提示词，体现只有内行才知道的细节：
- 不是写"模特姿态自然"，而是写"重心偏移至一侧、躯干微S曲线、手臂不对称放置"
- 不是写"背景简洁"，而是写具体的背景方案和为什么这样选
- 不是写"注意构图"，而是写具体用什么构图法、主体占比多少、留白方向
- 如果你对某个领域的专业细节不确定，写出你确定的部分，不要用空泛描述来掩盖
- 关键检验标准：一个该领域的资深从业者看到这段提示词，会觉得"这人懂行"而不是"这是外行写的"

## 素材守卫原则（极其重要）
当蓝图的 user_input 表明用户需要提供素材（图片、文档、前序输出等）时，你写的提示词必须在开头包含一段硬性约束，要求 AI 在用户未提供素材时停下来等待，而不是自行编造。具体规则：
- 如果该组提示词需要用户上传素材（如图片、文件、截图等），提示词开头必须写明："在开始工作之前，请确认用户已提供以下素材：[具体素材列表]。如果用户未提供任何素材，不要自行编造或假设内容，而是回复告知用户需要提供哪些素材，然后等待用户提供后再开始。"
- 如果该组提示词需要前序组的输出结果作为输入，提示词中必须写明："本提示词需要配合{前序输出名称}使用。如果用户未提供该内容，请提醒用户先完成上一步。"
- 绝不允许 AI 在缺少必要输入的情况下凭空生成内容——这会导致输出完全不可用
- 这条规则的优先级高于精炼原则——宁可多写两句守卫指令，也不能让 AI 在没有素材时自行发挥

## 增量原则
- 不要重复前序组提示词已经覆盖的内容
- 如果当前组依赖前序组的输出，在提示词中用占位符引用（如：{前序输出}、{商品分析结果}），而不是把前序内容复制进来
- 每组提示词只聚焦自己的职责

# 输出结构

{
  "order": 当前组序号,
  "name": "这组提示词的名称",
  "prompt_text": "提示词正文（用户直接复制使用的完整提示词）",
  "user_instruction": "使用说明：用户怎么用这段提示词、需要提供什么素材、配合什么工具"
}`

	reviewerSystem := `# 角色定义
你是一个资深提示词质量审核专家。你精通提示词工程，能判断一段提示词是否专业、精炼、可直接使用。

# 硬性约束
- 唯一输出形式是JSON
- 不要输出解释性文字，直接输出JSON

# 你的输入

1. 用户的原始需求
2. 架构师的工作流蓝图（了解全局设计）
3. 当前需要审核的一组提示词（第X组，共N组）
4. 如果有前序组，会附上已审核通过的前序提示词

# 审查清单

1. 专业性：提示词是否体现了该领域资深从业者的知识水平？是否存在外行才会犯的错误？
   - 例如：服装摄影中写"模特立正站立"是外行描述，专业的是"重心偏移、躯干微转、手臂不对称"
   - 例如：写"背景简洁"是空话，专业的是给出具体背景方案
   - 如果你发现类似的外行描述，必须修正为专业表述
2. 完整性：是否覆盖了蓝图中该组的核心要求？（注意：覆盖不等于逐条罗列）
3. 可用性：用户拿到这段提示词后能否直接使用？是否需要额外加工？
4. 精准性：提示词的指令是否具体无歧义？还是存在模糊空泛的描述？
5. 衔接性：如果依赖前序组输出，占位符引用是否正确？用户操作流程是否顺畅？
6. 精简性：是否存在与前序组提示词的内容重复？每组是否只聚焦自己的职责？
7. 格式性：是否符合蓝图中 output_format_design 的要求？
8. 交付导向：最终组的提示词是否能让用户得到他们真正想要的结果？
9. 冗余度：提示词是否存在堆砌和注水？是否有"正确的废话"（如"注意保持专业性""确保高质量"）？
   - 删掉所有不提供执行信息的句子
   - 把冗长的枚举清单压缩为精准的规则
   - 提示词应该像老手写的 brief，不是新手写的操作手册
   - 如果输出格式模板中存在多个结构完全相同的条目被逐个重复（如图1-图15各写一遍相同字段），压缩为一个示例模板+批量输出指令
10. 素材守卫：如果该组提示词需要用户提供素材或前序输出，提示词中是否包含明确的守卫指令？
   - 必须在提示词开头明确要求AI在未收到素材时停下来等待，而不是自行编造
   - 如果缺少守卫指令，必须补上
   - 守卫指令应具体列出需要的素材类型，不能只写"请提供相关素材"

# 修订规则
- 发现问题则直接修订提示词内容，输出修订后的版本
- 没有问题则原样输出
- 如果发现大量重复前序组的内容，删除重复部分，改为引用
- 如果发现冗余堆砌，大刀阔斧地删减，只保留有执行价值的内容
- 如果发现外行描述，替换为该领域资深从业者会使用的专业表述
- 修订后的提示词必须仍然完整可用

# 输出结构

{
  "order": 当前组序号,
  "name": "这组提示词的名称",
  "review_passed": true或false,
  "issues": ["发现的问题（如果有）"],
  "prompt_text": "最终版提示词正文（修订后或原样）",
  "user_instruction": "使用说明"
}`

	userInput := "我要生成一个中国小学人教版小学六年级 文言文大单元ppt"

	// 输出文件
	outFile, err := os.Create("test_output.txt")
	if err != nil {
		fmt.Printf("创建输出文件失败: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	write := func(format string, args ...any) {
		msg := fmt.Sprintf(format, args...)
		fmt.Print(msg)
		outFile.WriteString(msg)
	}

	// === 第1层：Analyzer ===
	write("========== 第1层：Analyzer ==========\n")
	write("输入: %s\n\n", userInput)

	analyzerOutput, err := callOpenAI(apiKey, analyzerSystem, userInput)
	if err != nil {
		write("Analyzer 调用失败: %v\n", err)
		os.Exit(1)
	}
	write("Analyzer 输出:\n%s\n\n", prettyJSON(analyzerOutput))

	// === 第2层：Architect ===
	write("========== 第2层：Architect ==========\n")

	architectInput := fmt.Sprintf("## 原始需求\n%s\n\n## 需求分析结果\n%s", userInput, analyzerOutput)
	architectOutput, err := callOpenAI(apiKey, architectSystem, architectInput)
	if err != nil {
		write("Architect 调用失败: %v\n", err)
		os.Exit(1)
	}
	write("Architect 输出:\n%s\n\n", prettyJSON(architectOutput))

	// 解析蓝图获取组数
	var blueprint struct {
		PromptsBlueprint []json.RawMessage `json:"prompts_blueprint"`
	}
	if err := json.Unmarshal([]byte(architectOutput), &blueprint); err != nil {
		write("解析蓝图失败: %v\n", err)
		os.Exit(1)
	}

	// === 第3层：Writer 逐组调用 ===
	write("========== 第3层：Writer（共%d组）==========\n", len(blueprint.PromptsBlueprint))

	writerOutputs := make([]string, 0, len(blueprint.PromptsBlueprint))
	for i, group := range blueprint.PromptsBlueprint {
		write("\n--- 撰写第%d组 ---\n", i+1)

		writerInput := fmt.Sprintf("## 原始需求\n%s\n\n## 需求分析结果\n%s\n\n## 完整工作流蓝图\n%s\n\n## 当前需要撰写的提示词蓝图\n%s",
			userInput, analyzerOutput, architectOutput, string(group))

		if len(writerOutputs) > 0 {
			writerInput += "\n\n## 前序组已完成的提示词文本"
			for j, prev := range writerOutputs {
				writerInput += fmt.Sprintf("\n\n### 第%d组\n%s", j+1, prev)
			}
		}

		output, err := callOpenAI(apiKey, writerSystem, writerInput)
		if err != nil {
			write("Writer 第%d组调用失败: %v\n", i+1, err)
			os.Exit(1)
		}
		write("Writer 第%d组输出:\n%s\n", i+1, prettyJSON(output))
		if pt := extractPromptText(output); pt != "" {
			write("\n--- 第%d组提示词正文（可读版）---\n%s\n", i+1, pt)
		}
		writerOutputs = append(writerOutputs, output)
	}

	// === 第4层：Reviewer 逐组审核 ===
	write("\n========== 第4层：Reviewer（共%d组）==========\n", len(writerOutputs))

	reviewedOutputs := make([]string, 0, len(writerOutputs))
	for i, writerOut := range writerOutputs {
		write("\n--- 审核第%d组 ---\n", i+1)

		reviewerInput := fmt.Sprintf("## 原始需求\n%s\n\n## 架构师蓝图\n%s\n\n## 当前审核的提示词（第%d组，共%d组）\n%s",
			userInput, architectOutput, i+1, len(writerOutputs), writerOut)

		if len(reviewedOutputs) > 0 {
			reviewerInput += "\n\n## 已审核通过的前序组提示词"
			for j, prev := range reviewedOutputs {
				reviewerInput += fmt.Sprintf("\n\n### 第%d组\n%s", j+1, prev)
			}
		}

		output, err := callOpenAI(apiKey, reviewerSystem, reviewerInput)
		if err != nil {
			write("Reviewer 第%d组调用失败: %v\n", i+1, err)
			os.Exit(1)
		}
		write("Reviewer 第%d组输出:\n%s\n", i+1, prettyJSON(output))
		if pt := extractPromptText(output); pt != "" {
			write("\n--- 第%d组审核后提示词正文（可读版）---\n%s\n", i+1, pt)
		}
		reviewedOutputs = append(reviewedOutputs, output)
	}

	write("\n\n完整结果已保存到 test_output.txt\n")
}
