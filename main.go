package main

import (
	"aad-auth-poc/views"
	"bytes"
	"log"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/azureadv2"
	"github.com/shareed2k/goth_fiber"
)

func main() {
	app := setupFiberApp()
	sessions := setupSessionStorage()

	configureOAuthProviders()
	setupRoutes(app, sessions)

	startServer(app)
}

func setupFiberApp() *fiber.App {
	app := fiber.New()
	app.Use(logger.New())
	return app
}

func setupSessionStorage() *session.Store {
	storage := sqlite3.New()
	sessions := session.New(session.Config{
		Storage: storage,
	})
	goth_fiber.SessionStore = sessions
	return sessions
}

func configureOAuthProviders() {
	goth.UseProviders(
		azureadv2.New(
			os.Getenv("AUTH_TARGET_APPLICATION_ID"),
			os.Getenv("AUTH_CLIENT_SECRET"),
			"http://localhost:8088/auth/callback/azureadv2",
			azureadv2.ProviderOptions{
				Scopes: []azureadv2.ScopeType{azureadv2.OpenIDScope},
				Tenant: azureadv2.OrganizationsTenant,
			},
		),
	)
}

func setupRoutes(app *fiber.App, sessions *session.Store) {
	app.Get("/auth/callback/:provider", handleAuthCallback(sessions))
	app.Get("/login/:provider", handleLogin(sessions))
	app.Get("/logout", handleLogout)
	app.Get("/", handleHome(sessions))

	app.Use(func(ctx *fiber.Ctx) error {
		// Catch all for any route not defined.
		return ctx.SendStatus(http.StatusNotFound)
	})
}

func handleAuthCallback(sessions *session.Store) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		user, err := goth_fiber.CompleteUserAuth(ctx)
		if err != nil {
			return ctx.Status(http.StatusInternalServerError).SendString("Authentication failed")
		}

		sess, err := sessions.Get(ctx)
		if err != nil {
			return ctx.Status(http.StatusInternalServerError).SendString("Session retrieval failed")
		}

		storeUserInSession(sess, user)
		return ctx.Redirect("/")
	}
}

func handleLogin(sessions *session.Store) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		gothUser, err := goth_fiber.CompleteUserAuth(ctx)
		if err == nil {
			sess, err := sessions.Get(ctx)
			if err != nil {
				return ctx.Status(http.StatusInternalServerError).SendString("Session retrieval failed")
			}

			storeUserInSession(sess, gothUser)
			return ctx.Redirect("/")
		}

		goth_fiber.BeginAuthHandler(ctx)
		return nil
	}
}

func handleLogout(ctx *fiber.Ctx) error {
	if err := goth_fiber.Logout(ctx); err != nil {
		return ctx.Status(http.StatusInternalServerError).SendString("Logout failed")
	}

	return ctx.Redirect("/")
}

func handleHome(sessions *session.Store) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		sess, err := sessions.Get(ctx)
		if err != nil {
			return ctx.Status(http.StatusInternalServerError).SendString("Session retrieval failed")
		}
		defer sess.Save()

		if gothUser, ok := getUserFromSession(sess); ok {
			return renderTempl(ctx, views.HomeScreen(gothUser))
		}

		return renderTempl(ctx, views.LoginScreen())
	}
}

func startServer(app *fiber.App) {
	if err := app.Listen("localhost:8088"); err != nil {
		log.Fatal(err)
	}
}

// Helper functions below

func storeUserInSession(sess *session.Session, user goth.User) {
	sess.Set("user", user)
	if err := sess.Save(); err != nil {
		log.Println("Failed to save session:", err)
	}
}

func getUserFromSession(sess *session.Session) (goth.User, bool) {
	val := sess.Get("user")
	if val == nil {
		return goth.User{}, false
	}

	user, ok := val.(goth.User)
	return user, ok
}

func renderTempl(c *fiber.Ctx, cmpnt templ.Component) error {
	content := new(bytes.Buffer)
	err := cmpnt.Render(c.Context(), content)

	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Rendering failed")
	}

	c.Set("Content-Type", "text/html")
	return c.Send(content.Bytes())
}
