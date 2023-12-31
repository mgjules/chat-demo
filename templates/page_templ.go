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

func Page(user *user.User, room *chat.Room, cErr *chat.Error) templ.Component {
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
		_, err = templBuffer.WriteString("</title><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><style>")
		if err != nil {
			return err
		}
		var_3 := `
				[un-cloak] {
					display: none
				}
			`
		_, err = templBuffer.WriteString(var_3)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</style><script type=\"module\">")
		if err != nil {
			return err
		}
		var_4 := `
			  // UnoCSS
				import { presetWind, presetIcons } from 'https://cdn.jsdelivr.net/npm/unocss@0.55.7/+esm'
				import initUnocssRuntime from 'https://cdn.jsdelivr.net/npm/@unocss/runtime@0.55.7/+esm'
				import reset from 'https://cdn.jsdelivr.net/npm/@unocss/reset@0.55.7/tailwind-compat.css' assert { type: 'css' }

				document.adoptedStyleSheets = [reset];

				// UnoCSS default configuration.
				initUnocssRuntime({
					defaults: {
						presets: [
							presetWind(),
							presetIcons({
								cdn: 'https://esm.sh/'
							})
						],
						rules: [
							['overflow-anchor-none', { "overflow-anchor": 'none' }],
							['overflow-anchor-auto', { "overflow-anchor": 'auto' }],
						],
					}
				})
			`
		_, err = templBuffer.WriteString(var_4)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</script></head><body un-cloak class=\"bg-coolgray-800 text-coolgray-200 scroll-smooth\">")
		if err != nil {
			return err
		}
		err = Chat(user, room, cErr).Render(ctx, templBuffer)
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
