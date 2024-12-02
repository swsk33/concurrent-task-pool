package concurrent_task_pool

import "sync"

// returnableWorker 是任务池中的每一个任务运行器
//
// 泛型T表示任务对象参数类型
// 泛型R表示任务执行后的返回值类型
//
// 一个worker持有一个线程，并一直从任务队列（通道）中获取任务并执行
// 该worker所执行的任务是有返回值的
type returnableWorker[T, R comparable] struct {
	// 自定义任务运行的回调函数
	run func(task T) R
	// 存放全部任务的队列通道的引用
	taskQueue chan T
	// 存放当前正在执行的任务的集合引用
	currentTask *mapSet[T]
	// 收集存放任务结果的切片引用
	resultList *[]R
}

// returnableWorker 构造函数
func newReturnableWorker[T, R comparable](run func(T) R, queue chan T, set *mapSet[T], result *[]R) *returnableWorker[T, R] {
	return &returnableWorker[T, R]{
		run:         run,
		taskQueue:   queue,
		currentTask: set,
		resultList:  result,
	}
}

// 启动worker，该函数会在一个单独的线程中启动并运行worker
// worker在单独的线程运行时，会一直从任务队列通道中获取任务参数，直到通道关闭，该worker才会结束
//
// waitGroup 线程组计数器，每次启动一个worker会将其+1，当worker将全部任务运行结束并关闭时，会将其-1，用于主线程等待，确保多个worker使用同一个waitGroup
// lock 用于收集结果的锁，确保多个worker使用同一个lock
// isShutdown 指示程序是否接收到终止信号的变量指针，当为true时，worker会在执行完当前任务后立即结束
// ignoreEmpty 是否收集空的任务执行返回值
func (worker *returnableWorker[T, R]) start(waitGroup *sync.WaitGroup, lock *sync.Mutex, isShutdown *bool, ignoreEmpty bool) {
	// 启动前计数器+1
	waitGroup.Add(1)
	// 零值
	var zero R
	// 在新的线程中运行任务
	go func() {
		// 结束后计数器-1
		defer waitGroup.Done()
		// 一直等待并取出队列任务
		for task := range worker.taskQueue {
			if *isShutdown {
				return
			}
			// 将当前任务存入当前正在运行的任务集合中
			worker.currentTask.add(task)
			// 执行任务
			result := worker.run(task)
			// 收集结果
			if result != zero || (result == zero && !ignoreEmpty) {
				lock.Lock()
				*worker.resultList = append(*worker.resultList, result)
				lock.Unlock()
			}
			// 执行完成后，从当前任务列表移除
			worker.currentTask.remove(task)
		}
	}()
}