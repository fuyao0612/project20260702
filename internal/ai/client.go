package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"project20260702/internal/config"
)

// Client 是一个 OpenAI 兼容 Chat Completions API 客户端。
//
// 现在支持两种协议：
// 1. chat_completions：/chat/completions，使用 messages 和 choices。
// 2. responses：/responses，使用 input 和 output。
type Client struct {
	baseURL    string
	apiKey     string
	model      string
	protocol   string
	endpoint   string
	httpClient *http.Client
}

// NewClient 创建 AI 客户端。
func NewClient(cfg config.AIConfig) *Client {
	return NewClientWithConfig(RuntimeConfig{
		BaseURL:  cfg.BaseURL,
		APIKey:   cfg.APIKey,
		Model:    cfg.Model,
		Protocol: cfg.Protocol,
		Endpoint: cfg.Endpoint,
	})
}

// RuntimeConfig 是实际调用 AI API 时使用的配置。
//
// 它既可以来自 .env，也可以来自用户保存在数据库里的自定义配置。
type RuntimeConfig struct {
	BaseURL  string
	APIKey   string
	Model    string
	Protocol string
	Endpoint string
}

// NewClientWithConfig 使用运行时配置创建 AI 客户端。
func NewClientWithConfig(cfg RuntimeConfig) *Client {
	protocol := normalizeProtocol(cfg.Protocol, cfg.Endpoint)
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		endpoint = defaultEndpoint(protocol)
	}
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}

	return &Client{
		baseURL:  strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:   cfg.APIKey,
		model:    cfg.Model,
		protocol: protocol,
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TransactionDraft 是 AI 生成的账单草稿。
//
// 注意：这只是草稿，不直接写入数据库。
// 小程序端应让用户确认后，再调用 POST /api/transactions 保存。
type TransactionDraft struct {
	Type       string `json:"type"`
	Amount     int    `json:"amount"`
	Category   string `json:"category"`
	Note       string `json:"note"`
	HappenedAt string `json:"happened_at"`
}

type chatCompletionRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	Temperature    float64        `json:"temperature"`
	ResponseFormat responseFormat `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type responsesRequest struct {
	Model        string         `json:"model"`
	Instructions string         `json:"instructions,omitempty"`
	Input        string         `json:"input"`
	Temperature  float64        `json:"temperature,omitempty"`
	Text         *responsesText `json:"text,omitempty"`
}

type responsesText struct {
	Format responseFormat `json:"format"`
}

type responsesResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// GenerateTransactionDraft 根据用户输入文本生成账单草稿。
func (c *Client) GenerateTransactionDraft(text string, now time.Time, expenseCategories []string, incomeCategories []string) (TransactionDraft, error) {
	if c.apiKey == "" {
		return TransactionDraft{}, errors.New("AI_API_KEY 未配置")
	}

	prompt := buildTransactionDraftPrompt(text, now, expenseCategories, incomeCategories)

	var content string
	var err error

	switch c.protocol {
	case "responses":
		content, err = c.generateWithResponses(prompt)
	default:
		content, err = c.generateWithChatCompletions(prompt)
	}
	if err != nil {
		return TransactionDraft{}, err
	}

	var draft TransactionDraft
	if err := json.Unmarshal([]byte(content), &draft); err != nil {
		return TransactionDraft{}, fmt.Errorf("AI 返回内容不是合法 JSON：%w", err)
	}

	return draft, nil
}

func (c *Client) generateWithChatCompletions(prompt string) (string, error) {
	requestBody := chatCompletionRequest{
		Model: c.model,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: "你是一个记账助手。你只能输出 JSON，不要输出 Markdown，不要解释。",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.1,
		ResponseFormat: responseFormat{
			Type: "json_object",
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+c.endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("AI API 请求失败：%s", string(respBytes))
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(respBytes, &completion); err != nil {
		return "", err
	}

	if completion.Error != nil {
		return "", errors.New(completion.Error.Message)
	}

	if len(completion.Choices) == 0 {
		return "", errors.New("AI 没有返回结果")
	}

	content := strings.TrimSpace(completion.Choices[0].Message.Content)
	return content, nil
}

func (c *Client) generateWithResponses(prompt string) (string, error) {
	requestBody := responsesRequest{
		Model:        c.model,
		Instructions: "你是一个记账助手。你只能输出 JSON，不要输出 Markdown，不要解释。",
		Input:        prompt,
		Temperature:  0.1,
		Text: &responsesText{
			Format: responseFormat{
				Type: "json_object",
			},
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+c.endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("AI API 请求失败：%s", string(respBytes))
	}

	var result responsesResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", err
	}

	if result.Error != nil {
		return "", errors.New(result.Error.Message)
	}

	content := strings.TrimSpace(extractResponsesText(result))
	if content == "" {
		return "", errors.New("AI 没有返回文本结果")
	}

	return content, nil
}

// TestConnection 测试当前 AI 配置是否可以成功调用。
func (c *Client) TestConnection() error {
	if c.apiKey == "" {
		return errors.New("AI API Key 未配置")
	}

	if c.protocol == "responses" {
		return c.testResponses()
	}

	return c.testChatCompletions()
}

func (c *Client) testChatCompletions() error {
	requestBody := chatCompletionRequest{
		Model: c.model,
		Messages: []chatMessage{
			{
				Role:    "user",
				Content: "请只返回 JSON：{\"ok\":true}",
			},
		},
		Temperature: 0,
		ResponseFormat: responseFormat{
			Type: "json_object",
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+c.endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("AI API 请求失败：%s", string(respBytes))
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(respBytes, &completion); err != nil {
		return err
	}

	if completion.Error != nil {
		return errors.New(completion.Error.Message)
	}

	if len(completion.Choices) == 0 {
		return errors.New("AI 没有返回结果")
	}

	return nil
}

func (c *Client) testResponses() error {
	requestBody := responsesRequest{
		Model: c.model,
		Input: "Return ok.",
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+c.endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("AI API 请求失败：%s", string(respBytes))
	}

	var result responsesResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return err
	}

	if result.Error != nil {
		return errors.New(result.Error.Message)
	}

	if strings.TrimSpace(extractResponsesText(result)) == "" {
		return errors.New("AI 没有返回文本结果")
	}

	return nil
}

func extractResponsesText(result responsesResponse) string {
	if strings.TrimSpace(result.OutputText) != "" {
		return result.OutputText
	}

	for _, output := range result.Output {
		for _, content := range output.Content {
			if content.Type == "output_text" || content.Type == "text" {
				if strings.TrimSpace(content.Text) != "" {
					return content.Text
				}
			}
		}
	}

	return ""
}

func normalizeProtocol(protocol string, endpoint string) string {
	protocol = strings.TrimSpace(protocol)
	if protocol == "responses" || strings.Contains(endpoint, "responses") {
		return "responses"
	}

	return "chat_completions"
}

func defaultEndpoint(protocol string) string {
	if protocol == "responses" {
		return "/responses"
	}

	return "/chat/completions"
}

func buildTransactionDraftPrompt(text string, now time.Time, expenseCategories []string, incomeCategories []string) string {
	return fmt.Sprintf(`请从用户输入中提取一条记账草稿。

当前时间：%s
用户输入：%s

支出分类只能从这些值里选择：%s
收入分类只能从这些值里选择：%s

请严格返回 JSON 对象，字段如下：
{
  "type": "expense 或 income",
  "amount": 金额，整数，单位是分,
  "category": "分类名称",
  "note": "简短备注",
  "happened_at": "RFC3339 时间，例如 2026-07-03T12:00:00+08:00"
}

规则：
1. 日常消费通常是 expense，工资、红包、兼职、理财收益通常是 income。
2. 如果用户没有说明具体时间，使用当前时间。
3. 如果用户只说今天、昨天、前天，要按当前时间推断日期。
4. 金额必须转成分，例如 18.5 元返回 1850。
5. 分类必须从给定分类中选择，无法判断时使用“其他”。
6. note 尽量保留消费对象，例如 午饭、打车、买资料。`,
		now.Format(time.RFC3339),
		text,
		strings.Join(expenseCategories, "、"),
		strings.Join(incomeCategories, "、"),
	)
}
