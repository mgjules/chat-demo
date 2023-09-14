// Code generated by templ@v0.2.334 DO NOT EDIT.

package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import (
	"strconv"
	"time"

	"github.com/mgjules/chat-demo/chat"
	"github.com/mgjules/chat-demo/user"
)

func ChatHeaderNumUsers(numUsers uint64) templ.Component {
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
		_, err = templBuffer.WriteString("<div id=\"online\" class=\"text-xs text-coolgray-400\" hx-swap-oob=\"true\">")
		if err != nil {
			return err
		}
		var var_2 string = strconv.Itoa(int(numUsers))
		_, err = templBuffer.WriteString(templ.EscapeString(var_2))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(" ")
		if err != nil {
			return err
		}
		var_3 := `user`
		_, err = templBuffer.WriteString(var_3)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(" ")
		if err != nil {
			return err
		}
		if numUsers > 1 {
			var_4 := `s`
			_, err = templBuffer.WriteString(var_4)
			if err != nil {
				return err
			}
		}
		_, err = templBuffer.WriteString("</div>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = templBuffer.WriteTo(w)
		}
		return err
	})
}

func ChatHeader(numUsers uint64, userName string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_5 := templ.GetChildren(ctx)
		if var_5 == nil {
			var_5 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, err = templBuffer.WriteString("<div class=\"flex-none flex justify-between items-center flex-wrap gap-4\"><div><div class=\"uppercase\"><span class=\"font-extralight\">")
		if err != nil {
			return err
		}
		var_6 := `Chatroom`
		_, err = templBuffer.WriteString(var_6)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</span>")
		if err != nil {
			return err
		}
		var_7 := `Demo`
		_, err = templBuffer.WriteString(var_7)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</div>")
		if err != nil {
			return err
		}
		err = ChatHeaderNumUsers(numUsers).Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</div><div class=\"text-lightblue-200 text-sm\">")
		if err != nil {
			return err
		}
		var var_8 string = userName
		_, err = templBuffer.WriteString(templ.EscapeString(var_8))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</div></div>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = templBuffer.WriteTo(w)
		}
		return err
	})
}

func ChatMessageWrapped(user *user.User, message *chat.Message) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_9 := templ.GetChildren(ctx)
		if var_9 == nil {
			var_9 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, err = templBuffer.WriteString("<div hx-swap-oob=\"beforebegin:#messages&gt;li:last-child\">")
		if err != nil {
			return err
		}
		err = ChatMessage(user, message).Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</div>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = templBuffer.WriteTo(w)
		}
		return err
	})
}

func ChatMessage(user *user.User, message *chat.Message) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_10 := templ.GetChildren(ctx)
		if var_10 == nil {
			var_10 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		var var_11 = []any{templ.KV("flex justify-end", user.ID == message.User.ID), "overflow-anchor-none"}
		err = templ.RenderCSSItems(ctx, templBuffer, var_11...)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("<li class=\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(templ.EscapeString(templ.CSSClasses(var_11).String()))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"><div class=\"w-fit flex flex-col px-3 py-2 mr-4 text-xs bg-coolgray-700 border-t-1 border-t-coolgray-500 border-t-opacity-50 shadow-sm bg-opacity-50 rounded-md\">")
		if err != nil {
			return err
		}
		if user.ID != message.User.ID {
			_, err = templBuffer.WriteString("<div class=\"font-semibold\">")
			if err != nil {
				return err
			}
			var var_12 string = message.User.Name
			_, err = templBuffer.WriteString(templ.EscapeString(var_12))
			if err != nil {
				return err
			}
			_, err = templBuffer.WriteString("</div>")
			if err != nil {
				return err
			}
		}
		var var_13 = []any{templ.KV("mt-1", user.ID != message.User.ID), "flex flex-justify-between gap-2"}
		err = templ.RenderCSSItems(ctx, templBuffer, var_13...)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("<div class=\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(templ.EscapeString(templ.CSSClasses(var_13).String()))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"><div class=\"flex-nowrap font-light break-words\">")
		if err != nil {
			return err
		}
		var var_14 string = message.Content
		_, err = templBuffer.WriteString(templ.EscapeString(var_14))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</div><div class=\"timeago self-end shrink-0 mt-1 text-[0.65rem] line-height-[0.80rem] font-light text-coolgray-400\" datetime=\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(templ.EscapeString(message.Time.String()))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"></div></div></div></li>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = templBuffer.WriteTo(w)
		}
		return err
	})
}

func ChatMessages(user *user.User, messages []*chat.Message) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_15 := templ.GetChildren(ctx)
		if var_15 == nil {
			var_15 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, err = templBuffer.WriteString("<ul id=\"messages\" class=\"flex-initial grow mt-4 space-y-2 overflow-y-scroll transition-all\" hx-on::load=\"applyTimeago()\">")
		if err != nil {
			return err
		}
		for _, msg := range messages {
			err = ChatMessage(user, msg).Render(ctx, templBuffer)
			if err != nil {
				return err
			}
		}
		_, err = templBuffer.WriteString("<li class=\"overflow-anchor-auto h-0.5\"></li></ul>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = templBuffer.WriteTo(w)
		}
		return err
	})
}

