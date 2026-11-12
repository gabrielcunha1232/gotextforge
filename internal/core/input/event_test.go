package input

import "testing"

func TestEventZeroValue(t *testing.T) {
	var ev Event
	if ev.Kind != EventNone {
		t.Fatalf("esperava zero value de Event com Kind EventNone, obteve %v", ev.Kind)
	}
}

func TestKeyEqualityAndMapKey(t *testing.T) {
	k1 := Key{Rune: 'a', Mod: ModCtrl}
	k2 := Key{Rune: 'a', Mod: ModCtrl}
	k3 := Key{Rune: 'A'}

	if k1 != k2 {
		t.Fatalf("esperava que teclas idênticas fossem iguais: %+v != %+v", k1, k2)
	}
	if k1 == k3 {
		t.Fatalf("esperava que teclas diferentes não fossem iguais: %+v == %+v", k1, k3)
	}

	keymap := map[Key]string{
		k1: "action_ctrl_a",
		k3: "action_shift_a",
	}

	if got := keymap[k2]; got != "action_ctrl_a" {
		t.Errorf("falha ao buscar chave equivalente no map: obteve %q, esperava %q", got, "action_ctrl_a")
	}
}

func TestKeyString(t *testing.T) {
	tests := []struct {
		name string
		key  Key
		want string
	}{
		{
			name: "letra minuscula simples",
			key:  Key{Rune: 'd'},
			want: "d",
		},
		{
			name: "letra maiuscula",
			key:  Key{Rune: 'A'},
			want: "A",
		},
		{
			name: "caractere acentuado",
			key:  Key{Rune: 'ç'},
			want: "ç",
		},
		{
			name: "tecla Esc",
			key:  Key{Code: KeyEsc},
			want: "<Esc>",
		},
		{
			name: "tecla Enter",
			key:  Key{Code: KeyEnter},
			want: "<CR>",
		},
		{
			name: "Shift+Tab",
			key:  Key{Code: KeyTab, Mod: ModShift},
			want: "<S-Tab>",
		},
		{
			name: "Ctrl+a normalizado",
			key:  Key{Rune: 'a', Mod: ModCtrl},
			want: "<C-a>",
		},
		{
			name: "Alt+x",
			key:  Key{Rune: 'x', Mod: ModAlt},
			want: "<M-x>",
		},
		{
			name: "Alt+Shift+a",
			key:  Key{Rune: 'A', Mod: ModAlt},
			want: "<M-A>",
		},
		{
			name: "Ctrl+Alt+Up",
			key:  Key{Code: KeyUp, Mod: ModCtrl | ModAlt},
			want: "<C-M-Up>",
		},
		{
			name: "espaco",
			key:  Key{Rune: ' '},
			want: "<Space>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.key.String(); got != tt.want {
				t.Errorf("Key.String() = %q, esperava %q", got, tt.want)
			}
		})
	}
}
