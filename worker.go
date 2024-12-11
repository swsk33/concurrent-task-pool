package concurrent_task_pool

// worker 是任务池中的每一个任务运行器
//
// 泛型T表示任务对象参数类型
//
// 一个worker持有一个线程，并一直从任务队列（通道）中获取任务并执行
// 该worker所执行的任务是无返回值的
type worker[T comparable] struct {
	// 自定义任务运行的回调函数
	run func(task T, taskPool *TaskPool[T])
	// 该worker所属的并发任务池对象的引用
	taskPool *TaskPool[T]
}

// worker 构造函数
func newWorker[T comparable](run func(T, *TaskPool[T]), pool *TaskPool[T]) *worker[T] {
	return &worker[T]{
		run:      run,
		taskPool: pool,
	}
}

// 启动worker，该函数会在一个单独的线程中启动并运行worker
// worker在单独的线程运行，会一直从任务队列中获取任务对象，直到isShutdown为true才结束
//
// isShutdown 指示全部任务是否结束的指针，当为true时，worker会在执行完当前任务后立即结束
func (worker *worker[T]) start(isShutdown *bool) {
	// 当前任务池
	pool := worker.taskPool
	// 泛型零值
	var zero T
	// 在新的线程中运行任务
	go func() {
		// 除非isShutdown为true，否则将会一直尝试从队列取值
		for !*isShutdown {
			// 从队列取值
			task := pool.taskQueue.poll()
			if task == zero {
				continue
			}
			// 将当前任务存入当前正在运行的任务集合中
			pool.runningTasks.add(task)
			// 执行任务
			worker.run(task, worker.taskPool)
			// 执行完成后，从当前任务列表移除
			pool.runningTasks.remove(task)
		}
	}()
}