package pkg

import (
	"log"
	"os"
)

func DirExists(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		log.Println("创建目录: ", path)
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
		return nil
	}
	log.Printf("目录检查失败: %s, 错误: %s", path, err)
	return err
}
