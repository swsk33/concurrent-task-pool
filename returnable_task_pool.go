package concurrent_task_pool

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ReturnableTaskPool 并发任务池，用于执行指定数量的并发多任务，其中任务是无返回值的
type ReturnableTaskPool[T, R comparable] struct {
	basePool[T]
	// 执行每个任务的回调函数逻辑
	//
	// 回调函数参数：
	//  - task 从任务队列中取出的一个任务对象，该任务对象可在该函数中被处理并进一步执行任务，该函数调用在一个单独的线程中运行
	//  - taskPool 并发任务池本身，可在每个任务执行时通过该任务池访问任务池中的队列或者中断任务池等
	//
	// 返回值：任务执行完成后的返回结果
	run func(task T, taskPool *ReturnableTaskPool[T, R]) R
	// 接收到终止信号后的操作
	//
	// 参数为当前并发任务池对象，可从其中获取任务状态并执行保存
	shutdown func(taskPool *ReturnableTaskPool[T, R])
	// 任务池执行时，可用于实时查看任务池状态的自定义回调函数，可以指定为nil
	// 该回调函数会在任务池执行任务时被不间断调用
	// 任务池全部任务执行完成后，该回调函数不会再被调用
	//
	// 参数为当前并发任务池对象，可从中实时读取任务池状态
	lookup func(pool *ReturnableTaskPool[T, R])
}

// NewReturnableTaskPool 通过现有的任务列表创建任务池
//
//   - concurrent 任务并发数，即worker数量，每一个worker负责在一个单独的线程中运行任务，当队列中任务数量足够时，并发任务池会一直保持有concurrent个任务一直在并发运行
//   - createInterval 创建worker时的时间间隔
//     若设为0则会在开启并发任务池时同时创建完成全部worker
//     该参数不影响任务池执行时worker从队列取出任务的速度，仅仅代表任务池初始化时创建worker的间隔
//   - executeDelay worker执行每个任务之前的延迟
//     若设为0则所有worker每次从任务队列取出任务后就立即执行
//     否则，当worker每次从任务队列取出任务时，会延迟一段时间再执行任务
//   - taskList 存放全部任务的切片
//   - runFunction 自定义执行任务逻辑的回调函数，其参数为：
//     task 从任务队列中取出的一个任务对象，该任务对象可在该函数中被处理并进一步执行任务，该函数调用在一个单独的线程中运行
//     taskPool 并发任务池本身，可通过任务池对象进行重试操作或者中断等
//     返回值：每个任务执行完成后的返回结果
//   - shutdownFunction 接收到终止信号后的自定义停机逻辑回调函数，可以指定为nil，其参数为：
//     taskPool 并发任务池本身，可在每个任务执行时通过该任务池访问任务池中的队列或者中断任务池等
//   - lookup 任务池执行时，可用于实时查看任务池状态的自定义回调函数，可以指定为nil
//     该回调函数会在任务池执行任务时被不间断调用
//     任务池全部任务执行完成后，该回调函数不会再被调用
//     其参数为：
//     taskPool 当前并发任务池对象，可从中实时读取任务池状态
//
// 返回一个新建的有返回值的并发任务池对象指针
func NewReturnableTaskPool[T, R comparable](concurrent int, createInterval, executeDelay time.Duration, taskList []T, runFunction func(task T, taskPool *ReturnableTaskPool[T, R]) R, shutdownFunction func(taskPool *ReturnableTaskPool[T, R]), lookupFunction func(taskPool *ReturnableTaskPool[T, R])) *ReturnableTaskPool[T, R] {
	return &ReturnableTaskPool[T, R]{
		basePool: basePool[T]{
			concurrent:         concurrent,
			taskCreateInterval: createInterval,
			workerExecuteDelay: executeDelay,
			taskQueue:          newArrayQueueFromSlice(taskList),
			runningTasks:       newMapSet[T](),
			isInterrupt:        false,
			isAutoSaving:       false,
		},
		run:      runFunction,
		shutdown: shutdownFunction,
		lookup:   lookupFunction,
	}
}

