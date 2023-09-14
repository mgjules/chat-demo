// Code generated by templ@v0.2.334 DO NOT EDIT.

package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import (
	"github.com/mgjules/chat-demo/chat"
	"github.com/mgjules/chat-demo/user"
)

func Page(user *user.User, room *chat.Room, disabled bool, cErr string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_1 := templ.GetChildren(ctx)
		if var_1 == nil {
			var_1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, err = templBuffer.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge\"><title>")
		if err != nil {
			return err
		}
		var_2 := `Chat Demo`
		_, err = templBuffer.WriteString(var_2)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</title><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><script src=\"https://unpkg.com/htmx.org@1.9.5\" integrity=\"sha384-xcuj3WpfgjlKF+FXhSQFQ0ZNr39ln+hwjN3npfM9VBnUskLolQAcN80McRIVOPuO\" crossorigin=\"anonymous\">")
		if err != nil {
			return err
		}
		var_3 := ``
		_, err = templBuffer.WriteString(var_3)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</script><link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/@unocss/reset/tailwind.min.css\"><style>")
		if err != nil {
			return err
		}
		var_4 := `
				[un-cloak] {
					display: none
				}
			`
		_, err = templBuffer.WriteString(var_4)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</style><script type=\"text/javascript\">")
		if err != nil {
			return err
		}
		var_5 := `
				// UnoCSS options.
				window.__unocss = {
					rules: [
						['overflow-anchor-none', { "overflow-anchor": 'none' }],
						['overflow-anchor-auto', { "overflow-anchor": 'auto' }],
					],
				}
			`
		_, err = templBuffer.WriteString(var_5)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</script><script src=\"https://cdn.jsdelivr.net/npm/@unocss/runtime\">")
		if err != nil {
			return err
		}
		var_6 := ``
		_, err = templBuffer.WriteString(var_6)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</script></head><body un-cloak class=\"bg-coolgray-800 text-coolgray-200 scroll-smooth\">")
		if err != nil {
			return err
		}
		err = Chat(user, room, disabled, cErr).Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</body></html>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = templBuffer.WriteTo(w)
		}
		return err
	})
}