package concurrent_task_pool

import (
	"fmt"
	"testing"
	"time"
)

// DownloadTask 一个示例下载任务（参数）表示
type DownloadTask struct {
	// 下载地址
	Url string
	// 文件名
	Filename string
	// 进度(0-100)
	Process int
}

// 创建示例任务对象列表
func createTaskList() []*DownloadTask {
	list := make([]*DownloadTask, 0)
	for i := 1; i <= 30; i++ {
		list = append(list, &DownloadTask{
			Url:      fmt.Sprintf("http://example.com/file/%d.txt", i),
			Filename: fmt.Sprintf("file-%d.txt", i),
			Process:  0,
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
				Url:      "",
				Filename: fmt.Sprintf("file-%d.txt", i),
				Process:  0,
			})
			continue
		}
		list = append(list, &DownloadTask{
			Url:      fmt.Sprintf("http://example.com/file/%d.txt", i),
			Filename: fmt.Sprintf("file-%d.txt", i),
			Process:  0,
		})
	}
	return list
}

// 测试无返回值的并发任务池
func TestTaskPool_Start(t *testing.T) {
	// 1.创建任务列表
	list := createTaskList()
	// 2.创建任务池
	pool := NewTaskPool[*DownloadTask](3, 0, 0, list,
		// 每个任务的自定义执行逻辑回调函数
		func(task *DownloadTask, pool *TaskPool[*DownloadTask]) {
			fmt.Printf("正在下载：%s...\n", task.Filename)
			// 模拟执行任务
			for i := 0; i < 4; i++ {
				task.Process += 25
				time.Sleep(100 * time.Millisecond)
			}
			fmt.Printf("下载%s完成！\n", task.Filename)
		},
		// 接收到终止信号时的停机逻辑回调函数
		func(pool *TaskPool[*DownloadTask]) {
			fmt.Println("接收到终止信号！")
			fmt.Println("当前任务：")
			for _, task := range pool.GetRunningTaskList() {
				fmt.Println(task.Url)
			}
		}, nil)
	// 3.启动任务池
	pool.Start()
}

// 测试无返回值的并发任务池-重试
func TestTaskPool_Retry(t *testing.T) {
	// 1.创建任务队列
	list := createTaskListWithError()
	// 2.创建任务池
	pool := NewTaskPool[*DownloadTask](3, 0, 0, list,
		// 每个任务的自定义执行逻辑回调函数
		func(task *DownloadTask, pool *TaskPool[*DownloadTask]) {
			fmt.Printf("正在下载：%s...\n", task.Filename)
			// 模拟出现错误
			if task.Url == "" {
				fmt.Println("出现错误！")
				task.Url = fmt.Sprintf("http://example.com/file/%s", task.Filename)
				// 稍后重试任务
				// 调用并发任务池对象的Retry方法，传入当前任务对象，即可将任务重新放回并发任务池的任务队列
				pool.Retry(task)
				return
			}
			// 模拟执行任务
			for i := 0; i < 4; i++ {
				task.Process += 25
				time.Sleep(100 * time.Millisecond)
			}
			fmt.Printf("下载%s完成！\n", task.Filename)
		},
		// 接收到终止信号时的停机逻辑回调函数
		func(pool *TaskPool[*DownloadTask]) {
			fmt.Println("接收到终止信号！")
			fmt.Println("当前任务：")
			for _, task := range pool.GetRunningTaskList() {
				fmt.Println(task.Url)
			}
		}, nil)
	// 3.启动任务池
	pool.Start()
}

// 测试无返回值的并发任务池-中断
func TestTaskPool_Interrupt(t *testing.T) {
	// 1.创建任务队列
	list := createTaskListWithError()
	// 2.创建任务池
	pool := NewTaskPool[*DownloadTask](3, 0, 0, list,
		// 每个任务的自定义执行逻辑回调函数
		func(task *DownloadTask, pool *TaskPool[*DownloadTask]) {
			fmt.Printf("正在下载：%s...\n", task.Filename)
			// 模拟出现错误
			if task.Url == "" {
				fmt.Println("出现错误！")
				// 调用pool对象的Interrupt方法直接中断整个任务池
				pool.Interrupt()
				fmt.Println("已结束任务池！")
				return
			}
			// 模拟执行任务
			for i := 0; i < 4; i++ {
				task.Process += 25
				time.Sleep(100 * time.Millisecond)
			}
			fmt.Printf("下载%s完成！\n", task.Filename)
		}, nil, nil)
	// 3.启动任务池
	pool.Start()
}

// 测试并发任务池，观测正在执行任务的实时状态
func TestTaskPool_LookupTasks(t *testing.T) {
	// 1.创建任务列表
	list := createTaskList()
	// 2.创建任务池
	pool := NewTaskPool[*DownloadTask](3, 0, 0, list,
		// 每个任务的自定义执行逻辑回调函数
		func(task *DownloadTask, pool *TaskPool[*DownloadTask]) {
			// 模拟执行任务
			for i := 0; i < 10; i++ {
				task.Process += 10
				time.Sleep(100 * time.Millisecond)
			}
		},
		// 接收到终止信号时的停机逻辑回调函数
		func(pool *TaskPool[*DownloadTask]) {
			fmt.Println("接收到终止信号！")
			fmt.Println("当前任务：")
			for _, task := range pool.GetRunningTaskList() {
				fmt.Println(task.Url)
			}
		},
		// 任务池运行时实时查看并输出任务池任务状态的自定义逻辑回调函数
		func(pool *TaskPool[*DownloadTask]) {
			// 每次输出时清屏（实现实时输出效果）
			fmt.Print("\033[H\033[J")
			// 获取当前正在执行的任务
			tasks := pool.GetRunningTaskList()
			// 遍历获取全部任务状态
			for _, task := range tasks {
				fmt.Printf("正在下载：%s，进度：%d%%\n", task.Filename, task.Process)
			}
			time.Sleep(50 * time.Millisecond)
		})
	// 3.启动任务池
	pool.Start()
}