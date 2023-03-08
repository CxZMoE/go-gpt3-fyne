package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type PromptEntry struct {
	widget.Entry
	chat  *ChatContext
	rcbox *widget.Entry
	alt   bool
}

func NewPromptEntry(chat *ChatContext, rcbox *widget.Entry) *PromptEntry {
	e := &PromptEntry{
		chat:  chat,
		rcbox: rcbox,
		alt:   false,
	}
	e.MultiLine = true
	e.Wrapping = fyne.TextTruncate
	e.ExtendBaseWidget(e)

	return e
}

func (e *PromptEntry) KeyDown(key *fyne.KeyEvent) {
	if e.Disabled() {
		return
	}

	if key.Name == desktop.KeyAltLeft || key.Name == desktop.KeyAltRight {
		e.alt = true
	}

	if e.alt && key.Name == fyne.KeyReturn {
		log.Println("Send")
		e.Disable()

		msg, err := e.chat.Send(e.Text, e.rcbox)
		if err != nil {
			log.Println(err)
		}
		if !e.chat.Options.IsStreamming {
			e.chat.content += msg.Content + "\n"
			e.rcbox.Text = e.chat.content
			e.rcbox.CursorRow = 65536
			e.rcbox.Refresh()
		}
		e.Text = ""
		e.Enable()
		e.FocusGained()

	}
}

func (e *PromptEntry) KeyUp(key *fyne.KeyEvent) {
	if e.Disabled() {
		return
	}

	if key.Name == desktop.KeyAltLeft || key.Name == desktop.KeyAltRight {
		e.alt = false
	}
}

// RunUI start fyne ui
func RunUI(chat *ChatContext) {
	fyneApp := app.New()
	win := fyneApp.NewWindow("GOGPT Fyne 3.5")

	// vbox := layout.NewVBoxLayout()
	lay := layout.NewGridLayoutWithRows(2)

	// rcboxContent := ""
	rcbox := widget.NewMultiLineEntry()

	promptInput := NewPromptEntry(chat, rcbox)
	promptInput.SetPlaceHolder("message")

	// btn := widget.NewButton("SUBMIT", func() {
	// 	chat.Send(promptInput.Text, rcbox)
	// 	promptInput.Text = ""
	// 	promptInput.Refresh()
	// })

	c := container.New(lay, rcbox, promptInput)
	win.Resize(fyne.NewSize(1280, 720))

	promptInput.FocusGained()
	// win.SetFixedSize(true)

	win.CenterOnScreen()
	for _, msg := range chat.History {
		if msg.Role == CHAT_ROLE_USER {
			chat.content += fmt.Sprintf("==> %s\n\n", msg.Content)
		}
		if msg.Role == CHAT_ROLE_ASSIST {
			chat.content += fmt.Sprintf("%s\n\n", msg.Content)
		}
		rcbox.Text = chat.content
		rcbox.CursorRow = 65536
		rcbox.Refresh()
	}

	chat.content += "\n\n========== History ==========\n\n"
	rcbox.Text = chat.content
	rcbox.CursorRow = 65536
	rcbox.Refresh()
	win.SetContent(c)
	win.ShowAndRun()
}
