# 并发任务池-Go

一个简单的Golang并发任务池，是对工作池模式的一个基本实现。

## 1，简介

在一些情况下，我们需要批量执行任务，通常我们会将任务存放至队列中，并依次从任务队列取出任务并执行。

使用多线程并发执行任务是一个提高效率的办法，假设我们的任务队列有`m`个任务，每次批量执行`n`个任务，通常`m`远大于`n`，那么大概的执行逻辑是：

- 开始时，启动`n`个线程，并取出`n`个任务执行
- 一旦有任务执行完成，就立即从队列取出任务给空闲队列执行，保证一直有`n`个任务在执行（除非队列里面没有任务了）
- 等待队列为空，且全部线程都执行完成，说明全部任务执行完成，收集结果

在这个过程中，我们需要编写管理并发任务的逻辑，包括但不限于根据当前并发的任务数决定是否取出任务、结果收集的线程安全问题、并发计数器的维护等等，这是一个比较麻烦的过程。

该并发任务池提供了一个基本的工作池模式的实现，对上述批处理执行任务场景中任务管理的逻辑进行了封装，仅需传入任务列表、任务执行逻辑以及一些参数，即可运行一个并发任务池，实现大量任务的多线程批处理工作。

该并发任务池主要功能如下：

- 支持无返回值和有返回值的任务池创建
- 自定义并发数
- 自定义每个任务的执行逻辑
- 自定义程序接收到终止信号（例如`Ctrl + C`时）的自定义停机逻辑

## 2，使用方法

### (1) 安装依赖

通过下列命令：

```bash
go get gitee.com/swsk33/concurrent-task-pool
```

### (2) 创建任务对象结构体

通常，我们需要对我们的任务进行抽象和建模，假设现在有一个批量下载文件的场景，我们将每个下载文件任务表示为如下：

```go
// DownloadTask 一个示例下载任务（参数）表示
type DownloadTask struct {
	url      string
	filename string
}
```

可见下载文件任务对象包含了下载文件任务的必要信息或者参数，例如下载地址与保存的文件名等。

### (3) 创建任务池并运行

现在，只需在准备好任务列表后，创建任务池并启动即可：

```go
package main

import (
	"fmt"
	"gitee.com/swsk33/concurrent-task-pool"
	"time"
)

// 省略DownloadTask结构体声明...

// 创建示例任务对象列表
func createTaskList() []*DownloadTask {
	list := make([]*DownloadTask, 0)
	for i := 1; i <= 30; i++ {
		list = append(list, &DownloadTask{
			url:      fmt.Sprintf("http://example.com/file/%d.txt", i),
			filename: fmt.Sprintf("file-%d.txt", i),
		})
	}
	return list
}

func main() {
	// 1.创建任务列表
	list := createTaskList()
	// 2.创建任务池
	pool := concurrent_task_pool.NewTaskPool[*DownloadTask](3, list, func(task *DownloadTask, pool *concurrent_task_pool.TaskPool[*DownloadTask]) {
		fmt.Printf("正在下载：%s...\n", task.filename)
		// 模拟执行任务
		time.Sleep(350 * time.Millisecond)
		fmt.Printf("下载%s完成！\n", task.filename)
	}, func(tasks []*DownloadTask) {
		fmt.Println("接收到终止信号！")
		fmt.Println("当前任务：")
		for _, task := range tasks {
			fmt.Println(task.url)
		}
	})
	// 3.启动任务池
	pool.Start()
}
```

可见主要分为下列几个简单步骤：

- **准备任务列表**：创建自定义的任务对象列表，并组织为切片形式
- **创建任务池**：通过`NewTaskPool`构造函数创建任务池对象`TaskPool`，其中：
	- 泛型`T`：表示**自定义任务对象的类型**，若任务对象为结构体建议使用指针形式
	- 参数`1`：**并发数**，表示任务池中同时有多少个任务并发执行
	- 参数`2`：**任务列表**，传入我们自定义的任务对象切片
	- 参数`3`：**任务执行逻辑**，为一个回调函数，用于自定义每个任务的执行逻辑，该回调函数有下列参数：
		- 参数`1`：**每次执行任务时从队列取出的那个任务对象**，可在回调函数中通过对该任务对象进行处理，实现自定义的任务执行逻辑，该函数调用会在一个单独的Goroutine中异步执行
		- 参数`2`：**并发任务池对象本身**，可在每个任务执行时通过该任务池访问任务池中的队列或者中断任务池等
	- 参数`4`：**停机逻辑**，为一个回调函数，用于自定义接收到终止信号（例如`Ctrl + C`）时执行的逻辑，其参数表示**接收到终止信号时正在执行的任务列表**，可将其持久化
