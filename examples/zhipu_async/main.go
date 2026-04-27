package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-spring/ai/core"
	"github.com/go-spring/ai/zhipu"
)

func main() {
	apiKey := "your-zhipu-api-key"
	if apiKey == "your-zhipu-api-key" {
		log.Println("请设置有效的智谱API密钥")
		return
	}

	ctx := context.Background()

	client := zhipu.NewClient(apiKey,
		zhipu.WithDebug(true),
	)

	imageModel := zhipu.NewImageModel(client)
	videoModel := zhipu.NewVideoModel(client)
	taskQuery := zhipu.NewTaskQuery(client)

	fmt.Println("=== 图像生成示例 ===")
	watermark := true
	imageTask, err := imageModel.CreateImageTask(ctx, "一只可爱的小猫咪，坐在阳光明媚的窗台上", core.AsyncImageOptions{
		Model:            "glm-image",
		Size:             "1280x1280",
		Quality:          "hd",
		WatermarkEnabled: &watermark,
	})
	if err != nil {
		log.Printf("创建图像任务失败: %v\n", err)
	} else {
		fmt.Printf("图像任务已创建，ID: %s，状态: %s\n", imageTask.ID, imageTask.TaskStatus)

		imageResult, err := taskQuery.WaitForTask(ctx, imageTask.ID, core.PollOptions{
			Interval:   5,
			MaxRetries: 60,
		})
		if err != nil {
			log.Printf("等待图像任务失败: %v\n", err)
		} else {
			fmt.Printf("图像任务完成，状态: %s\n", imageResult.Task.TaskStatus)
			for i, img := range imageResult.ImageResults {
				fmt.Printf("图像 %d: %s\n", i+1, img.URL)
			}
		}
	}

	fmt.Println("\n=== 视频生成示例 ===")
	withAudio := true
	videoTask, err := videoModel.CreateVideoTask(ctx, "A cat is playing with a ball.", core.AsyncVideoOptions{
		Model:     "cogvideox-3",
		Quality:   "quality",
		WithAudio: &withAudio,
		Size:      "1920x1080",
		FPS:       30,
		Duration:  5,
	})
	if err != nil {
		log.Printf("创建视频任务失败: %v\n", err)
	} else {
		fmt.Printf("视频任务已创建，ID: %s，状态: %s\n", videoTask.ID, videoTask.TaskStatus)

		videoResult, err := taskQuery.WaitForTask(ctx, videoTask.ID, core.PollOptions{
			Interval:   10,
			MaxRetries: 60,
		})
		if err != nil {
			log.Printf("等待视频任务失败: %v\n", err)
		} else {
			fmt.Printf("视频任务完成，状态: %s\n", videoResult.Task.TaskStatus)
			for i, vid := range videoResult.VideoResults {
				fmt.Printf("视频 %d: %s (封面: %s)\n", i+1, vid.URL, vid.CoverImageURL)
			}
		}
	}

	fmt.Println("\n测试完成！")
}
