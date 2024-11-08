package internal

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/golang-jwt/jwt/v5"
)

type Responder struct {
	fiber.Ctx
	DB Database
}

// Otherwise json, graphql or something.
func (r Responder) IsHTMX() bool {
	return r.Ctx.Get("HX-Request", "") != ""
}

// Call it instead of Redirect().To().
func (r Responder) HTMXRedirect(to string) {
	r.Set("HX-Redirect", to)
}

func (r Responder) HTMXCurrentURL() string {
	return r.Get("HX-Current-URL")
}

// Render a page using a template.
func (r Responder) RenderPage(templatePath string, bind fiber.Map, layouts ...string) error {
	return r.Render(templatePath, r.PageMap(bind), layouts...)
}

// This type describes ALL values in EVERY partial, which can be passed into ./templates/partials
// and used by htmx requests to replace DOM, using template generation through get requests
//
// EXAMPLE:
//
//	<div hx-get="/partials/chat?mode=compact">
//
// NOTE: wont move this to internal/htmx.go
// since its only for the RenderTemplate
type HTMXPartialQuery struct {
	Id           string `query:"id"`
	Message      string `query:"message"`
	OpenSettings bool   `query:"open-settings"`
	OpenRegister bool   `query:"open-register"`
	OpenLogin    bool   `query:"open-login"`
	User         User   `query:"user"` // its safe
}

func (r *Responder) PageMap(bind fiber.Map) fiber.Map {
	user, errToken := r.GetOwner()
	if errToken != nil {
		log.Error(errToken)
		r.Cookie(&fiber.Cookie{
			Name:    "Authorization",
			Value:   "",
			Expires: time.Now(),
		})
	}
	result := fiber.Map{}
	if errToken != nil {
		result["TokenError"] = true
		result["Message"] = "Authorization error"
		result["Id"] = "local-token-error"
	} else {
		result["User"] = user
	}

	for k, v := range bind {
		result[k] = v
	}
	return result
}

// Renders a template, requested by a client.
func (r Responder) RenderTemplate() error {
	q := new(HTMXPartialQuery)
	err := r.Bind().Query(q)
	r.GetOwner()
	if err != nil {
		return err
	}
	return r.Render(r.Path()[1:], r.PageMap(fiber.Map{}))
}

// Renders the danger message html element.
func (r Responder) RenderDanger(message, id string) error {
	return r.Render("partials/danger", fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

// Renders the warning message html element.
func (r Responder) RenderWarning(message, id string) error {
	return r.Render("partials/warning", fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

// Renders the success message html element.
func (r Responder) RenderSuccess(message, id string) error {
	return r.Render("partials/success", fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

// Uses the current request information.
func (r Responder) UserRegister() error {
	id := "auth-error"
	req := new(RegisterRequest)
	err := r.Ctx.Bind().Form(req)
	if err != nil {
		log.Error(err)
		message := "Invalid request payload"
		return r.RenderWarning(message, id)
	}

	if req.IsBadPasswordMatch() {
		message := "Passwords do not match"
		return r.RenderWarning(message, id)
	}

	if req.IsMissing() {
		message := "Missing required fields"
		return r.RenderWarning(message, id)
	}

	user, err := req.CreateUser()
	if err != nil {
		log.Error(err)
		message := "Unable to register user"
		return r.RenderWarning(message, id)
	}

	return r.GiveToken(id, *user)
}

// Uses the current request information.
func (r Responder) UserLogin() error {
	id := "auth-error"
	req := new(LoginRequest)
	err := r.Ctx.Bind().Form(req)
	if err != nil {
		log.Error(err)
		message := "Invalid request payload"
		return r.RenderWarning(message, id)
	}

	if req.IsBad() {
		message := "Missing email or password"
		return r.RenderWarning(message, id)
	}

	user, err := r.DB.UserByEmail(req.Email)
	if err != nil {
		log.Error(err)
		message := "User not found"
		return r.RenderWarning(message, id)
	}

	if !CheckPassword(user.Password, req.Password) {
		message := "Invalid email or password"
		return r.RenderWarning(message, id)
	}

	r.HTMXRedirect(r.HTMXCurrentURL())
	return r.GiveToken(id, *user)
}

// Uses the current request information.
func (r Responder) UserLogout() error {
	r.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "",
		Expires: time.Now(),
	})

	r.HTMXRedirect(r.HTMXCurrentURL())
	return r.Render("partials/auth-logout", fiber.Map{})
}

// Authorize the user, using the current request information and new cookies.
func (r Responder) GiveToken(errorElementId string, user User) error {
	token, err := user.GenerateToken()
	if err != nil {
		message := "Error generating token"
		return r.RenderWarning(message, errorElementId)
	}

	message := "Success! Redirecting..."
	r.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "Bearer " + token,
		Expires: time.Now().Add(tokenExpiration),
	})
	return r.Render("partials/auth-success", fiber.Map{
		"Id":      errorElementId,
		"Message": message,
		"Token":   token,
	})
}

// Get the owner of the request using the "Authorization" header.
// Returns (nil, nil), if the header is empty.
func (r Responder) GetOwner() (*User, error) {
	authHeader := r.Cookies("Authorization")
	if authHeader == "" {
		return nil, nil
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return nil, errors.New("invalid authorization format. Expected Authorization header: Bearer and the token string")
	}

	tokenString := authHeader[7:]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims := token.Claims.(jwt.MapClaims)

	userIDFloat, ok := claims["id"].(float64)
	if !ok {
		return nil, errors.New("user ID not found or is invalid")
	}
	userID := int(userIDFloat)

	user, err := r.DB.UserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
