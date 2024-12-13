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
	dbname   = "query"
)

// ▎Структура Handlers для работы с запросами через Echo
type Handlers struct {
	dbProvider DatabaseProvider
}

// ▎Провайдер базы данных
type DatabaseProvider struct {
	db *sql.DB
}

// ▎Метод для обработки GET-запроса
func (h *Handlers) GreetGet(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		return c.String(http.StatusOK, "Hello, stranger!")
	}

	// Проверяем наличие имени в базе
	exists, err := h.dbProvider.SelectHello(name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if !exists {
		return c.String(http.StatusOK, "Such user does not exist!")
	}
	return c.String(http.StatusOK, "Hello, "+name+"!")
}

// ▎Метод для обработки POST-запроса
func (h *Handlers) GreetPost(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		return c.String(http.StatusOK, "Hello, stranger!")
	}

	// Добавляем имя в базу данных
	err := h.dbProvider.InsertHello(name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusCreated)
}

// ▎Методы для работы с базой данных
func (dp *DatabaseProvider) SelectHello(name string) (bool, error) {
	var exists string
	query := `SELECT name_user FROM usernames WHERE name_user = $1`
	err := dp.db.QueryRow(query, name).Scan(&exists)

	// Проверяем, если пользователь не найден
	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (dp *DatabaseProvider) InsertHello(name string) error {
	_, err := dp.db.Exec("INSERT INTO usernames (name_user) VALUES ($1)", name)
	return err
}

// ▎Основная функция
func main() {
	// Адрес сервера
	address := flag.String("address", "127.0.0.1:8081", "адрес для запуска сервера")
	flag.Parse()

	// Формирование строки подключения для PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Подключение к базе данных
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем провайдер базы данных и обработчики
	dbProvider := DatabaseProvider{db: db}
	handlers := Handlers{dbProvider: dbProvider}

	// Инициализируем Echo
	e := echo.New()

	// ▎Регистрируем маршруты
	e.GET("/api/user/get", handlers.GreetGet)
	e.POST("/api/user/post", handlers.GreetPost)

	// Запуск сервера
	fmt.Println("Starting server on:", *address)
	log.Fatal(e.Start(*address))
}
