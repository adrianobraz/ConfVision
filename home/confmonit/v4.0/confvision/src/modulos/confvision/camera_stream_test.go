package confvision

import "testing"

func TestCameraPodeStream(t *testing.T) {
	cases := []struct {
		bloqueado bool
		ativo     bool
		plano     string
		want      bool
		motivo    string
	}{
		{true, true, "online", false, "camera_bloqueada"},
		{true, false, "gravacao_7d", false, "camera_bloqueada"},
		{false, false, "online", true, ""},
		{false, true, "gravacao_7d", true, ""},
		{false, false, "gravacao_7d", false, "camera_inativa"},
	}
	for _, c := range cases {
		ok, motivo := CameraPodeStream(c.bloqueado, c.ativo, c.plano)
		if ok != c.want || motivo != c.motivo {
			t.Fatalf("bloq=%v ativo=%v plano=%q => (%v,%q) want (%v,%q)",
				c.bloqueado, c.ativo, c.plano, ok, motivo, c.want, c.motivo)
		}
	}
}
