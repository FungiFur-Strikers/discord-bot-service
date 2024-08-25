package dify

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// DownloadFile は指定されたURLからファイルをダウンロードします
func DownloadFile(fileURL string) (*os.File, error) {
	println("http://dify.kajidog.com" + fileURL)
	resp, err := http.Get("https://dify.kajidog.com" + fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// 一時ファイルを作成
	tmpFile, err := os.CreateTemp("", "downloaded-*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	// レスポンスボディを一時ファイルにコピー
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// ファイルポインタを先頭に戻す
	tmpFile.Seek(0, 0)

	return tmpFile, nil
}
