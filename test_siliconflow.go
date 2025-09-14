package main

import (
	"fmt"
	"log"
	"os"

	"hajimi-king-go/internal/config"
	"hajimi-king-go/internal/siliconflow"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()
	
	// 检查是否提供了API密钥
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_siliconflow.go <siliconflow_api_key>")
		fmt.Println("Example: go run test_siliconflow.go sk-...")
		os.Exit(1)
	}
	
	apiKey := os.Args[1]
	
	fmt.Printf("🔍 Testing SiliconFlow API key: %s...\n", apiKey[:10]+"...")
	
	// 创建SiliconFlow客户端
	client := siliconflow.NewClient(apiKey)
	
	// 测试API密钥验证
	fmt.Println("\n1. Testing API key validation...")
	result, err := client.ValidateAPIKey()
	if err != nil {
		log.Fatalf("❌ Validation failed: %v", err)
	}
	
	fmt.Printf("   Result: %s\n", result)
	
	if result == "ok" {
		fmt.Println("✅ API key is valid!")
		
		// 测试聊天功能
		fmt.Println("\n2. Testing chat completion...")
		testResult, err := client.TestAPIKey()
		if err != nil {
			fmt.Printf("⚠️ Chat test failed: %v\n", err)
		} else {
			fmt.Printf("   Chat test result: %s\n", testResult)
			if testResult == "ok" {
				fmt.Println("✅ Chat completion works!")
			}
		}
		
		// 获取可用模型
		fmt.Println("\n3. Getting available models...")
		models, err := client.GetAvailableModels()
		if err != nil {
			fmt.Printf("⚠️ Failed to get models: %v\n", err)
		} else {
			fmt.Printf("   Found %d models\n", len(models))
			if len(models) > 0 {
				fmt.Printf("   First few models: %v\n", models[:min(5, len(models))])
			}
		}
		
		// 测试特定模型信息
		if len(models) > 0 {
			fmt.Println("\n4. Testing model info...")
			modelInfo, err := client.GetModelInfo(models[0])
			if err != nil {
				fmt.Printf("⚠️ Failed to get model info: %v\n", err)
			} else {
				fmt.Printf("   Model ID: %s\n", modelInfo.ID)
				fmt.Printf("   Model Object: %s\n", modelInfo.Object)
				fmt.Printf("   Model Owner: %s\n", modelInfo.OwnedBy)
			}
		}
		
	} else {
		fmt.Printf("❌ API key validation failed: %s\n", result)
	}
	
	fmt.Println("\n🎉 Test completed!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}