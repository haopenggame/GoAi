package zhipu

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-spring/ai/core"
)

type TaskQuery struct {
	client *Client
}

func NewTaskQuery(client *Client) *TaskQuery {
	return &TaskQuery{client: client}
}

func (q *TaskQuery) QueryTask(ctx context.Context, taskID string) (core.AsyncTaskResult, error) {
	if taskID == "" {
		return core.AsyncTaskResult{}, fmt.Errorf("taskID不能为空")
	}

	path := fmt.Sprintf("/async-result/%s", taskID)
	respBody, err := q.client.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return core.AsyncTaskResult{}, fmt.Errorf("查询异步任务失败: %w", err)
	}

	var resp asyncResultResponse
	if err := jsonUnmarshal(respBody, &resp); err != nil {
		return core.AsyncTaskResult{}, fmt.Errorf("反序列化响应失败: %w", err)
	}

	return toCoreTaskResult(resp, "zhipu"), nil
}

func (q *TaskQuery) WaitForTask(ctx context.Context, taskID string, options core.PollOptions) (core.AsyncTaskResult, error) {
	if taskID == "" {
		return core.AsyncTaskResult{}, fmt.Errorf("taskID不能为空")
	}

	interval := options.Interval
	if interval <= 0 {
		interval = 5
	}

	maxRetries := options.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 60
	}

	for i := 0; i < maxRetries; i++ {
		result, err := q.QueryTask(ctx, taskID)
		if err != nil {
			return core.AsyncTaskResult{}, err
		}

		switch result.Task.TaskStatus {
		case core.TaskSuccess:
			return result, nil
		case core.TaskFail:
			return result, fmt.Errorf("异步任务失败，taskID: %s", taskID)
		case core.TaskProcessing:
			select {
			case <-ctx.Done():
				return core.AsyncTaskResult{}, fmt.Errorf("等待任务被取消: %w", ctx.Err())
			case <-time.After(time.Duration(interval) * time.Second):
			}
		default:
			return result, fmt.Errorf("未知的任务状态: %s", result.Task.TaskStatus)
		}
	}

	return core.AsyncTaskResult{}, fmt.Errorf("等待任务超时，已重试 %d 次", maxRetries)
}

func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
