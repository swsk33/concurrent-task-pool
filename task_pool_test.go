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

// 创建示例任务对象列表
func createTaskList() []*DownloadTask {
	list := make([]*DownloadTask, 0)
	for i := 1; i <= 30; i++ {
		list = append(list, &DownloadTask{
			url:      fmt.Sprintf("http://example.com/file/%d.txt", i),
			filename: fmt.Sprintf("file-%d.txt", i),
		})
	}
	return list
}

// 创建带有错误的任务对象列表
func createTaskListWithError() []*DownloadTask {
	list := make([]*DownloadTask, 0)
	for i := 1; i <= 10; i++ {
		// 模拟第3个任务有错误
		if i == 3 {
			list = append(list, &DownloadTask{
				url:      "",
				filename: fmt.Sprintf("file-%d.txt", i),
			})
			continue
		}
		list = append(list, &DownloadTask{
			url:      fmt.Sprintf("http://example.com/file/%d.txt", i),
			filename: fmt.Sprintf("file-%d.txt", i),
		})
	}
	return list
}

// 测试无返回值的并发任务池
func TestTaskPool_Start(t *testing.T) {
	// 1.创建任务列表
	list := createTaskList()
	// 2.创建任务池
	pool := NewTaskPool[*DownloadTask](3, list, func(task *DownloadTask, pool *TaskPool[*DownloadTask]) {
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

// 测试无返回值的并发任务池-重试
func TestTaskPool_Retry(t *testing.T) {
	// 1.创建任务队列
	list := createTaskListWithError()
	// 2.创建任务池
	pool := NewTaskPool[*DownloadTask](3, list, func(task *DownloadTask, pool *TaskPool[*DownloadTask]) {
		fmt.Printf("正在下载：%s...\n", task.filename)
		// 模拟出现错误
		if task.url == "" {
			fmt.Println("出现错误！")
			task.url = fmt.Sprintf("http://example.com/file/%s", task.filename)
			// 将任务重新放回并发任务池的任务队列
			// pool的TaskQueue属性即为任务池中存放全部任务的队列
			// Offer方法可将对象入队列
			pool.TaskQueue.Offer(task)
			return
		}
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

// 测试无返回值的并发任务池-中断
func TestTaskPool_Interrupt(t *testing.T) {
	// 1.创建任务队列
	list := createTaskListWithError()
	// 2.创建任务池
	pool := NewTaskPool[*DownloadTask](3, list, func(task *DownloadTask, pool *TaskPool[*DownloadTask]) {
		fmt.Printf("正在下载：%s...\n", task.filename)
		// 模拟出现错误
		if task.url == "" {
			fmt.Println("出现错误！")
			// 调用pool对象的Interrupt方法直接中断整个任务池
			pool.Interrupt()
			fmt.Println("已结束任务池！")
			return
		}
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