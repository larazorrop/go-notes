package main

import (
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
)

//go:embed templates/*
var resources embed.FS

var store *session.Store

func main() {
	// Inicializa a Base de Dados
	if err := InitDB(); err != nil {
		log.Fatalf("Erro fatal na DB: %v", err)
	}
	defer db.Close()

	// Configura o motor de templates para usar os ficheiros embutidos na memória
	engine := html.NewFileSystem(http.FS(resources), ".html")

	store = session.New(session.Config{
		CookieHTTPOnly: true,
		// Removido o gerador de chaves fixo para segurança
	})

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use(logger.New())

	// --- ROTAS ---
	app.Get("/", func(c *fiber.Ctx) error {
		//return c.Redirect("/app/login")
		return c.Render("templates/index", nil)
	})

	api := app.Group("/app")

	// Rotas Públicas
	api.Get("/login", HandleGetLoginPage)
	api.Post("/login", HandlePostLoginPage)
	api.Get("/register", HandleGetRegisterPage)
	api.Post("/register", HandlePostRegisterPage)

	// Rotas Protegidas (Middleware inline para simplificar)
	api.Use(authMiddleware)
	api.Get("/logout", HandleLogout)
	api.Get("/", HandleGetNotesPage)
	api.Post("/notes/add", HandleCreateNote)
	api.Post("/notes/delete/:id", HandleDeleteNote)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = ":8080"
	}

	log.Printf("A servir na porta %s", port)
	log.Fatal(app.Listen(port))
}

func authMiddleware(c *fiber.Ctx) error {
	sess, err := store.Get(c)
	if err != nil || sess.Get("user_id") == nil {
		return c.Redirect("/app/login")
	}
	c.Locals("userID", sess.Get("user_id"))
	c.Locals("username", sess.Get("username"))
	return c.Next()
}