- **启动任务池**：调用任务池对象`Start`方法启动即可，此时开始并发地执行任务，该方法会阻塞当前线程直到任务队列中的任务全部执行完成

### (4) 实现失败重试

在`TaskPool`对象中，都使用了一个线程安全的队列来实现存放未完成任务，这样的话如果有的任务执行失败了，我们就可以将其放回队列，实现后续重试该任务。

以无返回值的并发任务池为例，实现在任务执行失败时，将任务放回队列以后续重试的操作：

```go
package main

import (
	"fmt"
	"gitee.com/swsk33/concurrent-task-pool"
	"time"
)

// 省略DownloadTask结构体声明...

// 创建带有错误的任务对象列表
func createTaskListWithError() []*DownloadTask {
	list := make([]*DownloadTask, 0)
	for i := 1; i <= 10; i++ {
		// 模拟第3个任务有错误
		if i == 3 {
			list = append(list, &DownloadTask{
				url:      "",
				filename: fmt.Sprintf("file-%d.txt", i),
			})
			continue
		}
		list = append(list, &DownloadTask{
			url:      fmt.Sprintf("http://example.com/file/%d.txt", i),
			filename: fmt.Sprintf("file-%d.txt", i),
		})
	}
	return list
}

func main() {
	// 1.创建任务队列
	list := createTaskListWithError()
	// 2.创建任务池
	pool := concurrent_task_pool.NewTaskPool[*DownloadTask](3, list, func(task *DownloadTask, pool *concurrent_task_pool.TaskPool[*DownloadTask]) {
		fmt.Printf("正在下载：%s...\n", task.filename)
		// 模拟出现错误
		if task.url == "" {
			fmt.Println("出现错误！")
			task.url = fmt.Sprintf("http://example.com/file/%s", task.filename)
			// 将任务重新放回并发任务池的任务队列
			// pool的TaskQueue属性即为任务池中存放全部任务的队列
			// Offer方法可将对象入队列
			pool.TaskQueue.Offer(task)
			return
		}
		// 模拟执行任务
		time.Sleep(350 * time.Millisecond)
		fmt.Printf("下载%s完成！\n", task.filename)
	}, func(tasks []*DownloadTask) {
		fmt.Println("接收到终止信号！")
		fmt.Println("当前任务：")
		for _, task := range tasks {
			fmt.Println(task.url)
		}
	})
	// 3.启动任务池
	pool.Start()
}
```

可见这里有以下的不同：

- 首先仍然是通过任务对象列表切片创建了并发任务池对象
- 在自定义的任务执行回调函数中，当遇到任务失败的情况时，就可以调用回调函数参数的`pool`对象（并发任务池本身）的`TaskQueue`属性，将当前任务对象放回队列，实现后续重试该任务

`TaskPool`对象的`TaskQueue`属性即为**任务池中存放全部任务的队列**，该队列为内部封装的`ArrayQueue`指针类型，有下列方法：

- `Poll()` 队头元素出队列
- `Offer(e)` 元素进入队列
- `Peek()` 查看队头元素，但是不移除
- `IsEmpty()` 返回队列是否为空
- `Size()` 返回队列元素个数
- `Clear()` 清空队列
- `ToSlice()` 将队列以切片形式返回

### (5) 任务池中断

`TaskPool`对象提供了`Interrupt`方法，能够实现在某个任务出现关键错误或者在其它特殊情况下立即中断并发任务池的执行：

