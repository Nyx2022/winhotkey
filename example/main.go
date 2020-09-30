package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/roffe/winhotkey"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	hotkeyChan := make(chan *winhotkey.Hotkey)

	hotkeys := []*winhotkey.Hotkey{
		{ // Register CTRL+SHIFT+A with NoRepat for CallbackChan
			ID:           1,
			Modifiers:    winhotkey.ModCtrl + winhotkey.ModShift + winhotkey.ModNoRepeat,
			KeyCode:      'A',
			CallbackChan: hotkeyChan,
		},
		{ // Register CTRL+ALT+A for CallbackFunc
			ID:           2,
			Modifiers:    winhotkey.ModCtrl + winhotkey.ModAlt,
			KeyCode:      'A',
			CallbackFunc: handleCallback,
		},
		{ // Register CTRL+WIN+A for CallbackFunc and CallbackChan
			ID:           3,
			Modifiers:    winhotkey.ModCtrl + winhotkey.ModWin,
			KeyCode:      'A',
			CallbackChan: hotkeyChan,
			CallbackFunc: handleCallback,
		},
	}
	for _, hk := range hotkeys {
		if err := winhotkey.RegisterHotkey(hk); err != nil {
			log.Fatal(err)
		}
	}

	// CTRL-C handling
	q := make(chan os.Signal)
	signal.Notify(q, os.Interrupt)

	go func() {
		for {
			select {
			case h := <-hotkeyChan:
				log.Println("CallbackChan", h.String(), "pressed")

			case sig := <-q:
				log.Println(sig)
				os.Exit(0)
				return

			}
		}
	}()

	// The main loop must run in the mainthread or else we won't be getting any messages from user32
	if err := winhotkey.Run(); err != nil {
		log.Fatal(err)
	}
}

func handleCallback(h *winhotkey.Hotkey) {
	log.Println("CallbackFunc", h.String(), "pressed")
}
