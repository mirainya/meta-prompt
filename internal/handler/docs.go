package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DocsHandler struct{}

func NewDocsHandler() *DocsHandler { return &DocsHandler{} }

func (h *DocsHandler) OpenAPISpec(c *gin.Context) {
	spec := gin.H{
		"openapi": "3.0.3",
		"info": gin.H{
			"title":       "Meta Prompt Open API",
			"version":     "1.0.0",
			"description": "AI 提示词推演平台开放接口，使用 API Key 鉴权。",
		},
		"servers": []gin.H{{"url": "/open/v1"}},
		"security": []gin.H{{"ApiKeyAuth": []string{}}},
		"components": gin.H{
			"securitySchemes": gin.H{
				"ApiKeyAuth": gin.H{
					"type": "apiKey",
					"in":   "header",
					"name": "X-API-Key",
				},
			},
		},
		"paths": gin.H{
			"/generate": gin.H{
				"post": gin.H{
					"summary":     "创建推演任务",
					"description": "提交需求文本，异步生成提示词。消耗 1 积分。",
					"requestBody": gin.H{
						"required": true,
						"content": gin.H{
							"application/json": gin.H{
								"schema": gin.H{
									"type":     "object",
									"required": []string{"input"},
									"properties": gin.H{
										"input":        gin.H{"type": "string", "description": "需求描述"},
										"llm_provider": gin.H{"type": "string", "description": "LLM 提供商（可选，默认使用系统配置）"},
										"mode":         gin.H{"type": "string", "enum": []string{"sync", "async"}, "description": "sync=同步等待结果，async=立即返回 task_id（默认）"},
										"webhook_url":  gin.H{"type": "string", "description": "完成后回调地址（可选）"},
									},
								},
							},
						},
					},
					"responses": gin.H{
						"202": gin.H{"description": "异步模式：返回 task_id"},
						"200": gin.H{"description": "同步模式：返回完整结果"},
						"402": gin.H{"description": "积分不足"},
					},
				},
			},
			"/tasks/{id}": gin.H{
				"get": gin.H{
					"summary":     "查询任务状态",
					"description": "根据 task_id 查询推演进度和结果。",
					"parameters": []gin.H{
						{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "integer"}},
					},
					"responses": gin.H{
						"200": gin.H{"description": "任务详情（含 status, current_step, reviewer_output 等）"},
					},
				},
			},
			"/tasks/{id}/stream": gin.H{
				"get": gin.H{
					"summary":     "SSE 实时进度流",
					"description": "通过 Server-Sent Events 实时接收推演进度。事件类型：running, done, failed。",
					"parameters": []gin.H{
						{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "integer"}},
					},
				},
			},
			"/tasks/{id}/export": gin.H{
				"get": gin.H{
					"summary":     "导出结果",
					"description": "导出推演结果为 Markdown 或 JSON 格式。",
					"parameters": []gin.H{
						{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "integer"}},
						{"name": "format", "in": "query", "schema": gin.H{"type": "string", "enum": []string{"markdown", "json"}}},
					},
				},
			},
			"/tasks/{id}/cancel": gin.H{
				"post": gin.H{
					"summary":     "取消任务",
					"description": "取消正在运行的推演任务，退还积分。",
					"parameters": []gin.H{
						{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "integer"}},
					},
				},
			},
			"/providers": gin.H{
				"get": gin.H{
					"summary":     "获取可用模型列表",
					"description": "返回当前已启用的 LLM 提供商及其模型。",
				},
			},
		},
	}
	c.JSON(http.StatusOK, spec)
}
