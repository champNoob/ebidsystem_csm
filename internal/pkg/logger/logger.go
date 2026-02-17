package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logDir     string
	logDirOnce sync.Once
)

type Logger struct {
	ch        chan string
	file      *os.File
	wg        sync.WaitGroup
	isConsole bool
}

func NewLogger(buffer int, filePath string, console bool) (*Logger, error) {
	var f *os.File
	var err error

	realPath := getLogDir() + "/" + filePath

	// 确保路径合法：
	if realPath == "" {
		return nil, fmt.Errorf("文件路径不能为空")
	}
	// 确保目录存在：
	dir := filepath.Dir(realPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %v", err)
	}
	// 打开目标文件：
	f, err = os.OpenFile(realPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败: %v", err)
	}
	// 创建日志实例：
	l := &Logger{
		ch:        make(chan string, buffer),
		file:      f,
		isConsole: console,
	}

	l.wg.Add(1)
	go l.run()

	return l, nil
}

func (l *Logger) run() {
	defer l.wg.Done()

	if l.file != nil {
		log.Printf("日志实例已启动，目标文件路径: %s", l.file.Name())
	}

	for msg := range l.ch {
		if l.isConsole {
			log.Println(msg)
		}
		if l.file != nil {
			_, _ = l.file.WriteString(msg + "\n")
			// 移除强制同步，提高性能
			// _ = l.file.Sync() // 强制日志立即写入文件
		}
	}
}

func (l *Logger) Log(msg string) {
	select {
	case l.ch <- msg:
	default:
		log.Printf("日志被丢弃，channel已满: %s", msg)
		//# 丢弃日志，避免阻塞撮合主流程
	}
}

func (l *Logger) Close() {
	close(l.ch)

	// 添加超时机制，避免无限等待：
	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		//正常等待完成
	case <-time.After(5 * time.Second):
		log.Printf("WARNING: Logger close timeout, forcing shutdown") //超时后强制关闭
	}

	if l.file != nil {
		// 在关闭前进行一次同步
		_ = l.file.Sync()
		_ = l.file.Close()
	}
}

func getLogDir() string {
	logDirOnce.Do(func() {
		workDir, _ := os.Getwd()

		// 向上查找项目根目录（包含 go.mod 文件的目录）：
		projectRoot := workDir
		for {
			if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
				break
			}

			parent := filepath.Dir(projectRoot)
			if parent == projectRoot { //已经到达文件系统根目录，无法找到 go.mod
				log.Fatalf("无法找到项目根目录（go.mod 文件）")
			}
			projectRoot = parent
		}

		logDstDir := filepath.Join(projectRoot, "log")
		if err := os.MkdirAll(logDstDir, 0755); err != nil {
			log.Fatalf("创建日志目录失败: %v", err)
		}
		logDir = logDstDir
	})
	return logDir
}
