package common

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

// 文件类型常量
const (
	TXT_TYPE  = "text/plain"
	PDF_TYPE  = "application/pdf"
	DOC_TYPE  = "application/msword"
	JPG_TYPE  = "image/jpeg"
	PNG_TYPE  = "image/png"
	WEBP_TYPE = "image/webp"
)

// 检测文件类型结果
type FileTypeResult struct {
	MimeType    string
	Extension   string
	Description string
	IsValid     bool
}

// 从带前缀的base64数据中直接解析MIME类型
func getMimeTypeFromDataURI(dataURI string) string {
	// data:text/plain;base64,xxxxx 格式
	regex := regexp.MustCompile(`data:([^;]+);base64,`)
	matches := regex.FindStringSubmatch(dataURI)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// 检测是否为文本文件的函数 - 增强版
func isTextFile(data []byte) bool {
	// 检查多种文本文件格式

	// 如果数据为空，则不是有效的文本文件
	if len(data) == 0 {
		return false
	}

	// 检查是否有BOM (UTF-8, UTF-16)
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) || // UTF-8 BOM
		bytes.HasPrefix(data, []byte{0xFE, 0xFF}) || // UTF-16 BE BOM
		bytes.HasPrefix(data, []byte{0xFF, 0xFE}) { // UTF-16 LE BOM
		return true
	}

	// 检查是否只包含ASCII字符或常见UTF-8序列
	// 我们会检查文件的前4KB和最后1KB（或整个文件如果小于5KB）
	checkSize := 4096
	if len(data) < checkSize {
		checkSize = len(data)
	}

	totalNonPrintable := 0
	totalChars := 0

	// 检查文件开头
	for i := 0; i < checkSize; i++ {
		b := data[i]
		totalChars++

		// 允许常见控制字符：TAB(9), LF(10), CR(13)
		if b != 9 && b != 10 && b != 13 {
			// 检查是否为可打印ASCII或常见UTF-8多字节序列的开始
			if (b < 32 || b > 126) && b < 192 { // 非可打印ASCII且不是UTF-8多字节序列开始
				totalNonPrintable++
			}
		}
	}

	// 如果文件较大，也检查文件结尾
	if len(data) > 5120 {
		endOffset := len(data) - 1024
		for i := 0; i < 1024; i++ {
			b := data[endOffset+i]
			totalChars++

			if b != 9 && b != 10 && b != 13 {
				if (b < 32 || b > 126) && b < 192 {
					totalNonPrintable++
				}
			}
		}
	}

	// 如果非可打印字符比例低于5%，则认为是文本文件
	return float64(totalNonPrintable)/float64(totalChars) < 0.05
}

// 增强的文件类型检测，专门处理text/plain
func DetectFileType(base64Data string) *FileTypeResult {
	// 检查是否有数据URI前缀
	mimeFromPrefix := getMimeTypeFromDataURI(base64Data)
	if mimeFromPrefix == TXT_TYPE {
		// 直接从前缀确认是文本类型
		return &FileTypeResult{
			MimeType:    TXT_TYPE,
			Extension:   ".txt",
			Description: "Plain Text Document",
			IsValid:     true,
		}
	}

	// 移除base64前缀
	commaIndex := strings.Index(base64Data, ",")
	if commaIndex != -1 {
		base64Data = base64Data[commaIndex+1:]
	}

	// 解码base64
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return &FileTypeResult{
			IsValid:     false,
			Description: "Base64 解码失败",
		}
	}

	// 检查常见文件魔数
	if len(data) >= 4 && bytes.HasPrefix(data, []byte("%PDF")) {
		return &FileTypeResult{
			MimeType:    PDF_TYPE,
			Extension:   ".pdf",
			Description: "PDF Document",
			IsValid:     true,
		}
	}

	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return &FileTypeResult{
			MimeType:    JPG_TYPE,
			Extension:   ".jpg",
			Description: "JPEG Image",
			IsValid:     true,
		}
	}

	if len(data) >= 8 && bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return &FileTypeResult{
			MimeType:    PNG_TYPE,
			Extension:   ".png",
			Description: "PNG Image",
			IsValid:     true,
		}
	}

	if len(data) >= 12 && bytes.HasPrefix(data, []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return &FileTypeResult{
			MimeType:    WEBP_TYPE,
			Extension:   ".webp",
			Description: "WebP Image",
			IsValid:     true,
		}
	}

	if len(data) >= 8 && bytes.HasPrefix(data, []byte{0xD0, 0xCF, 0x11, 0xE0}) {
		return &FileTypeResult{
			MimeType:    DOC_TYPE,
			Extension:   ".doc",
			Description: "Microsoft Word Document",
			IsValid:     true,
		}
	}

	// 增强的文本检测
	if isTextFile(data) {
		return &FileTypeResult{
			MimeType:    TXT_TYPE,
			Extension:   ".txt",
			Description: "Plain Text Document",
			IsValid:     true,
		}
	}

	// 默认返回未知类型
	return &FileTypeResult{
		IsValid:     false,
		Description: "未识别文件类型",
	}
}

func main() {
	// 示例：检测携带MIME前缀的TXT文件
	textWithPrefix := "data:text/plain;base64,SGVsbG8gV29ybGQh" // "Hello World!" 的base64

	result := DetectFileType(textWithPrefix)
	if result.IsValid {
		fmt.Printf("检测结果: %s (%s)\n", result.Description, result.MimeType)
	} else {
		fmt.Printf("错误: %s\n", result.Description)
	}
}
