package main

import (
	"bytes"
	"crypto/md5"
	"embed"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg" // 注册 JPEG/EXIF 解码器供 image.DecodeConfig 使用
	_ "image/png"  // 注册 PNG 解码器供 image.DecodeConfig 使用
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"golang.org/x/sys/windows"
)

//go:embed frontend
var frontendFS embed.FS

func main() {
	r := gin.Default()

	// 静态文件服务（内嵌前端，单 exe 发布）
	r.Use(static.Serve("/", static.EmbedFolder(frontendFS, "frontend")))

	// 处理表单提交；sourceDir 留空时自动扫描所有盘符的常见安装路径
	r.POST("/convert", func(c *gin.Context) {
		sourceDir := c.PostForm("sourceDir")

		var sourceDirs []string
		if sourceDir != "" {
			sourceDirs = []string{sourceDir}
		} else {
			sourceDirs = findSourceDirs()
			if len(sourceDirs) == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "未找到人工桌面缓存目录，请手动输入源目录"})
				return
			}
		}

		outputDir, err := getOutputDir()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		converted, skipped := 0, 0
		for _, dir := range sourceDirs {
			n, s, err := convertFiles(dir, outputDir)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			converted += n
			skipped += s
		}

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("已从 %d 个目录提取 %d 个文件（跳过重复 %d 个）", len(sourceDirs), converted, skipped)})
	})

	// 页面打开时自动扫描所有盘符的常见安装路径
	r.GET("/scan", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"dirs": findSourceDirs()})
	})

	r.Run(":8080")
}

func getOutputDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	outputDir := filepath.Join(filepath.Dir(exePath), "output")
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.Mkdir(outputDir, 0755); err != nil {
			return "", err
		}
	}

	return outputDir, nil
}

// minPixels 是提取的最低像素数（宽×高），低于该值的图片和视频跳过
const minPixels = 936000

func convertFiles(sourceDir, targetDir string) (int, int, error) {
	files, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		return 0, 0, err
	}

	converted, skipped := 0, 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".ndf") {
			continue
		}

		if file.Size() < 300*1024 { // 小于300KB的文件不转换
			continue
		}

		srcPath := filepath.Join(sourceDir, file.Name())
		fileType := getFileType(srcPath)
		if fileType == UNRECOGNIZED_FILE {
			continue
		}

		data, err := ioutil.ReadFile(srcPath)
		if err != nil {
			return converted, skipped, err
		}

		// 只有视频文件需要删除前两个字节
		if fileType == FTYP_VIDEO_FILE {
			if len(data) > 2 {
				data = data[2:]
			}
		}

		// 分辨率相乘小于 minPixels 的图片和视频不转换；无法解析分辨率时保守保留
		if w, h, err := getDimensions(data, fileType); err == nil && w*h < minPixels {
			continue
		}

		// 以内容 MD5 命名，已存在同名同大小文件即重复，跳过
		sum := md5.Sum(data)
		destPath := filepath.Join(targetDir, hex.EncodeToString(sum[:])+getFileExtension(fileType))
		if info, err := os.Stat(destPath); err == nil && info.Size() == int64(len(data)) {
			skipped++
			continue
		}

		if err := ioutil.WriteFile(destPath, data, 0644); err != nil {
			return converted, skipped, err
		}
		converted++
	}

	return converted, skipped, nil
}

