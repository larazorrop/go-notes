package main

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// --- AUTENTICAÇÃO ---

func HandleGetLoginPage(c *fiber.Ctx) error {
	return c.Render("templates/login", fiber.Map{"Title": "Login"})
}

func HandlePostLoginPage(c *fiber.Ctx) error {
	user, err := GetUserByUsername(c.FormValue("username"))

	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(c.FormValue("password"))) != nil {
		return c.Render("templates/login", fiber.Map{
			"Title": "Login",
			"Error": "Dados inválidos.",
		})
	}

	sess, _ := store.Get(c)
	sess.Set("user_id", user.ID)
	sess.Set("username", user.Username)
	sess.Save()

	return c.Redirect("/app")
}

func HandleLogout(c *fiber.Ctx) error {
	sess, _ := store.Get(c)
	sess.Destroy()
	return c.Redirect("/app/login")
}

func HandleGetRegisterPage(c *fiber.Ctx) error {
	return c.Render("templates/register", fiber.Map{
		"Title": "Registo",
	})
}

func HandlePostRegisterPage(c *fiber.Ctx) error {
	err := CreateUser(c.FormValue("username"), c.FormValue("password"))
	if err != nil {
		return c.Render("templates/register", fiber.Map{
			"Title": "Registo",
			"Error": "Erro ao criar conta (utilizador já existe?)",
		})
	}
	return c.Redirect("/app/login")
}

// --- NOTAS ---

func HandleGetNotesPage(c *fiber.Ctx) error {
	notes, err := GetNotesByUserID(c.Locals("userID").(int))
	if err != nil {
		return c.Status(500).SendString("Erro na DB")
	}

	return c.Render("templates/notes", fiber.Map{
		"Title":    "Minhas Notas",
		"Username": c.Locals("username"),
		"Notes":    notes,
	})
}

func HandleCreateNote(c *fiber.Ctx) error {
	if content := c.FormValue("content"); content != "" {
		CreateNote(c.Locals("userID").(int), content)
	}
	return c.Redirect("/app")
}

func HandleDeleteNote(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	DeleteNote(id, c.Locals("userID").(int))
	return c.Redirect("/app")
}
