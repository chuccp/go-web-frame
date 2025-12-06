package util

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func EnsureDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// 创建目录及必要的父目录
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func WriteBase64File(base64Str string, dst string) error {
	base64, err := DecodeFileBase64(base64Str)
	if err != nil {
		return err
	}
	return WriteFile(base64, dst)
}
func WriteFile(bytes []byte, dst string) error {
	// 创建目标文件所在的目录
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	// 创建文件
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		// 确保文件关闭，并检查关闭时的错误
		closeErr := out.Close()
		if err == nil {
			err = closeErr
		}
	}()

	// 将字节数据写入文件
	_, err = out.Write(bytes)
	if err != nil {
		return err
	}
	// 确保数据被刷新到磁盘
	if err := out.Sync(); err != nil {
		return err
	}
	return nil
}
func ReadFileBytes(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println("Error closing file:", err)
		}
	}(file)
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fileInfo.Size()
	buffer := make([]byte, fileSize)
	_, err = file.Read(buffer)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}
func ReadFile(path string) (string, error) {
	data, err := ReadFileBytes(path)
	return string(data), err
}

type File struct {
	normal string
	file   *os.File
	isDir  bool
	isDisk bool
}

func (f *File) Abs() string {
	return f.normal
}
func (f *File) Parent() string {
	return filepath.Dir(f.normal)
}
func (f *File) Name() string {
	if f.isDisk {
		return f.normal[0 : len(f.normal)-1]
	}
	return filepath.Base(f.normal)
}
func (f *File) ParentFile() (*File, error) {
	return NewFile(f.Parent())
}
func (f *File) open() error {
	if f.file == nil {
		file, err := os.Open(f.normal)
		if err != nil {
			return err
		}
		f.file = file
	}
	return nil
}
func (f *File) OpenAppendOrCreate() error {
	err := f.mkParent()
	if err != nil {
		return err
	}
	file1, err := os.OpenFile(f.normal, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	f.file = file1
	return err
}
func (f *File) mkParent() error {
	f, err := f.ParentFile()
	if err != nil {
		return err
	}
	err2 := f.MkDirs()
	return err2
}
func (f *File) OpenOrCreate() error {
	err := f.mkParent()
	if err != nil {
		return err
	}
	file1, err := os.OpenFile(f.normal, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	f.file = file1
	f.isDir = false
	return err
}

func (f *File) Exists() (flag bool, err error) {
	err = f.open()
	if err != nil {
		if os.IsNotExist(err) {
			return false, os.ErrNotExist
		}
		return true, err
	}
	return true, nil
}

func (f *File) ModTime() (*time.Time, error) {
	stat, err := f.file.Stat()
	if err != nil {
		return nil, err
	}
	t := stat.ModTime()
	return &t, nil
}

func (f *File) MkDirs() error {
	err2 := os.MkdirAll(f.Abs(), 0666)
	return err2
}
func (f *File) OpenWrite() (*bufio.Writer, error) {
	err := f.OpenOrCreate()
	if err != nil {
		return nil, err
	}
	return bufio.NewWriter(f.file), nil
}
func (f *File) OpenAppendWrite() (*bufio.Writer, error) {
	err := f.OpenAppendOrCreate()
	if err != nil {
		return nil, err
	}
	return bufio.NewWriter(f.file), nil
}
func (f *File) Truncate() error {
	err := f.open()
	if err == nil {
		return f.file.Truncate(0)
	} else {
		return err
	}
}
func (f *File) Close() error {
	err := f.file.Sync()
	if err != nil {
		return err
	}
	return f.file.Close()

}
func (f *File) ReadBytes(p []byte) (n int, err error) {
	flag, err := f.Exists()
	if !flag {
		return 0, err
	}
	return f.file.Read(p)
}

func (f *File) ToRawFile() (*os.File, error) {
	flag, err := f.Exists()
	if !flag {
		return nil, err
	}
	return f.file, nil
}
func (f *File) ReadAll() ([]byte, error) {
	flag, err := f.Exists()
	if !flag {
		return nil, err
	}
	var allData = make([]byte, 0)
	var reader = bufio.NewReader(f.file)
	for {
		data := make([]byte, 1024)
		num, err := reader.Read(data)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				break
			}
			return nil, err
		}
		if num == 0 {
			break
		}
		allData = append(allData, data[0:num]...)
	}
	return allData, nil
}

func (f *File) WriteBytes(data []byte) error {
	if f.isDir {
		return errors.New(f.normal + " " + syscall.EISDIR.Error())
	}
	bw, err := f.OpenWrite()
	if err != nil {
		return err
	}
	err = f.file.Truncate(0)
	if err != nil {
		return err
	}
	//os.Truncate(name, size)
	_, err = bw.Write(data)
	if err != nil {
		return err
	}
	return bw.Flush()
}

