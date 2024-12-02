package concurrent_task_pool

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// TaskPool 并发任务池，用于执行指定数量的并发多任务，其中任务是无返回值的
type TaskPool[T comparable] struct {
	// 任务并发数，即worker数量，当队列中任务数量足够时，并发任务池会一直保持有concurrent个任务一直在并发运行
	Concurrent int
	// 存放全部任务列表
	TaskList []T
	// 执行每个任务的回调函数逻辑
	//
	// 函数参数为从任务队列中取出的一个任务对象，每个函数在一个线程中运行
	Run func(T)
	// 接收到终止信号后的操作
	//
	// 参数为正在运行的任务列表，可将其保存起来
	Shutdown func(tasks []T)
}

// NewTaskPool 通过现有的任务列表创建任务池
func NewTaskPool[T comparable](concurrent int, taskList []T, runFunction func(T), shutdownFunction func([]T)) *TaskPool[T] {
	return &TaskPool[T]{
		Concurrent: concurrent,
		TaskList:   taskList,
		Run:        runFunction,
		Shutdown:   shutdownFunction,
	}
}

// Start 启动并发任务池
func (pool *TaskPool[T]) Start() {
	// 存放当前正在运行的全部任务集合
	runningTasks := newMapSet[T]()
	// 创建通道作为任务队列
	taskQueue := make(chan T, len(pool.TaskList))
	// 计数器
	waitGroup := &sync.WaitGroup{}
	// 在一个新的线程接收终止信号
	isShutdown := false
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		// 标记程序被退出
		isShutdown = true
		// 执行Shutdown回调
		pool.Shutdown(runningTasks.toSlice())
	}()
	// 创建worker
	for i := 0; i < pool.Concurrent; i++ {
		eachWorker := newWorker[T](pool.Run, taskQueue, runningTasks)
		eachWorker.start(waitGroup, &isShutdown)
	}
	// 向队列通道发送任务
	for _, task := range pool.TaskList {
		taskQueue <- task
	}
	// 关闭通道
	close(taskQueue)
	// 等待全部worker执行完成
	waitGroup.Wait()
}