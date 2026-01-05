package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// 全局配置解析器实例
var config = make(map[string]map[string]string)

// 初始化配置：程序启动时加载配置文件 + 环境变量
func init() {
	// 获取配置文件路径（与Python逻辑一致：当前文件目录下的config.ini）
	configFile, err := getConfigFilePath()
	if err != nil {
		fmt.Printf("警告：获取配置文件路径失败，仅使用环境变量和默认值: %v\n", err)
		return
	}

	// 读取并解析配置文件
	err = parseIniFile(configFile)
	if err != nil {
		fmt.Printf("警告：配置文件解析失败，仅使用环境变量和默认值: %v\n", err)
	}
}

// 获取配置文件路径（兼容不同运行环境）
func getConfigFilePath() (string, error) {
	// 获取当前文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	execDir := filepath.Dir(execPath)

	// 优先查找当前文件目录下的config.ini
	configPath := filepath.Join(execDir, "config.ini")
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	// 备用：当前工作目录下的config/config.ini（与Python的Path(__file__).parent逻辑对齐）
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	altConfigPath := filepath.Join(wd, "config", "config.ini")
	if _, err := os.Stat(altConfigPath); err == nil {
		return altConfigPath, nil
	}

	return "", fmt.Errorf("配置文件未找到（已尝试：%s, %s）", configPath, altConfigPath)
}

// 解析INI格式配置文件
func parseIniFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// 匹配节（如 [app]）
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.TrimSpace(line[1 : len(line)-1])
			if _, exists := config[currentSection]; !exists {
				config[currentSection] = make(map[string]string)
			}
			continue
		}

		// 匹配键值对（如 port = 50100）
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // 跳过无效行
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// 移除值两侧的引号（兼容带引号的配置）
		value = strings.Trim(value, "\"'")

		if currentSection != "" {
			config[currentSection][key] = value
		}
	}

	return scanner.Err()
}

// GetConfig 统一读取配置：优先环境变量 → 配置文件 → 默认值
// 环境变量名格式：APP_{SECTION}_{KEY}（全大写）
func GetConfig(section, key string, defaultValue interface{}) interface{} {
	// 1. 优先读取环境变量
	envKey := fmt.Sprintf("APP_%s_%s", strings.ToUpper(section), strings.ToUpper(key))
	if envValue, exists := os.LookupEnv(envKey); exists {
		return envValue
	}

	// 2. 读取配置文件
	sectionMap, sectionExists := config[section]
	if sectionExists {
		if value, keyExists := sectionMap[key]; keyExists {
			return value
		}
	}

	// 3. 返回默认值
	return defaultValue
}

// -------------------------- 封装常用配置（直接导入使用） --------------------------

// 字符串类型配置
var (
	APP_NAME              = GetConfig("app", "name", "flask-echo").(string)
	APP_HOST              = GetConfig("server", "host", "0.0.0.0").(string)
	APP_LOG_PATH          = GetConfig("server", "log_path", "/app/log").(string)
	CONTAINER_LOG_PATH    = GetConfig("server", "container_log_path", "/var/log").(string)
	DOCKER_IMAGE_NAME     = GetConfig("docker", "image_name", "flask-echo").(string)
	DOCKER_CONTAINER_NAME = GetConfig("docker", "container_name", "flask-echo-container").(string)
)

// 数值/布尔类型配置（需要类型转换）
var (
	APP_PORT  = getIntConfig("app", "port", 50100)
	APP_DEBUG = getBoolConfig("app", "debug", false)
)

// 辅助函数：获取整数类型配置
func getIntConfig(section, key string, defaultValue int) int {
	value := GetConfig(section, key, fmt.Sprintf("%d", defaultValue))
	strVal, ok := value.(string)
	if !ok {
		return defaultValue
	}

	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return defaultValue
	}
	return intVal
}

// 辅助函数：获取布尔类型配置（兼容 true/false、1/0、yes/no）
func getBoolConfig(section, key string, defaultValue bool) bool {
	value := GetConfig(section, key, fmt.Sprintf("%t", defaultValue))
	strVal, ok := value.(string)
	if !ok {
		return defaultValue
	}

	strVal = strings.ToLower(strings.TrimSpace(strVal))
	switch strVal {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

// 辅助函数：格式化输出所有配置（调试用）
func PrintAllConfigs() {
	fmt.Println("=== 当前配置 ===")
	fmt.Printf("APP_NAME: %s\n", APP_NAME)
	fmt.Printf("APP_PORT: %d\n", APP_PORT)
	fmt.Printf("APP_HOST: %s\n", APP_HOST)
	fmt.Printf("APP_DEBUG: %t\n", APP_DEBUG)
	fmt.Printf("APP_LOG_PATH: %s\n", APP_LOG_PATH)
	fmt.Printf("CONTAINER_LOG_PATH: %s\n", CONTAINER_LOG_PATH)
	fmt.Printf("DOCKER_IMAGE_NAME: %s\n", DOCKER_IMAGE_NAME)
	fmt.Printf("DOCKER_CONTAINER_NAME: %s\n", DOCKER_CONTAINER_NAME)
}
