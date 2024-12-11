package concurrent_task_pool

import (
	"sync"
)

// returnableWorker 是任务池中的每一个任务运行器
//
// 泛型T表示任务对象参数类型
// 泛型R表示任务执行后的返回值类型
//
// 一个worker持有一个线程，并一直从任务队列（通道）中获取任务并执行
// 该worker所执行的任务是有返回值的
type returnableWorker[T, R comparable] struct {
	// 自定义任务运行的回调函数
	run func(task T, pool *ReturnableTaskPool[T, R]) R
	// 收集存放任务结果的切片引用
	resultList *[]R
	// 该worker所属的并发任务池对象的引用
	taskPool *ReturnableTaskPool[T, R]
}

// returnableWorker 构造函数
func newReturnableWorker[T, R comparable](run func(T, *ReturnableTaskPool[T, R]) R, result *[]R, pool *ReturnableTaskPool[T, R]) *returnableWorker[T, R] {
	return &returnableWorker[T, R]{
		run:        run,
		resultList: result,
		taskPool:   pool,
	}
}

// 启动worker，该函数会在一个单独的线程中启动并运行worker
// worker在单独的线程运行，会一直从任务队列中获取任务对象，直到isShutdown为true才结束
//
//   - lock 用于收集结果的锁，确保多个worker使用同一个lock
//   - isShutdown 指示全部任务是否结束的指针，当为true时，worker会在执行完当前任务后立即结束
//   - ignoreEmpty 是否收集空的任务执行返回值
func (worker *returnableWorker[T, R]) start(lock *sync.Mutex, isShutdown *bool, ignoreEmpty bool) {
	// 当前任务池
	pool := worker.taskPool
	// 泛型零值
	var taskZero T
	var resultZero R
	// 在新的线程中运行任务
	go func() {
		// 除非isShutdown为true，否则将会一直尝试从队列取值
		for !*isShutdown {
			// 从队列取值
			task := pool.taskQueue.poll()
			if task == taskZero {
				continue
			}
			// 将当前任务存入当前正在运行的任务集合中
			pool.runningTasks.add(task)
			// 执行任务
			result := worker.run(task, worker.taskPool)
			// 收集结果
			if result != resultZero || (result == resultZero && !ignoreEmpty) {
				lock.Lock()
				*worker.resultList = append(*worker.resultList, result)
				lock.Unlock()
			}
			// 执行完成后，从当前任务列表移除
			pool.runningTasks.remove(task)
		}
	}()
}