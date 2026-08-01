package confvision

import (
	"confvision/src/config"
	"os"
	"strings"
	"testing"
)

func TestChaveRtmpHashids(t *testing.T) {
	os.Setenv("RTMP_PUBLISH_SECRET", "teste-salt-confvision")
	config.RtmpPublishSecret = "teste-salt-confvision"

	chave := ChaveRtmp(5)
	if len(chave) < 12 {
		t.Fatalf("len=%d chave=%q", len(chave), chave)
	}
	for _, c := range chave {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z')) {
			t.Fatalf("char invalido %q em %q", c, chave)
		}
	}
	id, ok := ParseChaveRtmp(chave)
	if !ok || id != 5 {
		t.Fatalf("parse want 5 got %d ok=%v", id, ok)
	}
	path := StreamPath(5)
	if path != "cam/"+chave {
		t.Fatalf("StreamPath want cam/%s got %q", chave, path)
	}
	id2, ok2 := ParseChaveRtmp(path)
	if !ok2 || id2 != 5 {
		t.Fatalf("parse path want 5 got %d ok=%v", id2, ok2)
	}
	url := MontarRtmpPublishURL("rtmp://rtmp.dnsid.com.br:1935", 5, false)
	if !strings.HasSuffix(url, "/cam/"+chave) {
		t.Fatalf("url=%q", url)
	}
	t.Logf("chave=%s path=%s url=%s", chave, path, url)
}

func TestChaveRtmpSemSecret(t *testing.T) {
	config.RtmpPublishSecret = ""
	if ChaveRtmp(5) != "" {
		t.Fatal("esperava vazio sem secret")
	}
}