func ChatForm(disabled bool) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_16 := templ.GetChildren(ctx)
		if var_16 == nil {
			var_16 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, err = templBuffer.WriteString("<form id=\"form\" hx-swap-oob=\"true\" class=\"flex-none mt-4\" ws-send hx-on::load=\"this.querySelector(&#39;input[name=chat_message]&#39;).focus()\"><input name=\"chat_message\" type=\"text\" placeholder=\"Type here\"")
		if err != nil {
			return err
		}
		if disabled {
			_, err = templBuffer.WriteString(" disabled")
			if err != nil {
				return err
			}
		}
		_, err = templBuffer.WriteString(" maxlength=\"256\" required class=\"w-full px-3 py-2 text-sm bg-coolgray-700 bg-opacity-70 border-1 border-coolgray-600 outline-none ring-0 focus:ring-1 focus:ring-coolgray-600 transition-all disabled:opacity-40 disabled:cursor-not-allowed rounded-md\"></form>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = templBuffer.WriteTo(w)
		}
		return err
	})
}

func ChatError(cErr string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_17 := templ.GetChildren(ctx)
		if var_17 == nil {
			var_17 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, err = templBuffer.WriteString("<div id=\"error\" hx-swap-oob=\"true\">")
		if err != nil {
			return err
		}
		if cErr != "" {
			_, err = templBuffer.WriteString("<div class=\"flex-none text-red mt-2 text-xs capitalize\">")
			if err != nil {
				return err
			}
			var var_18 string = cErr
			_, err = templBuffer.WriteString(templ.EscapeString(var_18))
			if err != nil {
				return err
			}
			_, err = templBuffer.WriteString("</div>")
			if err != nil {
				return err
			}
		}
		_, err = templBuffer.WriteString("</div>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = templBuffer.WriteTo(w)
		}
		return err
	})
}

func ChatFooter() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_19 := templ.GetChildren(ctx)
		if var_19 == nil {
			var_19 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, err = templBuffer.WriteString("<div class=\"flex-none mt-4 text-xs text-center text-coolgray-400\">")
		if err != nil {
			return err
		}
		var_20 := `Copyright (c) `
		_, err = templBuffer.WriteString(var_20)
		if err != nil {
			return err
		}
		var var_21 string = time.Now().Format("2006")
		_, err = templBuffer.WriteString(templ.EscapeString(var_21))
		if err != nil {
			return err
		}
		var_22 := `. All rights reserved.`
		_, err = templBuffer.WriteString(var_22)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</div>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = templBuffer.WriteTo(w)
		}
		return err
	})
}

func Chat(user *user.User, room *chat.Room, disabled bool, cErr string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_23 := templ.GetChildren(ctx)
		if var_23 == nil {
			var_23 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, err = templBuffer.WriteString("<script src=\"https://unpkg.com/htmx.org/dist/ext/ws.js\">")
		if err != nil {
			return err
		}
		var_24 := ``
		_, err = templBuffer.WriteString(var_24)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</script><script src=\"https://unpkg.com/timeago.js@4.0.2/dist/timeago.min.js\">")
		if err != nil {
			return err
		}
		var_25 := ``
		_, err = templBuffer.WriteString(var_25)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</script><script type=\"text/javascript\">")
		if err != nil {
			return err
		}
		var_26 := `
		function applyTimeago() {
			// Select only element that has not been processed by timeago yet.
			const els = document.querySelectorAll('.timeago:not([timeago-id])')
			timeago.render(els, 'mini-locale', { minInterval: 10 })
		}

		window.onload = () => {
			// The defaults locales are too verbose.
			timeago.register('mini-locale', (number, index, totalSec) => {
				return [
					['now', 'soon'],
					['%ss', 'in %ss'],
					['1m', 'in 1m'],
					['%sm', 'in %sm'],
					['1h', 'in 1h'],
					['%sh', 'in %sh'],
					['1d', 'in 1d'],
					['%sd', 'in %sd'],
					['1w', 'in 1w'],
					['%sw', 'in %sw'],
					['1mo', 'in 1mo'],
					['%smo', 'in %smo'],
					['1yr', 'in 1yr'],
					['%syr', 'in %syr']
				][index]
			})
			applyTimeago()

			// Check if UnoCSS is loaded by watching the removal of the ` + "`" + `un-cloak` + "`" + ` attribute from the body.
			// It's a vanilla alternative to ` + "`" + `jQuery.ready` + "`" + `.
			const observer = new MutationObserver((mutationList) => {
				mutationList.forEach((mutation) => {
					switch (mutation.type) {
						case "attributes":
							switch (mutation.attributeName) {
								case "un-cloak":
									// Safe to assume that UnoCSS has loaded.
									// Scroll to last message and focus the form input.
									document.querySelector('#messages>li:last-child').scrollIntoView({
										behavior: "smooth",
										block: "end",
										inline: "nearest"
									})
									document.querySelector('#form>input[name=chat_message]').focus()
									// no need to keep observing since it's a one-off operation.
									observer.disconnect()
							}
							break
					}
				})
			})
			observer.observe(document.body, {
				attributeFilter: ["un-cloak"]
			})
		}
	`
		_, err = templBuffer.WriteString(var_26)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</script><div hx-ext=\"ws\" ws-connect=\"/chatroom\" class=\"flex flex-col p-4 container mx-auto max-h-screen\">")
		if err != nil {
			return err
		}
		err = ChatHeader(room.NumUsers(), user.Name).Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		err = ChatMessages(user, room.Messages()).Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		err = ChatForm(disabled).Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		err = ChatError(cErr).Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		err = ChatFooter().Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</div>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = templBuffer.WriteTo(w)
		}
		return err
	})
}