// NewSimpleReturnableTaskPool 创建一个有返回值的并发任务池，使用最简单的参数组合
// 其中：
//   - worker创建时间间隔为0
//   - worker执行任务延迟为0
//   - 没有自定义停机逻辑和自定义lookup逻辑
//
// 参数：
//   - concurrent 任务并发数，即worker数量，每一个worker负责在一个单独的线程中运行任务，当队列中任务数量足够时，并发任务池会一直保持有concurrent个任务一直在并发运行
//   - taskList 存放全部任务的切片
//   - runFunction 自定义执行任务逻辑的回调函数，其参数为：
//     task 从任务队列中取出的一个任务对象，该任务对象可在该函数中被处理并进一步执行任务，该函数调用在一个单独的线程中运行
//     taskPool 并发任务池本身，可通过任务池对象进行重试操作或者中断等
//     返回值：每个任务执行完成后的返回结果
func NewSimpleReturnableTaskPool[T, R comparable](concurrent int, taskList []T, runFunction func(task T, taskPool *ReturnableTaskPool[T, R]) R) *ReturnableTaskPool[T, R] {
	return NewReturnableTaskPool[T, R](concurrent, 0, 0, taskList, runFunction, nil, nil)
}

// NewNoDelayReturnableTaskPool 创建一个并发任务池，无任何延迟
// 其中：
//   - worker创建时间间隔为0
//   - worker执行任务延迟为0
//
// 参数：
//   - concurrent 任务并发数，即worker数量，每一个worker负责在一个单独的线程中运行任务，当队列中任务数量足够时，并发任务池会一直保持有concurrent个任务一直在并发运行
//   - taskList 存放全部任务的切片
//   - runFunction 自定义执行任务逻辑的回调函数，其参数为：
//     task 从任务队列中取出的一个任务对象，该任务对象可在该函数中被处理并进一步执行任务，该函数调用在一个单独的线程中运行
//     taskPool 并发任务池本身，可通过任务池对象进行重试操作或者中断等
//   - shutdownFunction 接收到终止信号后的自定义停机逻辑回调函数，可以指定为nil，其参数为：
//     taskPool 并发任务池本身，可在每个任务执行时通过该任务池访问任务池中的队列或者中断任务池等
//   - lookupFunction 任务池执行时，可用于实时查看任务池状态的自定义回调函数，可以指定为nil，
//     该回调函数会在任务池执行任务时被不间断调用
//     任务池全部任务执行完成后，该回调函数不会再被调用
//     其参数为：
//     taskPool 并发任务池本身，可从中实时读取任务池状态
func NewNoDelayReturnableTaskPool[T, R comparable](concurrent int, taskList []T, runFunction func(task T, taskPool *ReturnableTaskPool[T, R]) R, shutdownFunction func(taskPool *ReturnableTaskPool[T, R]), lookupFunction func(taskPool *ReturnableTaskPool[T, R])) *ReturnableTaskPool[T, R] {
	return NewReturnableTaskPool[T, R](concurrent, 0, 0, taskList, runFunction, shutdownFunction, lookupFunction)
}

// Start 启动并发任务池
//
//   - ignoreEmpty 是否收集空的任务执行返回值
//
// 返回全部任务执行后的返回值列表
func (pool *ReturnableTaskPool[T, R]) Start(ignoreEmpty bool) []R {
	// 结果收集锁
	lock := &sync.Mutex{}
	// 用于控制worker运行的变量，当为false时全部worker将一直等待从任务取出任务执行，否则都会立即停止运行
	workerShutdown := false
	// 在一个新的线程接收终止信号
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		// 标记程序执行完成，结束全部worker
		workerShutdown = true
		// 执行shutdown回调
		if pool.shutdown != nil {
			pool.shutdown(pool)
		}
		// 标记为中断
		pool.isInterrupt = true
	}()
	// 创建结果列表切片
	resultList := make([]R, 0)
	// 创建worker
	for i := 0; i < pool.concurrent; i++ {
		eachWorker := newReturnableWorker[T, R](pool.run, &resultList, pool)
		eachWorker.start(lock, &workerShutdown, ignoreEmpty)
		if pool.taskCreateInterval > 0 {
			time.Sleep(pool.taskCreateInterval)
		}
	}
	// 等待直到队列中无任务，且任务列表中也没有任务了，说明全部任务完成
	// 若被标记为中断，则会立即结束
	for !pool.isInterrupt && !pool.IsAllDone() {
		// 执行lookup函数
		if pool.lookup != nil {
			pool.lookup(pool)
		}
	}
	// 结束全部worker
	workerShutdown = true
	return resultList
}