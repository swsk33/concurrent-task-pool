package concurrent_task_pool

import (
	"sync"
)

// worker 是任务池中的每一个任务运行器
//
// 泛型T表示任务对象参数类型
//
// 一个worker持有一个线程，并一直从任务队列（通道）中获取任务并执行
// 该worker所执行的任务是无返回值的
type worker[T comparable] struct {
	// 自定义任务运行的回调函数
	run func(task T)
	// 存放全部任务的队列通道的引用
	taskQueue chan T
	// 存放当前正在执行的任务的集合引用
	currentTask *mapSet[T]
}

// worker 构造函数
func newWorker[T comparable](run func(T), queue chan T, set *mapSet[T]) *worker[T] {
	return &worker[T]{
		run:         run,
		taskQueue:   queue,
		currentTask: set,
	}
}

// 启动worker，该函数会在一个单独的线程中启动并运行worker
// worker在单独的线程运行时，会一直从任务队列通道中获取任务参数，直到通道关闭，该worker才会结束，确保多个worker使用同一个waitGroup
//
// waitGroup 线程组计数器，每次启动一个worker会将其+1，当worker将全部任务运行结束并关闭时，会将其-1，用于主线程等待
// isShutdown 指示程序是否接收到终止信号的变量指针，当为true时，worker会在执行完当前任务后立即结束
func (worker *worker[T]) start(waitGroup *sync.WaitGroup, isShutdown *bool) {
	// 启动前计数器+1
	waitGroup.Add(1)
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
			worker.run(task)
			// 执行完成后，从当前任务列表移除
			worker.currentTask.remove(task)
		}
	}()
}