package concurrent_task_pool

import "time"

// 并发任务池的基本类型，包含了一个并发任务池中的全部任务队列、正在运行的任务以及一些状态等等
type basePool[T comparable] struct {
	// 任务并发数，即worker数量，每一个worker负责在一个单独的线程中运行任务
	// 当队列中任务数量足够时，并发任务池会一直保持有concurrent个任务一直在并发运行
	concurrent int
	// 创建worker时的时间间隔
	// 若设为0则会在开启并发任务池时同时创建完成全部worker
	// 该属性不影响worker从队列取出任务的速度，仅仅代表任务池初始化时创建worker的间隔
	taskCreateInterval time.Duration
	// 存放全部任务的队列
	taskQueue *arrayQueue[T]
	// 当前正在执行的全部任务集合
	runningTasks *mapSet[T]
	// 是否被中断
	// 当该变量为true时，则会立即停止并发任务池的任务
	isInterrupt bool
}

// IsAllDone 返回该并发任务池是否完成了全部任务
// 任务队列中无任务，且正在执行的任务集合中也没有任务了，说明全部任务完成
//
// 当并发任务池全部任务执行完成时，返回true
func (pool *basePool[T]) IsAllDone() bool {
	return pool.taskQueue.isEmpty() && pool.runningTasks.size() == 0
}

// Interrupt 中断任务池，立即停止任务池中正在执行的任务
func (pool *basePool[T]) Interrupt() {
	pool.isInterrupt = true
}

// GetTaskList 获取并发任务池中的全部位于任务队列中的任务列表
//
// 返回当前并发任务池中，位于任务队列中的全部任务（还在排队且未执行的任务）
func (pool *basePool[T]) GetTaskList() []T {
	return pool.taskQueue.toSlice()
}

// GetRunningTaskList 获取并发任务池中正在执行的任务列表
//
// 返回当前并发任务池全部正在执行的任务
func (pool *basePool[T]) GetRunningTaskList() []T {
	return pool.runningTasks.toSlice()
}

// Retry 重试任务，若任务执行失败，可将当前任务对象重新放回并发任务池的任务队列中，使其在后续重新执行
//
// task 要放回任务队列进行重试的任务
func (pool *basePool[T]) Retry(task T) {
	pool.taskQueue.offer(task)
}