package concurrent_task_pool

import (
	"fmt"
	"testing"
	"time"
)

// DownloadTask 一个示例下载任务（参数）表示
type DownloadTask struct {
	url      string
	filename string
}

// 测试无返回值的并发任务池
func TestTaskPool(t *testing.T) {
	// 1.创建任务列表
	list := make([]*DownloadTask, 0)
	for i := 1; i <= 30; i++ {
		list = append(list, &DownloadTask{
			url:      fmt.Sprintf("http://example.com/file/%d.txt", i),
			filename: fmt.Sprintf("file-%d.txt", i),
		})
	}
	// 2.创建任务池
	pool := NewTaskPool[*DownloadTask](3, list, func(task *DownloadTask) {
		fmt.Printf("正在下载：%s...\n", task.filename)
		// 模拟执行任务
		time.Sleep(350 * time.Millisecond)
		fmt.Printf("下载%s完成！\n", task.filename)
	}, func(tasks []*DownloadTask) {
		fmt.Println("接收到终止信号！")
		fmt.Println("当前任务：")
		for _, task := range tasks {
			fmt.Println(task.url)
		}
	})
	// 3.启动任务池
	pool.Start()
}