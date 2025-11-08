package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	// Получаем токен из переменных окружения
	token := "f9LHodD0cOLWyve7bMP1johhuivskU_nvmuFpZPBQnfk7Ba0FZys46eabM65ctLT7LKbb1P_SfIH4hxtAVrY"
	if token == "" {
		log.Fatal("MAX_API_TOKEN environment variable is required")
	}

	// Создаем HTTP клиент
	client := &http.Client{}

	// Создаем запрос
	req, err := http.NewRequest("GET", "https://platform-api.max.ru/me", nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	// Добавляем Authorization header
	req.Header.Set("Authorization", token)

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Читаем и выводим ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))
}
