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
	dbname   = "sandbox"
)

// Определение структуры Handlers для работы с запросами через Echo
type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

// Метод для обработки Get-запроса (Echo style)
func (h *Handlers) GetHello(c echo.Context) error {
	msg, err := h.dbProvider.SelectHello()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": msg})
}

// Метод для обработки Post-запроса (Echo style)
func (h *Handlers) PostHello(c echo.Context) error {
	input := struct {
		Msg string `json:"msg"`
	}{}

	// Парсинг JSON-запроса с помощью Echo
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Вставка данных в базу
	if err := h.dbProvider.InsertHello(input.Msg); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusCreated)
}

// Метод для извлечения случайной строки из базы данных
func (dp *DatabaseProvider) SelectHello() (string, error) {
	var msg string
	row := dp.db.QueryRow("SELECT message FROM hello ORDER BY RANDOM() LIMIT 1")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}
	return msg, nil
}

// Метод для вставки данных в таблицу
func (dp *DatabaseProvider) InsertHello(msg string) error {
	_, err := dp.db.Exec("INSERT INTO hello (message) VALUES ($1)", msg)
	return err
}

func main() {
	address := flag.String("address", "127.0.0.1:8081", "адрес для запуска сервера")
	flag.Parse()

	// Формирование строки подключения PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Подключение к БД
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем провайдер базы данных и обработчики
	dp := DatabaseProvider{db: db}
	h := Handlers{dbProvider: dp}

	// Создаем экземпляр Echo
	e := echo.New()

	// Регистрируем маршруты
	e.GET("/get", h.GetHello)
	e.POST("/post", h.PostHello)

	// Запускаем сервер Echo
	fmt.Println("Starting server on:", *address)
	if err := e.Start(*address); err != nil {
		log.Fatal(err)
	}
}
