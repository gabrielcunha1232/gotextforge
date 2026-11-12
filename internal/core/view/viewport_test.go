package view

import "testing"

func TestViewportOperations(t *testing.T) {
	tests := []struct {
		name       string
		initTop    int
		initHeight int
		totalLines int
		action     func(v *Viewport, total int)
		wantTop    int
		wantStart  int
		wantEnd    int
	}{
		{
			name:       "rolar para cima estando no topo trava em 0",
			initTop:    0,
			initHeight: 10,
			totalLines: 50,
			action:     func(v *Viewport, total int) { v.Scroll(-5, total) },
			wantTop:    0,
			wantStart:  0,
			wantEnd:    10,
		},
		{
			name:       "rolar para baixo normalmente",
			initTop:    0,
			initHeight: 10,
			totalLines: 50,
			action:     func(v *Viewport, total int) { v.Scroll(15, total) },
			wantTop:    15,
			wantStart:  15,
			wantEnd:    25,
		},
		{
			name:       "rolar alem do fim trava com ultima linha na base",
			initTop:    35,
			initHeight: 10,
			totalLines: 50,
			action:     func(v *Viewport, total int) { v.Scroll(100, total) },
			wantTop:    40,
			wantStart:  40,
			wantEnd:    50,
		},
		{
			name:       "arquivo menor que a tela nao rola",
			initTop:    0,
			initHeight: 30,
			totalLines: 10,
			action:     func(v *Viewport, total int) { v.Scroll(5, total) },
			wantTop:    0,
			wantStart:  0,
			wantEnd:    10,
		},
		{
			name:       "ir para o fim (ScrollBottom)",
			initTop:    0,
			initHeight: 20,
			totalLines: 100,
			action:     func(v *Viewport, total int) { v.ScrollBottom(total) },
			wantTop:    80,
			wantStart:  80,
			wantEnd:    100,
		},
		{
			name:       "ir para o topo (ScrollTop)",
			initTop:    45,
			initHeight: 20,
			totalLines: 100,
			action:     func(v *Viewport, total int) { v.ScrollTop() },
			wantTop:    0,
			wantStart:  0,
			wantEnd:    20,
		},
		{
			name:       "resize maior no fim do arquivo puxa Top para cima",
			initTop:    80,
			initHeight: 20,
			totalLines: 100,
			action:     func(v *Viewport, total int) { v.Resize(30, total) },
			wantTop:    70,
			wantStart:  70,
			wantEnd:    100,
		},
		{
			name:       "resize menor mantem Top valido",
			initTop:    50,
			initHeight: 30,
			totalLines: 100,
			action:     func(v *Viewport, total int) { v.Resize(15, total) },
			wantTop:    50,
			wantStart:  50,
			wantEnd:    65,
		},
		{
			name:       "resize maior que o arquivo inteiro zera Top",
			initTop:    10,
			initHeight: 15,
			totalLines: 25,
			action:     func(v *Viewport, total int) { v.Resize(40, total) },
			wantTop:    0,
			wantStart:  0,
			wantEnd:    25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := Viewport{Top: tt.initTop, Height: tt.initHeight}
			tt.action(&vp, tt.totalLines)

			if vp.Top != tt.wantTop {
				t.Errorf("vp.Top = %d, esperava %d", vp.Top, tt.wantTop)
			}

			gotStart, gotEnd := vp.VisibleRange(tt.totalLines)
			if gotStart != tt.wantStart || gotEnd != tt.wantEnd {
				t.Errorf("VisibleRange(%d) = (%d, %d), esperava (%d, %d)",
					tt.totalLines, gotStart, gotEnd, tt.wantStart, tt.wantEnd)
			}
		})
	}
}
