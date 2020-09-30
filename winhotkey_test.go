package winhotkey

import (
	"errors"
	"fmt"
	"testing"
	"unsafe"
)

func TestHotkey_String(t *testing.T) {
	tests := []struct {
		name string
		h    *Hotkey
		want string
	}{
		{
			name: "CTRL+ALT+A",
			h: &Hotkey{
				ID:        1,
				Modifiers: ModCtrl + ModAlt,
				KeyCode:   'A',
			},
			want: "Hotkey[ID: 1, Ctrl+Alt+A]",
		},
		{
			name: "CTRL+SHIFT+A",
			h: &Hotkey{
				ID:        2,
				Modifiers: ModCtrl + ModShift,
				KeyCode:   'A',
			},
			want: "Hotkey[ID: 2, Ctrl+Shift+A]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.h.String(); got != tt.want {
				t.Errorf("Hotkey.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegisterHotkey(t *testing.T) {
	type args struct {
		h *Hotkey
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "CTRL+ALT+A",
			args: args{h: &Hotkey{
				ID:        1,
				Modifiers: ModCtrl + ModAlt + ModNoRepeat,
				KeyCode:   'A'},
			},
			wantErr: false,
		},
		{
			name: "CTRL+WIN+A",
			args: args{h: &Hotkey{
				ID:        3,
				Modifiers: ModCtrl + ModWin,
				KeyCode:   'A'},
			},
			wantErr: false,
		},
		{
			name: "CTRL+WIN+A",
			args: args{h: &Hotkey{
				ID:        3,
				Modifiers: ModCtrl + ModWin,
				KeyCode:   'A'},
			},
			wantErr: true,
		},
		{
			name: "CTRL+WIN+A",
			args: args{h: &Hotkey{
				ID:        4,
				Modifiers: ModCtrl + ModWin,
				KeyCode:   'A'},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RegisterHotkey(tt.args.h); (err != nil) != tt.wantErr {
				t.Errorf("RegisterHotkey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type testget struct {
	t     *testing.T
	c     chan *Hotkey
	count int
}

func (tg *testget) Call(a ...uintptr) (r1 uintptr, r2 uintptr, lastErr error) {
	tg.count++
	msg := (*MSG)(unsafe.Pointer(a[0]))
	msg.WPARAM = 1
	if tg.count >= 4 {
		lastErr = errors.New("test error")
	}
	return 0, 0, lastErr
}
func TestRun(t *testing.T) {
	c := make(chan *Hotkey, 10)
	getmsg = &testget{
		t: t,
		c: c,
	}

	keys = map[int16]*Hotkey{
		1: {
			ID:           1,
			Modifiers:    ModAlt + ModCtrl + ModNoRepeat,
			KeyCode:      'A',
			CallbackChan: c,
			CallbackFunc: func(h *Hotkey) {
				fmt.Println("func", h.String())
			},
		},
	}

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "test run",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Run(); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	close(c)

	for h := range c {
		fmt.Println("chan", h.String())
	}
}
