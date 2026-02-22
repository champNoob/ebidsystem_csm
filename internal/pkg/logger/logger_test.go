package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// 简单文件写入测试：
func TestLogger_SimpleWrite(t *testing.T) {
	submitLogger, err := NewLogger(
		1000,
		"engine/engine_submit.log",
		true,
		false,
	)
	if err != nil {
		log.Printf("创建引擎日志失败: %v", err)
		panic(err)
	}

	submitLogger.Log("测试消息1")
	submitLogger.Log("测试消息2")
	submitLogger.Log("测试消息3")
	time.Sleep(1 * time.Second)
	submitLogger.Close()
}

// 并发写入测试：
func TestLogger_ConcurrentWrites(t *testing.T) {
	logDir := filepath.Join(getLogDir(), "logger")
	logFile := filepath.Join(logDir, "logger_test.log")

	// 清空目标日志文件
	if _, err := os.Stat(logFile); err == nil {
		if err := os.Truncate(logFile, 0); err != nil {
			t.Fatalf("清空日志文件失败: %v", err)
		}
	}

	// 创建logger实例，增大缓冲区
	logger, err := NewLogger(
		1000,
		logFile,
		true,
		false,
	)
	if err != nil {
		t.Fatalf("创建logger失败: %v", err)
	}
	defer logger.Close()

	// 并发写入测试
	const (
		numGoroutines        = 5
		messagesPerGoroutine = 20
	)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 启动多个goroutine并发写入
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				msg := fmt.Sprintf("Goroutine-%d Message-%d", goroutineID, j)
				logger.Log(msg)
			}
		}(i)
	}

	// 等待所有写入完成
	wg.Wait()

	// 等待更长时间确保异步写入完成
	time.Sleep(2 * time.Second)

	// 先检查文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatalf("日志文件不存在: %s", logFile)
	}

	// 再验证文件内容
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	expectedLines := numGoroutines * messagesPerGoroutine

	if len(lines) != expectedLines {
		t.Errorf("期望 %d 行日志，实际得到 %d 行", expectedLines, len(lines))
	}

	// 再验证每行内容格式
	for i, line := range lines {
		if !strings.Contains(line, "Goroutine-") || !strings.Contains(line, "Message-") {
			t.Errorf("第 %d 行格式不正确: %s", i+1, line)
		}
	}

	t.Logf("成功写入 %d 行日志到文件: %s", len(lines), logFile)

	// 注释掉清理，方便查看文件内容
	// os.RemoveAll(logDir)
}

// 缓冲区溢出测试：
func TestLogger_BufferOverflow(t *testing.T) {
	logDir := filepath.Join(getLogDir(), "logger")
	logFile := filepath.Join(logDir, "buffer_test.log")

	// 清空目标日志文件
	if _, err := os.Stat(logFile); err == nil {
		if err := os.Truncate(logFile, 0); err != nil {
			t.Fatalf("清空日志文件失败: %v", err)
		}
	}

	// 创建小缓冲区的logger
	logger, err := NewLogger(10,
		logFile,
		true,
		false,
	)
	if err != nil {
		t.Fatalf("创建logger失败: %v", err)
	}
	defer logger.Close()

	// 快速写入超过缓冲区大小的消息
	for i := 0; i < 50; i++ {
		logger.Log(fmt.Sprintf("Message %d", i))
	}

	time.Sleep(100 * time.Millisecond)

	// 再验证文件存在且有内容
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	if len(content) == 0 {
		t.Error("日志文件为空")
	}

	t.Logf("缓冲区溢出测试完成，文件大小: %d 字节", len(content))
}
