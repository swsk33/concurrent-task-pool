package concurrent_task_pool

import (
	"fmt"
	"testing"
	"time"
)

// 测试有返回值的并发任务池
func TestReturnableTaskPool_Start(t *testing.T) {
	// 1.创建任务列表
	list := createTaskList()
	// 2.创建任务池（有返回值的）
	pool := NewReturnableTaskPool[*DownloadTask, string](3, list, func(task *DownloadTask, pool *ReturnableTaskPool[*DownloadTask, string]) string {
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

// 测试有返回值的并发任务池-重试
func TestReturnableTaskPool_Retry(t *testing.T) {
	// 1.创建任务队列
	list := createTaskListWithError()
	// 2.创建任务池，使用已有队列
	pool := NewReturnableTaskPool[*DownloadTask, string](3, list, func(task *DownloadTask, pool *ReturnableTaskPool[*DownloadTask, string]) string {
		fmt.Printf("正在下载：%s...\n", task.filename)
		// 模拟出现错误
		if task.url == "" {
			fmt.Println("出现错误！")
			task.url = fmt.Sprintf("http://example.com/file/%s", task.filename)
			// 将任务重新放回队列
			pool.TaskQueue.Offer(task)
			return ""
		}
		// 模拟执行任务
		time.Sleep(350 * time.Millisecond)
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

// 测试有返回值的并发任务池-中断
func TestReturnableTaskPool_Interrupt(t *testing.T) {
	// 1.创建任务队列
	list := createTaskListWithError()
	// 2.创建任务池，使用已有队列
	pool := NewReturnableTaskPool[*DownloadTask, string](3, list, func(task *DownloadTask, pool *ReturnableTaskPool[*DownloadTask, string]) string {
		fmt.Printf("正在下载：%s...\n", task.filename)
		// 模拟出现错误
		if task.url == "" {
			fmt.Println("出现错误！")
			// 中断任务池
			pool.Interrupt()
			fmt.Println("已中断整个任务池！")
			return ""
		}
		// 模拟执行任务
		time.Sleep(350 * time.Millisecond)
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