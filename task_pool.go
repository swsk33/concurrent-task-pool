package concurrent_task_pool

import (
	"os"
	"os/signal"
	"syscall"
)

// TaskPool 并发任务池，用于执行指定数量的并发多任务，其中任务是无返回值的
type TaskPool[T comparable] struct {
	// 任务并发数，即worker数量，当队列中任务数量足够时，并发任务池会一直保持有concurrent个任务一直在并发运行
	Concurrent int
	// 存放全部任务的队列
	TaskQueue *ArrayQueue[T]
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
		TaskQueue:  NewArrayQueueFromSlice(taskList),
		Run:        runFunction,
		Shutdown:   shutdownFunction,
	}
}

// NewTaskPoolUseQueue 通过现有的任务队列创建任务池，任务池中的队列将是传入队列的引用
func NewTaskPoolUseQueue[T comparable](concurrent int, taskQueue *ArrayQueue[T], runFunction func(T), shutdownFunction func([]T)) *TaskPool[T] {
	return &TaskPool[T]{
		Concurrent: concurrent,
		TaskQueue:  taskQueue,
		Run:        runFunction,
		Shutdown:   shutdownFunction,
	}
}

// Start 启动并发任务池
func (pool *TaskPool[T]) Start() {
	// 存放当前正在运行的全部任务集合
	runningTasks := newMapSet[T]()
	// 表示任务是否全部完成了
	isAllDone := false
	// 在一个新的线程接收终止信号
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		// 标记程序被退出
		isAllDone = true
		// 执行Shutdown回调
		pool.Shutdown(runningTasks.toSlice())
	}()
	// 创建worker
	for i := 0; i < pool.Concurrent; i++ {
		eachWorker := newWorker[T](pool.Run, pool.TaskQueue, runningTasks)
		eachWorker.start(&isAllDone)
	}
	// 等待直到队列中无任务，且任务列表中也没有任务了，说明全部任务完成
	for !pool.TaskQueue.IsEmpty() || runningTasks.size() != 0 {
		// 阻塞当前线程
		// 如果接收到终止信号，isAllDone改变，则退出
		if isAllDone {
			return
		}
	}
	// 标记全部完成
	isAllDone = true
}