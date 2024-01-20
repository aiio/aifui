package aifui

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/tidwall/gjson"
)

func TestUpdate(t *testing.T) {
	updateBefore()
	url := "https://api.github.com/repos/baidu/amis/releases/latest"
	req, _ := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req) // request remote
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close() // ！ close ReadCloser
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	result := gjson.ParseBytes(data)

	sdkUrl := "https://github.com/baidu/amis/releases/download/%v/sdk.tar.gz"
	schemaUrl := "https://github.com/baidu/amis/releases/download/%v/schema.json"

	updateVersion(result.Get("tag_name").String())
	tarGzURL := fmt.Sprintf(sdkUrl, result.Get("tag_name").String())

	// 下载文件
	err = downloadFile(tarGzURL, downloadDest)
	if err != nil {
		fmt.Println("下载文件时出错:", err)
		return
	}

	// 解压文件到目标目录
	err = extractTarGz(downloadDest, extractDest)
	if err != nil {
		fmt.Println("解压文件时出错:", err)
		return
	}
	defer os.Remove(downloadDest)

	err = downloadFile(fmt.Sprintf(schemaUrl, result.Get("tag_name").String()), extractDest+"/schema.json")
	if err != nil {
		fmt.Println("下载文件时出错:", err)
		return
	}

	fmt.Println("文件已下载并解压到目标目录:", extractDest)

}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func extractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}
		}
	}
}
