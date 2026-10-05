package visdata

import "strings"

type PlanoFlags struct {
	CapturaSensor     bool    `json:"captura_sensor"`
	CapturaAnalitico  bool    `json:"captura_analitico"`
	SomenteArmado     bool    `json:"somente_armado"`
	EventoGravaFoto   bool    `json:"evento_grava_foto"`
	EventoGravaVideo  bool    `json:"evento_grava_video"`
	SemAtivo          bool    `json:"sem_ativo"`
	GravaContinua     bool    `json:"grava_continua"`
	GravaMovimento    bool    `json:"grava_movimento"`
	GravaTimelapse    bool    `json:"grava_timelapse"`
	RetencaoDias      int     `json:"retencao_dias"`
	SegmentoMinutos   int     `json:"segmento_minutos"`
	Unidade           string  `json:"unidade"`
	Valor             float64 `json:"valor"`
	PlanoLabel        string  `json:"plano_label"`
}

func PlanoFlagsFrom(plano string) PlanoFlags {
	p := strings.TrimSpace(plano)
	norm := p
	switch norm {
	case "sensor":
		norm = "sensor_foto_video"
	case "analitico_armado":
		norm = "analitico_armado_foto_video"
	case "analitico_24h":
		norm = "analitico_24h_foto_video"
	}

	f := PlanoFlags{Unidade: "camera", SegmentoMinutos: 5, PlanoLabel: "Nenhum"}

	switch norm {
	case "online":
		f.SemAtivo = true
		f.Valor = 2.99
		f.PlanoLabel = "Camera online"
	case "sensor_foto":
		f.CapturaSensor, f.EventoGravaFoto, f.SemAtivo = true, true, true
		f.Valor = 7.99
		f.PlanoLabel = "Sensor foto"
	case "sensor_foto_video":
		f.CapturaSensor, f.EventoGravaFoto, f.EventoGravaVideo, f.SemAtivo = true, true, true, true
		f.Valor = 9.99
		f.PlanoLabel = "Sensor foto + video"
	case "analitico_armado_evento":
		f.CapturaAnalitico, f.SomenteArmado = true, true
		f.Valor = 11.99
		f.PlanoLabel = "Analitico armado — so evento"
	case "analitico_armado_foto":
		f.CapturaAnalitico, f.SomenteArmado, f.EventoGravaFoto = true, true, true
		f.Valor = 13.99
		f.PlanoLabel = "Analitico armado — foto"
	case "analitico_armado_foto_video":
		f.CapturaAnalitico, f.SomenteArmado, f.EventoGravaFoto, f.EventoGravaVideo = true, true, true, true
		f.Valor = 14.99
		f.PlanoLabel = "Analitico armado — foto + video"
	case "analitico_24h_evento":
		f.CapturaAnalitico = true
		f.Valor = 16.99
		f.PlanoLabel = "Analitico 24h — so evento"
	case "analitico_24h_foto":
		f.CapturaAnalitico, f.EventoGravaFoto = true, true
		f.Valor = 18.99
		f.PlanoLabel = "Analitico 24h — foto"
	case "analitico_24h_foto_video":
		f.CapturaAnalitico, f.EventoGravaFoto, f.EventoGravaVideo = true, true, true
		f.Valor = 19.99
		f.PlanoLabel = "Analitico 24h — foto + video"
	case "gravacao_7d":
		f.Unidade, f.GravaContinua, f.RetencaoDias = "gravacao", true, 7
		f.Valor = 12.99
		f.PlanoLabel = "Gravacao continua 7 dias"
	case "gravacao_15d":
		f.Unidade, f.GravaContinua, f.RetencaoDias = "gravacao", true, 15
		f.Valor = 17.99
		f.PlanoLabel = "Gravacao continua 15 dias"
	case "gravacao_30d":
		f.Unidade, f.GravaContinua, f.RetencaoDias = "gravacao", true, 30
		f.Valor = 24.99
		f.PlanoLabel = "Gravacao continua 30 dias"
	case "gravacao_movimento_7d":
		f.Unidade, f.GravaMovimento, f.RetencaoDias = "gravacao", true, 7
		f.Valor = 9.99
		f.PlanoLabel = "Gravacao por movimento 7 dias"
	case "gravacao_movimento_15d":
		f.Unidade, f.GravaMovimento, f.RetencaoDias = "gravacao", true, 15
		f.Valor = 13.99
		f.PlanoLabel = "Gravacao por movimento 15 dias"
	case "gravacao_movimento_30d":
		f.Unidade, f.GravaMovimento, f.RetencaoDias = "gravacao", true, 30
		f.Valor = 19.99
		f.PlanoLabel = "Gravacao por movimento 30 dias"
	case "gravacao_timelapse_7d":
		f.Unidade, f.GravaTimelapse, f.RetencaoDias, f.SegmentoMinutos = "gravacao", true, 7, 3
		f.Valor = 6.99
		f.PlanoLabel = "Gravacao timelapse inteligente 7 dias"
	case "gravacao_timelapse_15d":
		f.Unidade, f.GravaTimelapse, f.RetencaoDias, f.SegmentoMinutos = "gravacao", true, 15, 3
		f.Valor = 9.99
		f.PlanoLabel = "Gravacao timelapse inteligente 15 dias"
	case "gravacao_timelapse_30d":
		f.Unidade, f.GravaTimelapse, f.RetencaoDias, f.SegmentoMinutos = "gravacao", true, 30, 3
		f.Valor = 13.99
		f.PlanoLabel = "Gravacao timelapse inteligente 30 dias"
	}
	return f
}

func (f PlanoFlags) Map() map[string]any {
	return map[string]any{
		"captura_sensor":     f.CapturaSensor,
		"captura_analitico":  f.CapturaAnalitico,
		"somente_armado":     f.SomenteArmado,
		"evento_grava_foto":  f.EventoGravaFoto,
		"evento_grava_video": f.EventoGravaVideo,
		"sem_ativo":          f.SemAtivo,
		"grava_continua":     f.GravaContinua,
		"grava_movimento":    f.GravaMovimento,
		"grava_timelapse":    f.GravaTimelapse,
		"retencao_dias":      f.RetencaoDias,
		"segmento_minutos":   f.SegmentoMinutos,
		"unidade":            f.Unidade,
		"valor":              f.Valor,
		"plano_label":        f.PlanoLabel,
	}
}
