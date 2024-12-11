package concurrent_task_pool

import "sync"

// 空结构体别名
type void struct{}

// 基于map对象的集合实现
type mapSet[T comparable] struct {
	// 数据部分
	data map[T]*void
	// 锁
	lock sync.RWMutex
}

// 创建一个新的MapSet
//
// 返回空的MapSet
func newMapSet[T comparable]() *mapSet[T] {
	return &mapSet[T]{
		data: make(map[T]*void),
		lock: sync.RWMutex{},
	}
}

// 添加数据到集合
//
// item 要添加的数据
func (set *mapSet[T]) add(item T) {
	set.lock.Lock()
	defer set.lock.Unlock()
	set.data[item] = nil
}

// 从集合移除数据
//
// item 要移除的数据
func (set *mapSet[T]) remove(item T) {
	set.lock.Lock()
	defer set.lock.Unlock()
	delete(set.data, item)
}

// 判断集合是否包含对应数据
//
// item 要判断的数据
//
// 返回集合是否包含
func (set *mapSet[T]) contains(item T) bool {
	set.lock.RLock()
	defer set.lock.RUnlock()
	_, exists := set.data[item]
	return exists
}

// 返回集合大小
//
// 返回集合元素个数
func (set *mapSet[T]) size() int {
	set.lock.RLock()
	defer set.lock.RUnlock()
	return len(set.data)
}

// 清空集合
func (set *mapSet[T]) clear() {
	set.lock.Lock()
	defer set.lock.Unlock()
	for key := range set.data {
		delete(set.data, key)
	}
}

// 遍历集合
//
// lookup 自定义遍历每个元素的回调函数
func (set *mapSet[T]) forEach(lookup func(item T)) {
	set.lock.RLock()
	defer set.lock.RUnlock()
	for item := range set.data {
		lookup(item)
	}
}

// 将集合转换成切片返回
//
// 返回包含了全部集合元素的切片副本
func (set *mapSet[T]) toSlice() []T {
	set.lock.RLock()
	defer set.lock.RUnlock()
	slice := make([]T, 0, len(set.data))
	set.forEach(func(item T) {
		slice = append(slice, item)
	})
	return slice
}