package log

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
)

// stackFrame 代表一帧栈信息
type stackFrame struct {
	Func string // 函数名
	File string // 文件路径
	Line int    // 行号
}

// framedError 是带栈帧的错误
type framedError struct {
	msg    string       // 本层添加的上下文消息
	cause  error        // 原始错误（可为空）
	frames []stackFrame // 本层包装时的调用栈（从调用 Wrap 的地方开始）
}

// Error 实现 error 接口
func (e *framedError) Error() string {
	if e.cause == nil {
		return e.msg
	}
	return e.msg + ": " + e.cause.Error()
}

// Unwrap 返回原始错误，支持 errors.Unwrap
func (e *framedError) Unwrap() error {
	return e.cause
}

// Format 实现 fmt.Formatter，支持 %+v 打印完整栈
func (e *framedError) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v':
		if f.Flag('+') {
			// 先打印本层消息和栈
			if e.msg != "" {
				fmt.Fprintf(f, "%s\n", e.msg)
			}
			for _, frame := range e.frames {
				fmt.Fprintf(f, "    %s\n        %s:%d\n", frame.Func, frame.File, frame.Line)
			}
			// 递归打印 cause（如果 cause 也是 framedError，会继续打印栈）
			if e.cause != nil {
				var fe *framedError
				if errors.As(e.cause, &fe) {
					fmt.Fprintf(f, "%+v", fe)
				}
			}
			return
		}
		fallthrough
	case 's':
		io.WriteString(f, e.Error())
	case 'q':
		fmt.Fprintf(f, "%q", e.Error())
	}
}

// captureStack 从指定 skip 层开始捕获栈帧（通常 skip=2：跳过 runtime.Callers 和本函数）
func captureStack(skip int) []stackFrame {
	const maxFrames = 32
	pcs := make([]uintptr, maxFrames)
	n := runtime.Callers(skip+1, pcs) // +1 跳过本函数
	if n == 0 {
		return nil
	}
	pcs = pcs[:n]
	frames := make([]stackFrame, 0, n)
	callerFrames := runtime.CallersFrames(pcs)
	for {
		frame, more := callerFrames.Next()
		// 过滤掉本包（errors）的内部调用，保持栈干净
		if !strings.Contains(frame.File, "/errors.") { // 根据你的包路径调整
			frames = append(frames, stackFrame{
				Func: frame.Function,
				File: frame.File,
				Line: frame.Line,
			})
		}
		if !more {
			break
		}
	}
	return frames
}

// New 创建一个新错误（带当前栈）
func New(msg string) error {
	return &framedError{
		msg:    msg,
		cause:  nil,
		frames: captureStack(2), // 跳过 New 和 captureStack
	}
}

// Errorf 格式化创建错误
func Errorf(format string, args ...any) error {
	return &framedError{
		msg:    fmt.Sprintf(format, args...),
		cause:  nil,
		frames: captureStack(2),
	}
}

// Wrap 包装错误，添加上下文（推荐用法）
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &framedError{
		msg:    msg,
		cause:  err,
		frames: captureStack(2), // 跳过 Wrap 和 captureStack
	}
}
func WrapError(err error) error {
	if err == nil {
		return nil
	}
	return &framedError{
		msg:    "",
		cause:  err,
		frames: captureStack(2), // 跳过 Wrap 和 captureStack
	}
}

// Wrapf 格式化包装
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &framedError{
		msg:    fmt.Sprintf(format, args...),
		cause:  err,
		frames: captureStack(2),
	}
}