```go
package main

import (
	"fmt"
	"gitee.com/swsk33/concurrent-task-pool"
	"time"
)

// 省略DownloadTask结构体声明...

// 创建带有错误的任务对象列表
func createTaskListWithError() []*DownloadTask {
	list := make([]*DownloadTask, 0)
	for i := 1; i <= 10; i++ {
		// 模拟第3个任务有错误
		if i == 3 {
			list = append(list, &DownloadTask{
				url:      "",
				filename: fmt.Sprintf("file-%d.txt", i),
			})
			continue
		}
		list = append(list, &DownloadTask{
			url:      fmt.Sprintf("http://example.com/file/%d.txt", i),
			filename: fmt.Sprintf("file-%d.txt", i),
		})
	}
	return list
}

func main() {
	// 1.创建任务队列
	list := createTaskListWithError()
	// 2.创建任务池
	pool := concurrent_task_pool.NewTaskPool[*DownloadTask](3, list, func(task *DownloadTask, pool *concurrent_task_pool.TaskPool[*DownloadTask]) {
		fmt.Printf("正在下载：%s...\n", task.filename)
		// 模拟出现错误
		if task.url == "" {
			fmt.Println("出现错误！")
			// 调用pool对象的Interrupt方法直接中断整个任务池
			pool.Interrupt()
			fmt.Println("已结束任务池！")
			return
		}
		// 模拟执行任务
		time.Sleep(350 * time.Millisecond)
		fmt.Printf("下载%s完成！\n", task.filename)
	}, func(tasks []*DownloadTask) {
		fmt.Println("接收到终止信号！")
		fmt.Println("当前任务：")
		for _, task := range tasks {
			fmt.Println(task.url)
		}
	})
	// 3.启动任务池
	pool.Start()
}
```

上述示例中，当某个任务遇到错误时，直接调用了任务池对象的`Interrupt`方法进行了中断操作，此时任务池将不会继续执行任务，且无法恢复。

### (6) 有返回值的任务池

上述使用的任务池`TaskPool`中的异步任务是没有返回值的，可使用`ReturnableTaskPool`来执行有返回值的异步任务：

```go
package main

import (
	"fmt"
	"gitee.com/swsk33/concurrent-task-pool"
	"time"
)

// 省略DownloadTask结构体声明...

// 创建示例任务对象列表
func createTaskList() []*DownloadTask {
	list := make([]*DownloadTask, 0)
	for i := 1; i <= 30; i++ {
		list = append(list, &DownloadTask{
			url:      fmt.Sprintf("http://example.com/file/%d.txt", i),
			filename: fmt.Sprintf("file-%d.txt", i),
		})
	}
	return list
}

func main() {
	// 1.创建任务列表
	list := createTaskList()
	// 2.创建任务池（有返回值的）
	pool := concurrent_task_pool.NewReturnableTaskPool[*DownloadTask, string](3, list, func(task *DownloadTask, pool *concurrent_task_pool.ReturnableTaskPool[*DownloadTask, string]) string {
		fmt.Printf("正在下载：%s...\n", task.filename)
		// 模拟执行任务
		time.Sleep(350 * time.Millisecond)
		// 返回文件名
		return task.filename
	}, func(tasks []*DownloadTask) {
		fmt.Println("接收到终止信号！")
		fmt.Println("当前任务：")
		for _, task := range tasks {
			fmt.Println(task.url)
		}
	})
	// 3.启动任务池
	resultList := pool.Start(true)
	// 4.执行完成，读取结果
	fmt.Println("执行完成！全部结果：")
	for _, result := range resultList {
		fmt.Println(result)
	}
}
```

借助`NewReturnableTaskPool`构造函数可以创建一个有返回值的并发任务池，其参数和`NewTaskPool`构造函数几乎一样，只不过：

- 多了一个泛型`R`表示**任务返回的类型**
- 参数`3`自定义任务执行逻辑回调函数有返回值，需要在这个回调函数中**返回任务执行完成后的结果**

此外，`Start`方法启动任务池时，需要传入一个`bool`类型参数表示**是否忽略空的任务结果**，如果该参数为`true`，那么当一个任务返回的结果为`nil`或者对应类型零值时，这个结果就不会被包含在最终的结果中。此外，这里的`Start`的返回值就是全部任务执行后收集的全部返回结果的切片。

可以使用和`TaskPool`同样的方式，在有返回值的并发任务池中实现失败重试或者中断操作。
