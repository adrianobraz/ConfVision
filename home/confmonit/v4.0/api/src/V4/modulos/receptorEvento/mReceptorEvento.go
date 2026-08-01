package receptorEvento

import connV4 "api/src/V4/conexao"

type ReceptorEvento struct {
	ID_ReceptorEvento string `json:"idRecptorEvento"`
	ID_Vinculo        string `json:"idVinculo"`
	Nome              string `json:"nome"`
	Porta             string `json:"porta"`
	Rota              string `json:"rota"`
	Modulo            string `json:"modulo"`
	Producao          string `json:"producao"`
	Senha             string `json:"senha"`
	Ativo             string `json:"ativo"`
}

type DadosCarregamento struct {
	Nome   string `json:"nome"`
	Porta  string `json:"porta"`
	Rota   string `json:"rota"`
	Modulo string `json:"modulo"`
	Senha  string `json:"senha"`
}

func (r *ReceptorEvento) carregarDados(lista *[]DadosCarregamento) error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT 			 
			receptorEvento.Nome, 
			receptorEvento.Porta, 
			receptorEvento.Rota, 
			receptorEvento.Modulo, 
			receptorEvento.Senha

			FROM receptorEvento
			
			WHERE receptorEvento.Ativo = 'S'
			
			AND receptorEvento.Producao = 'S'
	`)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item DadosCarregamento
		if err := tab.Scan(
			&item.Nome,
			&item.Porta,
			&item.Rota,
			&item.Modulo,
			&item.Senha,
		); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}

	return nil
}