// getDimensions 返回图片或视频的宽高
func getDimensions(data []byte, fileType int) (int, int, error) {
	if fileType == FTYP_VIDEO_FILE {
		return getMP4Dimensions(data)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

// getMP4Dimensions 解析 mp4 box 结构，取第一个有效 trak 的 tkhd 宽高（16.16 定点数）
func getMP4Dimensions(data []byte) (int, int, error) {
	moov := findBox(data, "moov")
	if moov == nil {
		return 0, 0, fmt.Errorf("moov box not found")
	}
	for len(moov) >= 8 {
		boxLen, boxType, header := parseBoxHeader(moov)
		if boxLen < header || boxLen > len(moov) {
			break
		}
		if boxType == "trak" {
			if tkhd := findBox(moov[header:boxLen], "tkhd"); len(tkhd) >= 8 {
				// tkhd 内容最后 8 字节是 width 和 height（16.16 定点数）
				w := int(binary.BigEndian.Uint32(tkhd[len(tkhd)-8 : len(tkhd)-4]) >> 16)
				h := int(binary.BigEndian.Uint32(tkhd[len(tkhd)-4:]) >> 16)
				if w > 0 && h > 0 {
					return w, h, nil
				}
			}
		}
		moov = moov[boxLen:]
	}
	return 0, 0, fmt.Errorf("tkhd box not found")
}

// parseBoxHeader 解析 mp4 box 头，返回 box 总长（含头部）、类型和头部长度
func parseBoxHeader(data []byte) (int, string, int) {
	boxLen := int(binary.BigEndian.Uint32(data[0:4]))
	boxType := string(data[4:8])
	if boxLen == 1 { // 64 位扩展长度
		if len(data) < 16 {
			return 0, boxType, 8
		}
		return int(binary.BigEndian.Uint64(data[8:16])), boxType, 16
	}
	if boxLen == 0 { // 延伸到数据末尾
		return len(data), boxType, 8
	}
	return boxLen, boxType, 8
}

// findBox 在 data 的 box 序列中查找第一个指定类型的 box，返回其内容（不含头部）
func findBox(data []byte, wantType string) []byte {
	for len(data) >= 8 {
		boxLen, boxType, header := parseBoxHeader(data)
		if boxLen < header || boxLen > len(data) {
			return nil
		}
		if boxType == wantType {
			return data[header:boxLen]
		}
		data = data[boxLen:]
	}
	return nil
}

// gameDirTemplates 是相对盘符根目录的常见缓存路径
var gameDirTemplates = []string{
	`Program Files\N0vaDesktop\N0vaDesktopCache\game`,
	`Program Files (x86)\N0vaDesktop\N0vaDesktopCache\game`,
	`N0vaDesktop\N0vaDesktopCache\game`,
}

// getDriveLetters 优先用 GetLogicalDrives 获取所有盘符，失败则遍历 A-Z
func getDriveLetters() []string {
	var drives []string
	bitmask, err := windows.GetLogicalDrives()
	if err == nil {
		for i := 0; i < 26; i++ {
			if bitmask&(1<<uint(i)) != 0 {
				drives = append(drives, string(rune('A'+i)))
			}
		}
	}
	if len(drives) == 0 {
		for c := 'A'; c <= 'Z'; c++ {
			drives = append(drives, string(c))
		}
	}
	return drives
}

// findSourceDirs 拼接各盘符与预设路径，返回实际存在的 game 目录
func findSourceDirs() []string {
	var dirs []string
	for _, drive := range getDriveLetters() {
		for _, tpl := range gameDirTemplates {
			dir := drive + `:\` + tpl
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				dirs = append(dirs, dir)
			}
		}
	}
	return dirs
}

func getFileType(filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		return UNRECOGNIZED_FILE
	}
	defer file.Close()

	magic := make([]byte, 10)
	_, err = file.Read(magic)
	if err != nil {
		return UNRECOGNIZED_FILE
	}

	if compareFromHead(magic, EXIF_IMAGE_MAGIC, 4) {
		return EXIF_FILE
	} else if compareFromHead(magic, PNG_IMAGE_MAGIC, 8) {
		return PNG_FILE
	} else if compareFromHead(magic, JPEG_IMAGE_MAGIC, 4) {
		return JPEG_FILE
	} else if compareFromHead(magic, FTYPMP42_VIDEO_MAGIC, 10) || compareFromHead(magic, FTYPISOM_VIDEO_MAGIC, 10) {
		return FTYP_VIDEO_FILE
	} else {
		return UNRECOGNIZED_FILE
	}
}

func compareFromHead(toBeCompared, pattern []byte, nPattern int) bool {
	for i := 0; i < nPattern; i++ {
		if toBeCompared[i] != pattern[i] {
			return false
		}
	}
	return true
}

func getFileExtension(fileType int) string {
	switch fileType {
	case EXIF_FILE, JPEG_FILE:
		return ".jpg"
	case PNG_FILE:
		return ".png"
	case FTYP_VIDEO_FILE:
		return ".mp4"
	default:
		return ""
	}
}

const (
	UNRECOGNIZED_FILE = iota
	EXIF_FILE
	PNG_FILE
	JPEG_FILE
	FTYP_VIDEO_FILE
)

var (
	EXIF_IMAGE_MAGIC     = []byte{0xff, 0xd8, 0xff, 0xe1}
	PNG_IMAGE_MAGIC      = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	JPEG_IMAGE_MAGIC     = []byte{0xff, 0xd8, 0xff, 0xe0}
	FTYPISOM_VIDEO_MAGIC = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70}
	FTYPMP42_VIDEO_MAGIC = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70}
)
