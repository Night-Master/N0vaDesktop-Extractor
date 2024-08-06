package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	r := gin.Default()

	// 静态文件服务
	r.Use(static.Serve("/", static.LocalFile("./frontend", false)))

	// 处理表单提交
	r.POST("/convert", func(c *gin.Context) {
		sourceDir := c.PostForm("sourceDir")

		if sourceDir == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Source directory is required"})
			return
		}

		outputDir, err := getOutputDir()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := convertFiles(sourceDir, outputDir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Files converted successfully"})
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

func convertFiles(sourceDir, targetDir string) error {
	files, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		return err
	}

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

		destPath := filepath.Join(targetDir, uuid.New().String()+getFileExtension(fileType))

		data, err := ioutil.ReadFile(srcPath)
		if err != nil {
			return err
		}

		// 只有视频文件需要删除前两个字节
		if fileType == FTYP_VIDEO_FILE {
			if len(data) > 2 {
				data = data[2:]
			}
		}

		if err := ioutil.WriteFile(destPath, data, 0644); err != nil {
			return err
		}
	}

	return nil
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
