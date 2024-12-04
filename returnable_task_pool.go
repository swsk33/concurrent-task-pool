package concurrent_task_pool

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// ReturnableTaskPool 并发任务池，用于执行指定数量的并发多任务，其中任务是无返回值的
type ReturnableTaskPool[T, R comparable] struct {
	// 任务并发数，即worker数量，当队列中任务数量足够时，并发任务池会一直保持有concurrent个任务一直在并发运行
	Concurrent int
	// 存放全部任务的队列
	TaskQueue *ArrayQueue[T]
	// 执行每个任务的回调函数逻辑
	//
	// 回调函数参数：
	//
	// task 从任务队列中取出的一个任务对象，该任务对象可在该函数中被处理并进一步执行任务，该函数调用在一个单独的线程中运行
	//
	// taskPool 并发任务池本身，可在每个任务执行时通过该任务池访问任务池中的队列或者中断任务池等
	//
	// 返回值：任务执行完成后的返回结果
	Run func(task T, taskPool *ReturnableTaskPool[T, R]) R
	// 接收到终止信号后的操作
	//
	// 参数为正在运行的任务列表，可将其保存起来
	Shutdown func(tasks []T)
	// 是否被中断
	//
	// 当该变量为true时，则会立即停止并发任务池的任务
	isInterrupt bool
}

// NewReturnableTaskPool 通过现有的任务列表创建任务池
func NewReturnableTaskPool[T, R comparable](concurrent int, taskList []T, runFunction func(T, *ReturnableTaskPool[T, R]) R, shutdownFunction func([]T)) *ReturnableTaskPool[T, R] {
	return &ReturnableTaskPool[T, R]{
		Concurrent:  concurrent,
		TaskQueue:   NewArrayQueueFromSlice(taskList),
		Run:         runFunction,
		Shutdown:    shutdownFunction,
		isInterrupt: false,
	}
}

// Start 启动并发任务池
//
// ignoreEmpty 是否收集空的任务执行返回值
//
// 返回全部任务执行后的返回值列表
func (pool *ReturnableTaskPool[T, R]) Start(ignoreEmpty bool) []R {
	// 存放当前正在运行的全部任务集合
	runningTasks := newMapSet[T]()
	// 结果收集锁
	lock := &sync.Mutex{}
	// 任务是否全部完成
	isAllDone := false
	// 在一个新的线程接收终止信号
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		// 标记程序执行完成，结束全部worker
		isAllDone = true
		// 执行Shutdown回调
		pool.Shutdown(runningTasks.toSlice())
		// 标记为中断
		pool.isInterrupt = true
	}()
	// 创建结果列表切片
	resultList := make([]R, 0)
	// 创建worker
	for i := 0; i < pool.Concurrent; i++ {
		eachWorker := newReturnableWorker[T, R](pool.Run, pool.TaskQueue, runningTasks, &resultList, pool)
		eachWorker.start(lock, &isAllDone, ignoreEmpty)
	}
	// 等待直到队列中无任务，且任务列表中也没有任务了，说明全部任务完成
	// 若被标记为中断，则会立即结束
	for !pool.isInterrupt && (!pool.TaskQueue.IsEmpty() || runningTasks.size() != 0) {
		// 阻塞当前线程
	}
	// 结束全部worker
	isAllDone = true
	return resultList
}

// Interrupt 中断任务池，立即停止任务池中正在执行的任务
func (pool *ReturnableTaskPool[T, R]) Interrupt() {
	pool.isInterrupt = true
}