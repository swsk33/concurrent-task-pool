package concurrent_task_pool

import (
	"fmt"
	"testing"
	"time"
)

// 测试有返回值的并发任务池
func TestReturnableTaskPool(t *testing.T) {
	// 1.创建任务列表
	list := make([]*DownloadTask, 0)
	for i := 1; i <= 30; i++ {
		list = append(list, &DownloadTask{
			url:      fmt.Sprintf("http://example.com/file/%d.txt", i),
			filename: fmt.Sprintf("file-%d.txt", i),
		})
	}
	// 2.创建任务池（有返回值的）
	pool := NewReturnableTaskPool[*DownloadTask, string](3, list, func(task *DownloadTask) string {
		fmt.Printf("正在下载：%s...\n", task.filename)
		// 模拟执行任务
		time.Sleep(350 * time.Millisecond)
		// 返回文件名
		return task.filename
	}, func(tasks []*DownloadTask) {
		fmt.Println("接收到终止信号！")
		fmt.Println("当前任务：")
		for _, task := range tasks {
			fmt.Println(task.url)
		}
	})
	// 3.启动任务池
	resultList := pool.Start(true)
	// 4.执行完成，读取结果
	fmt.Println("执行完成！全部结果：")
	for _, result := range resultList {
		fmt.Println(result)
	}
}