func (f *File) WriteAppendBytes(data []byte) error {
	if f.isDir {
		return errors.New(f.normal + " " + syscall.EISDIR.Error())
	}
	bw, err := f.OpenAppendWrite()
	if err != nil {
		return err
	}
	_, err = bw.Write(data)
	if err != nil {
		return err
	}
	return bw.Flush()
}
func (f *File) List() ([]*File, error) {
	dirs, err := os.ReadDir(f.normal)
	if err != nil {
		return nil, err
	}
	var files = make([]*File, 0)
	for _, dir := range dirs {
		filePath := path.Join(f.normal, dir.Name())
		file, err3 := NewFile(filePath)
		if err3 == nil {
			files = append(files, file)
		}
	}
	return files, err
}
func (f *File) Child(path string) (*File, error) {
	return NewFile(filepath.Join(f.normal, path))
}
func (f *File) IsDir() bool {
	return f.isDir
}
func (f *File) IsDisk() bool {
	return f.isDisk
}

func NewFile(path string) (*File, error) {
	normal, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	fa, _ := regexp.MatchString(`^[a-zA-Z]:\\$`, path)
	if fa {
		return &File{normal: normal, isDir: true, isDisk: true}, nil
	}
	fileInfo, err := os.Stat(normal)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") || strings.Contains(err.Error(), "cannot find the file") || strings.Contains(err.Error(), "cannot find the path") {
			return &File{normal: normal}, nil
		}
		return nil, err
	}
	return &File{normal: normal, isDir: fileInfo.IsDir()}, nil
}
func GetRootPath() ([]*File, error) {
	if runtime.GOOS == "windows" {
		return getWindowsRootPath()
	}
	return getOtherRootPath()
}

type storageInfo struct {
	Name       string
	Size       uint64
	FreeSpace  uint64
	FileSystem string
}

func getWindowsRootPath() ([]*File, error) {
	storageInfos, err := logicalDisk()
	if err != nil {
		return nil, err
	}
	files := make([]*File, 0)
	for _, v := range storageInfos {
		f, err := NewFile(v.Name + (string)(filepath.Separator))
		if err == nil {
			files = append(files, f)
		}
	}

	return files, err
}

func logicalDisk() (storageInfos []*storageInfo, err error) {
	storageInfos = make([]*storageInfo, 0)
	cmd := exec.Command("wmic", "logicaldisk", "get", "DeviceID,FreeSpace,Size,DriveType")
	stdout, err1 := cmd.StdoutPipe()
	if err1 != nil {
		return storageInfos, err
	}
	err = cmd.Start()
	if err == nil {
		scanner := bufio.NewScanner(stdout)
		scanner.Scan()
		scanner.Text()
		for scanner.Scan() {
			text := scanner.Text()
			values := strings.Fields(text)
			if len(values) == 4 {
				var si storageInfo
				si.Name = values[0]
				si.Size, _ = strconv.ParseUint(values[3], 10, 64)
				si.FreeSpace, _ = strconv.ParseUint(values[2], 10, 64)
				si.FileSystem = values[1]
				storageInfos = append(storageInfos, &si)
			}
		}
	}
	return storageInfos, err
}
func WriteBytesFilePath(path string, dataS ...[]byte) error {
	file, err := NewFile(path)
	if err != nil {
		return err
	}
	b := new(bytes.Buffer)
	for _, data := range dataS {
		b.Write(data)
	}
	err = file.WriteBytes(b.Bytes())
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println("file close fail:", err)
			return
		}
	}()
	return err
}
func WriteBytesFile(file *os.File, dataS ...[]byte) error {
	w := bufio.NewWriter(file)
	for _, data := range dataS {
		_, err := w.Write(data)
		if err != nil {
			return err
		}
	}
	err := w.Flush()
	if err != nil {
		return err
	}
	err = file.Sync()
	if err != nil {
		return err
	}
	return file.Truncate(0)
}
func ExistsFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return os.IsExist(err)
	}
	err = f.Close()
	if err != nil {
		log.Println("file close fail:", err)
	}
	return true
}

// ReadOrWriteFile 尝试只读打开文件，若不存在则调用 f 生成内容并原子写入
// 成功时总是返回一个可读的 *os.File（调用者负责 Close）
func ReadOrWriteFile(path string, f func() ([]byte, error)) (*os.File, error) {
	if file, err := os.Open(path); err == nil {
		return file, nil // 存在且可读 → 直接返回
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("无法读取已有文件 %s: %w", path, err)
	} // 2. 文件不存在，走写入流程
	data, err := f()
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		if os.IsExist(err) {
			return os.Open(path)
		}
		return nil, fmt.Errorf("创建文件失败: %w", err)
	}
	cleanup := true
	defer func() {
		if !cleanup {
			return
		}
		file.Close()
		_ = os.Remove(path)
	}()

	if _, err = file.Write(data); err != nil {
		return nil, err
	}
	if err = file.Sync(); err != nil {
		return nil, err
	}
	if err = file.Close(); err != nil {
		return nil, fmt.Errorf("close 失败（数据已安全）: %w", err)
	}
	cleanup = false
	return os.Open(path)
}

func getOtherRootPath() ([]*File, error) {
	dirs, err := os.ReadDir("/")
	if err == nil {
		files := make([]*File, 0)
		for _, v := range dirs {
			fi, err8 := NewFile(v.Name())
			if err8 == nil {
				files = append(files, fi)
			}
		}
		return files, nil
	}
	return nil, err
}
