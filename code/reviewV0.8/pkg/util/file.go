package util

import (
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func CreateNewFileName(originalFilename string) string {
	//获取文件后缀名
	ext := filepath.Ext(originalFilename)
	//生成文件名
	uuid := uuid.New().String()

	//拼接新的文件名
	return uuid + ext
}

// 验证图片文件类型
func IsValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	return validExts[ext]
}
