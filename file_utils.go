package concurrent_task_pool

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
)

// 将数据保存为文件
//
//   - data 要保存的数据
//   - path 保存文件位置，不存在会创建，存在会覆盖
func saveDataToFile(data []byte, path string) error {
	// 创建文件
	file, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if e != nil {
		return e
	}
	defer func() {
		_ = file.Close()
	}()
	// 读取文件
	writer := bufio.NewWriter(file)
	_, e = writer.Write(data)
	if e != nil {
		return e
	}
	return writer.Flush()
}

// 读取文件内容
//
//   - path 文件路径
//
// 返回读取的字节内容
func readDataFromFile(path string) ([]byte, error) {
	// 打开文件
	file, e := os.OpenFile(path, os.O_RDONLY, 0755)
	if e != nil {
		return nil, e
	}
	defer func() {
		_ = file.Close()
	}()
	// 读取文件
	reader := bufio.NewReader(file)
	return io.ReadAll(reader)
}

// LoadTaskFile 从保存的任务文件中读取任务对象
//
//   - path 读取保存的任务文件
//
// 返回读取并反序列化后的任务对象切片
func LoadTaskFile[T comparable](path string) ([]T, error) {
	// 读取文件
	data, e := readDataFromFile(path)
	if e != nil {
		return nil, e
	}
	// 反序列化
	var tasks []T
	e = json.Unmarshal(data, &tasks)
	if e != nil {
		return nil, e
	}
	return tasks, nil
}