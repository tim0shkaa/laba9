package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "300705"
	dbname   = "laba8_count"
)

// ▎Определение структуры Handlers для работы с запросами через Echo
type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

// ▎Метод для обработки GET-запроса (получение текущего значения)
func (h *Handlers) GetCount(c echo.Context) error {
	count, err := h.dbProvider.SelectCount()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]int{"count": count})
}

// ▎Метод для обработки POST-запроса (обновление значения)
func (h *Handlers) PostCount(c echo.Context) error {
	input := struct {
		Count int `json:"count"`
	}{}

	// Парсинг JSON-запроса
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	// Обновление значения в базе
	if err := h.dbProvider.UpdateCount(input.Count); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusCreated)
}

// ▎Методы для работы с базой данных
// Получение текущего значения count
func (dp *DatabaseProvider) SelectCount() (int, error) {
	var count int
	row := dp.db.QueryRow("SELECT number FROM counts")
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Обновление значения count
func (dp *DatabaseProvider) UpdateCount(addition int) error {
	_, err := dp.db.Exec("UPDATE counts SET number = number + $1", addition)
	return err
}

// ▎Основной код приложения
func main() {
	// Адрес для запуска сервера
	address := flag.String("address", "127.0.0.1:8081", "адрес для запуска сервера")
	flag.Parse()

	// Формирование строки подключения к PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Подключение к базе данных
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем провайдер базы данных и обработчики
	dp := DatabaseProvider{db: db}
	handlers := Handlers{dbProvider: dp}

	// Инициализация сервера Echo
	e := echo.New()

	// ▎Маршруты
	e.GET("/count", handlers.GetCount)   // Получение текущего значения
	e.POST("/count", handlers.PostCount) // Обновление значения

	// Запуск сервера
	fmt.Println("Starting server on:", *address)
	if err := e.Start(*address); err != nil {
		log.Fatal(err)
	}
}
