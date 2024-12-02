package concurrent_task_pool

import (
	"sync"
)

// ArrayQueue 是一个基于切片的顺序结构循环队列的实现
//
// 循环队列通过对数组容量取余的操作，避免假溢出问题
type ArrayQueue[T any] struct {
	// 队列数据
	data []T
	// 队首指针（下标），指向队首元素的位置
	front int
	// 当前队列中元素个数
	size int
	// 锁
	lock sync.RWMutex
}

// NewArrayQueue 顺序队列构造函数
//
// 返回一个空的顺序队列对象指针
func NewArrayQueue[T any]() *ArrayQueue[T] {
	return &ArrayQueue[T]{
		// 初始容量为10
		data:  make([]T, 10),
		front: 0,
		size:  0,
		lock:  sync.RWMutex{},
	}
}

// NewArrayQueueFromSlice 从一个现有切片创建顺序队列
//
// slice 给定切片，切片中下标为0的元素会被放置于队头，最后一个元素放置于队尾
//
// 包含给定切片元素的顺序队列
func NewArrayQueueFromSlice[T any](slice []T) *ArrayQueue[T] {
	queue := &ArrayQueue[T]{
		data:  make([]T, len(slice)),
		front: 0,
		size:  len(slice),
		lock:  sync.RWMutex{},
	}
	copy(queue.data, slice)
	return queue
}

// 复制队列中的全部元素到一个新的切片中并返回该切片副本
//
// targetSize 复制到的新的目标切片的长度，若小于队列大小，则实际长度会调整为队列大小
//
// 包含全部队列元素的切片副本，长度为targetSize，元素顺序：从队头到队尾
func (queue *ArrayQueue[T]) copy(targetSize int) []T {
	// 检查大小
	if queue.size == 0 {
		return []T{}
	}
	if targetSize < queue.size {
		targetSize = queue.size
	}
	// 获取尾指针
	rear := queue.getRear()
	// 创建新切片
	newCopy := make([]T, targetSize)
	// 视情况复制
	if rear-1 >= queue.front {
		copy(newCopy, queue.data[queue.front:rear])
	} else {
		// 原始切片中队头到切片末端
		frontToEnd := queue.data[queue.front:]
		// 原始切片中队尾及其之前
		startToRear := queue.data[:rear]
		// 复制拼接到新的切片
		copy(newCopy, frontToEnd)
		copy(newCopy[len(frontToEnd):], startToRear)
	}
	return newCopy
}

// 队列扩容
func (queue *ArrayQueue[T]) scale() {
	// 扩容两倍并复制新元素
	queue.data = queue.copy(len(queue.data) * 2)
	// 重置指针
	queue.front = 0
}

// 判断队列是否已满，即队列data切片是否已被占满
//
// 返回队列是否已满
func (queue *ArrayQueue[T]) queueFull() bool {
	return queue.size == len(queue.data)
}

// 计算并返回队尾指针位置
//
// 队尾指针（下标），指向队尾元素的后一个位置
func (queue *ArrayQueue[T]) getRear() int {
	return (queue.front + queue.size) % len(queue.data)
}

func (queue *ArrayQueue[T]) Offer(element T) {
	queue.lock.Lock()
	defer queue.lock.Unlock()
	// 如果队列已满，则先进行扩容操作
	if queue.queueFull() {
		queue.scale()
	}
	// 元素放到队尾指针处
	queue.data[queue.getRear()] = element
	queue.size++
}

// Poll 队列头取出一个元素
//
// 返回队列头元素
func (queue *ArrayQueue[T]) Poll() T {
	queue.lock.Lock()
	polledElement := queue.Peek()
	if queue.size != 0 {
		queue.front = (queue.front + 1) % len(queue.data)
		queue.size--
	}
	queue.lock.Unlock()
	return polledElement
}

// Peek 查看队头元素，但是不从队列移除
//
// 返回队列头元素
func (queue *ArrayQueue[T]) Peek() T {
	if queue.size == 0 {
		var zero T
		return zero
	}
	return queue.data[queue.front]
}

// Size 返回队列元素个数
//
// 队列元素个数
func (queue *ArrayQueue[T]) Size() int {
	return queue.size
}

// Clear 清空队列
func (queue *ArrayQueue[T]) Clear() {
	queue.lock.Lock()
	defer queue.lock.Unlock()
	queue.size = 0
	queue.front = 0
}

// IsEmpty 判断队列是否为空
//
// 为空返回true
func (queue *ArrayQueue[T]) IsEmpty() bool {
	return queue.size == 0
}

// ToSlice 队列转换成切片
//
// 返回存放队列全部元素的切片
func (queue *ArrayQueue[T]) ToSlice() []T {
	queue.lock.RLock()
	slice := queue.copy(queue.size)
	queue.lock.RUnlock()
	return slice
}