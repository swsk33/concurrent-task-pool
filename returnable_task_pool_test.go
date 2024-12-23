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
	pool := NewReturnableTaskPool[*DownloadTask, string](3, 0, 0, list,
		// 每个任务的自定义执行逻辑回调函数
		func(task *DownloadTask, pool *ReturnableTaskPool[*DownloadTask, string]) string {
			fmt.Printf("正在下载：%s...\n", task.Filename)
			// 模拟执行任务
			for i := 0; i < 4; i++ {
				task.Process += 25
				time.Sleep(100 * time.Millisecond)
			}
			// 返回文件名
			return task.Filename
		},
		// 接收到终止信号时的停机逻辑回调函数
		func(pool *ReturnableTaskPool[*DownloadTask, string]) {
			fmt.Println("接收到终止信号！")
			fmt.Println("当前任务：")
			for _, task := range pool.GetRunningTaskList() {
				fmt.Println(task.Url)
			}
		}, nil)
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
	pool := NewReturnableTaskPool[*DownloadTask, string](3, 0, 0, list,
		// 每个任务的自定义执行逻辑回调函数
		func(task *DownloadTask, pool *ReturnableTaskPool[*DownloadTask, string]) string {
			fmt.Printf("正在下载：%s...\n", task.Filename)
			// 模拟出现错误
			if task.Url == "" {
				fmt.Println("出现错误！")
				task.Url = fmt.Sprintf("http://example.com/file/%s", task.Filename)
				// 重试任务，将任务重新放回队列
				pool.Retry(task)
				return ""
			}
			// 模拟执行任务
			for i := 0; i < 4; i++ {
				task.Process += 25
				time.Sleep(100 * time.Millisecond)
			}
			return task.Filename
		},
		// 接收到终止信号时的停机逻辑回调函数
		func(pool *ReturnableTaskPool[*DownloadTask, string]) {
			fmt.Println("接收到终止信号！")
			fmt.Println("当前任务：")
			for _, task := range pool.GetRunningTaskList() {
				fmt.Println(task.Url)
			}
		}, nil)
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
	pool := NewReturnableTaskPool[*DownloadTask, string](3, 0, 0, list,
		// 每个任务的自定义执行逻辑回调函数
		func(task *DownloadTask, pool *ReturnableTaskPool[*DownloadTask, string]) string {
			fmt.Printf("正在下载：%s...\n", task.Filename)
			// 模拟出现错误
			if task.Url == "" {
				fmt.Println("出现错误！")
				// 中断任务池
				pool.Interrupt()
				fmt.Println("已中断整个任务池！")
				return ""
			}
			// 模拟执行任务
			time.Sleep(350 * time.Millisecond)
			return task.Filename
		},
		// 接收到终止信号时的停机逻辑回调函数
		func(pool *ReturnableTaskPool[*DownloadTask, string]) {
			fmt.Println("接收到终止信号！")
			fmt.Println("当前任务：")
			for _, task := range pool.GetRunningTaskList() {
				fmt.Println(task.Url)
			}
		}, nil)
	// 3.启动任务池
	resultList := pool.Start(true)
	// 4.执行完成，读取结果
	fmt.Println("执行完成！全部结果：")
	for _, result := range resultList {
		fmt.Println(result)
	}
}