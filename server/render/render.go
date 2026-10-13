package render

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

func RenderBuffer(app *fiber.App, templateName string, bind any) (bytes.Buffer, error) {
	buf := bytes.NewBuffer([]byte{})
	err := app.Config().Views.Render(buf, templateName, bind)
	if err != nil {
		log.Error(err)
		buf.WriteString(template.HTMLEscapeString(err.Error()))
	}

	return *buf, err
}

func WrapOob(swap string, message *string) string {
	msg := ""
	if message != nil {
		msg = *message
	}

	return "<div hx-swap-oob=\"" + swap + "\">" + msg + "</div>"
}

func RenderUnreadBadgeOob(unreadCount int) string {
	unreadBadgeHtml := `<span id="notification-indicator" class="badge badge-error badge-xs indicator-item hidden"></span>`
	if unreadCount > 0 {
		countStr := fmt.Sprintf("%d", unreadCount)
		if unreadCount > 99 {
			countStr = "99+"
		}
		unreadBadgeHtml = `<span id="notification-indicator" class="badge badge-error badge-xs indicator-item">` + countStr + `</span>`
	}
	return WrapOob("outerHTML:#notification-indicator", &unreadBadgeHtml)
}
