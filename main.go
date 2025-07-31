package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/lib/pq"
)

type Usuarios struct {
	Nome  string `json:"nome"`
	Score int    `json:"score"`
}

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("Variável DATABASE_URL não definida")
	}
	db, err := sql.Open("postgres", connStr)
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("erro no ping: ", err)
	}

	fmt.Println("conectado ao banco de dados")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Content-Type",
	}))

	app.Post("/enviarUsers", func(c *fiber.Ctx) error {
		var usuarios Usuarios
		if err := c.BodyParser(&usuarios); err != nil {
			fmt.Println("erro ao enviar o json: ", err)
			return c.Status(400).JSON(fiber.Map{"erro": "JSON inválido"})
		}

		query := `INSERT INTO usuarios (nome, score) VALUES ($1, $2)`
		_, err := db.Exec(query, usuarios.Nome, usuarios.Score)
		if err != nil {
			log.Println("erro ao tentar adicionar ao banco de dados:", err)
			return c.Status(500).JSON(fiber.Map{"erro": "Erro ao inserir"})
		}

		return c.JSON(fiber.Map{
			"nome":  usuarios.Nome,
			"score": usuarios.Score,
		})
	})

	app.Post("/mostrarUsuarios", func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT nome, score FROM usuarios ORDER BY score DESC")
		if err != nil {
			fmt.Println("Erro ao tentar selecionar os usuários:", err)
			return c.Status(500).JSON(fiber.Map{"erro": "Erro ao buscar usuários"})
		}
		defer rows.Close()

		var users []map[string]interface{}
		for rows.Next() {
			var nome string
			var score int

			if err := rows.Scan(&nome, &score); err != nil {
				fmt.Println("Erro ao escanear linha:", err)
				return c.Status(500).JSON(fiber.Map{"erro": "Erro ao processar dados"})
			}

			users = append(users, map[string]interface{}{"nome": nome, "score": score})
		}

		return c.JSON(users)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}
