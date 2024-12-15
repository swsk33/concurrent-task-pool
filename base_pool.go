package concurrent_task_pool

import (
	"encoding/json"
	"fmt"
	"time"
)

// 并发任务池的基本类型，包含了一个并发任务池中的全部任务队列、正在运行的任务以及一些状态等等
type basePool[T comparable] struct {
	// 任务并发数，即worker数量，每一个worker负责在一个单独的线程中运行任务
	// 当队列中任务数量足够时，并发任务池会一直保持有concurrent个任务一直在并发运行
	concurrent int
	// 创建worker时的时间间隔
	// 若设为0则会在开启并发任务池时同时创建完成全部worker
	// 该属性不影响worker从队列取出任务的速度，仅仅代表任务池初始化时创建worker的间隔
	taskCreateInterval time.Duration
	// worker执行每个任务之前的延迟
	// 若设为0则所有worker每次从任务队列取出任务后就立即执行
	// 否则，当worker每次从任务队列取出任务时，会延迟一段时间再执行任务
	workerExecuteDelay time.Duration
	// 存放全部任务的队列
	taskQueue *arrayQueue[T]
	// 当前正在执行的全部任务集合
	runningTasks *mapSet[T]
	// 是否被中断
	// 当该变量为true时，则会立即停止并发任务池的任务
	isInterrupt bool
	// 用于自动任务保存的计时器
	saveTicker *time.Ticker
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

// GetQueuedTaskList 获取并发任务池中的全部位于任务队列中的任务列表
//
// 返回当前并发任务池中，位于任务队列中的全部任务（还在排队且未执行的任务）
func (pool *basePool[T]) GetQueuedTaskList() []T {
	return pool.taskQueue.toSlice()
}

// GetRunningTaskList 获取并发任务池中正在执行的任务列表
//
// 返回当前并发任务池全部正在执行的任务
func (pool *basePool[T]) GetRunningTaskList() []T {
	return pool.runningTasks.toSlice()
}

// GetAllTaskList 获取全部任务，即：任务队列中正在排队的任务 + 正在执行的任务
//
// 返回任务池中全部任务
func (pool *basePool[T]) GetAllTaskList() []T {
	// 使用集合去重
	// 全部排队任务
	taskSet := newMapSetFromSlice(pool.GetQueuedTaskList())
	// 加入正在执行的任务
	runningTasks := pool.GetRunningTaskList()
	for _, task := range runningTasks {
		taskSet.add(task)
	}
	return taskSet.toSlice()
}

// Retry 重试任务，若任务执行失败，可将当前任务对象重新放回并发任务池的任务队列中，使其在后续重新执行
//
// task 要放回任务队列进行重试的任务
func (pool *basePool[T]) Retry(task T) {
	pool.taskQueue.offer(task)
}

// SaveTaskList 将并发任务池中的全部任务（包括队列任务和正在执行的任务）序列化并保存至本地
// 需要将任务对象的必要字段导出，并使用json标签才能够保存
//
//   - file 任务文件保存位置
//
// 若出现错误，则返回错误对象
func (pool *basePool[T]) SaveTaskList(file string) error {
	// 序列化为JSON
	taskJson, e := json.Marshal(pool.GetAllTaskList())
	if e != nil {
		return e
	}
	// 保存
	return saveDataToFile(taskJson, file)
}

// EnableTaskAutoSave 启用自动任务保存
// 调用该方法后，每隔指定的时间，就会调用 SaveTaskList 方法一次保存任务
//
//   - file 任务文件保存位置
//   - interval 自动保存间隔
func (pool *basePool[T]) EnableTaskAutoSave(file string, interval time.Duration) {
	pool.saveTicker = time.NewTicker(interval)
	go func() {
		for range pool.saveTicker.C {
			e := pool.SaveTaskList(file)
			if e != nil {
				fmt.Printf("保存任务出现错误：%s\n", e)
			}
		}
	}()
}

// DisableTaskAutoSave 关闭自动任务保存
// 在使用 EnableTaskAutoSave 后，若后续不再需要自动保存任务，则可以调用该函数关闭自动保存
// 此外，任务池全部任务执行完成后或者被中断时，会自动关闭自动任务保存
func (pool *basePool[T]) DisableTaskAutoSave() {
	if pool.saveTicker != nil {
		pool.saveTicker.Stop()
	}
}