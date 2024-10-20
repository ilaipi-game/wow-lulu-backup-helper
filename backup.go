package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// 配置结构体
type Config struct {
	Addons   []string `yaml:"addons"`
	Accounts []string `yaml:"accounts"`
	Servers  []string `yaml:"servers"`
	Roles    []string `yaml:"roles"`
}

func main() {
	// 工作目录
	workdir, err := os.Getwd()
	if err != nil {
		fmt.Println("获取当前工作目录失败:", err)
		return
	}

	// 检查是否在正确目录
	if _, err := os.Stat(filepath.Join(workdir, "World of Warcraft Launcher.exe")); os.IsNotExist(err) {
		fmt.Println("请在安装目录执行，即和 World of Warcraft Launcher.exe 同目录")
		return
	}

	// 提示用户选择操作
	fmt.Println("选择操作：")
	fmt.Println("1. 备份")
	fmt.Println("2. 还原")
	fmt.Println("3. 初始化")
	var choice string
	fmt.Print("输入你的选择（1/2/3）：")
	fmt.Scanln(&choice)

	// 检查用户输入
	if choice != "1" && choice != "2" && choice != "3" {
		fmt.Println("无效的选择，请输入1/2/3。")
		return
	}

	// 创建临时文件夹
	tempDir := filepath.Join(workdir, "temp")
	if choice == "2" {
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			fmt.Println("未发现临时文件夹，无法还原。")
			return
		}
		fmt.Println("发现临时文件夹，开始还原...")
		restore(tempDir, workdir)
	} else if choice == "3" {
		// 删除指定目录
		cleanUpClassic(workdir)
	} else {
		err = os.MkdirAll(tempDir, os.ModePerm)
		if err != nil {
			fmt.Println("创建临时文件夹失败:", err)
			return
		}
		backup(tempDir, workdir)

		fmt.Println("备份完成！")
		return
	}

	fmt.Println("还原完成！")
}

// 备份逻辑
func backup(tempDir, workdir string) {
	// 读取配置文件
	config := Config{}
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Println("读取配置文件失败:", err)
		return
	}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		fmt.Println("解析配置文件失败:", err)
		return
	}

	// 复制 addons 到临时文件夹
	addonsDir := filepath.Join(workdir, "_classic_", "Interface", "AddOns")
	addonsTempDir := filepath.Join(tempDir, "AddOns")
	err = os.MkdirAll(addonsTempDir, os.ModePerm)
	if err != nil {
		fmt.Println("创建 AddOns 临时文件夹失败:", err)
		return
	}

	for _, addon := range config.Addons {
		source := filepath.Join(addonsDir, addon) // 确保使用原始 addon 名称
		destination := filepath.Join(addonsTempDir, addon)

		// 检查源目录是否存在
		if _, err := os.Stat(source); os.IsNotExist(err) {
			fmt.Printf("源目录 %s 不存在\n", source)
			continue
		}

		// 复制 AddOn 文件夹
		err := copyDir(source, destination)
		if err != nil {
			fmt.Printf("复制 AddOn %s 失败: %s\n", addon, err)
		} else {
			fmt.Printf("成功复制 AddOn %s\n", addon)
		}
	}

	// 复制 SavedVariables 文件
	for _, account := range config.Accounts {
		savedVarsDir := filepath.Join(workdir, "_classic_", "WTF", "Account", account, "SavedVariables")

		// 创建 SavedVariables 临时目录
		for _, addon := range config.Addons {
			source := filepath.Join(savedVarsDir, addon+".lua")

			// 目标路径保留原来的文件路径
			destination := filepath.Join(tempDir, "WTF", "Account", account, "SavedVariables", addon+".lua")

			// 创建目标路径
			err = os.MkdirAll(filepath.Dir(destination), os.ModePerm)
			if err != nil {
				fmt.Printf("创建目标目录 %s 失败: %s\n", filepath.Dir(destination), err)
				continue
			}

			// 检查 SavedVariables 文件是否存在
			if _, err := os.Stat(source); os.IsNotExist(err) {
				fmt.Printf("SavedVariables 文件 %s 不存在\n", source)
				continue
			}

			// 复制 SavedVariables 文件
			err := copyFile(source, destination)
			if err != nil {
				fmt.Printf("复制 SavedVariables 文件 %s 失败: %s\n", addon+".lua", err)
			} else {
				fmt.Printf("成功复制 SavedVariables 文件 %s\n", addon+".lua")
			}
		}
	}
}

// 还原逻辑
func restore(tempDir, workdir string) {
	// 还原 AddOns 目录
	addonsTempDir := filepath.Join(tempDir, "AddOns")
	addonsDir := filepath.Join(workdir, "_classic_", "Interface", "AddOns")

	err := copyDir(addonsTempDir, addonsDir)
	if err != nil {
		fmt.Printf("还原 AddOns 目录失败: %s\n", err)
	}

	// 还原 WTF 目录
	wtfTempDir := filepath.Join(tempDir, "WTF")
	wtfDir := filepath.Join(workdir, "_classic_", "WTF")

	err = copyDir(wtfTempDir, wtfDir)
	if err != nil {
		fmt.Printf("还原 WTF 目录失败: %s\n", err)
	}

	// 删除临时目录
	err = os.RemoveAll(tempDir)
	if err != nil {
		fmt.Printf("删除临时目录 %s 失败: %s\n", tempDir, err)
	} else {
		fmt.Printf("成功删除临时目录 %s\n", tempDir)
	}
}

// 删除指定的文件夹
func cleanUpClassic(workdir string) {
	directories := []string{"WTF", "Fonts", "Interface", "Cache", "blob_storage", "Errors", "CPUCache", "Interface.obsoleted"}

	for _, dir := range directories {
		path := filepath.Join(workdir, "_classic_", dir)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			err := os.RemoveAll(path)
			if err != nil {
				fmt.Printf("删除目录 %s 失败: %s\n", path, err)
			} else {
				fmt.Printf("成功删除目录 %s\n", path)
			}
		}
	}
}

// 复制文件
func copyFile(src string, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, input, 0644)
}

// 复制目录
func copyDir(src string, dst string) error {
	err := os.MkdirAll(dst, os.ModePerm) // 创建目标目录
	if err != nil {
		return err
	}
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
	for _, file := range files {
		srcFile := filepath.Join(src, file.Name())
		dstFile := filepath.Join(dst, file.Name())
		if file.IsDir() {
			err := copyDir(srcFile, dstFile) // 递归复制目录
			if err != nil {
				return err
			}
		} else {
			err := copyFile(srcFile, dstFile) // 复制文件
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// 检查目录是否为空
func isEmpty(dir string) bool {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("读取目录失败:", err)
		return false
	}
	return len(files) == 0
}
