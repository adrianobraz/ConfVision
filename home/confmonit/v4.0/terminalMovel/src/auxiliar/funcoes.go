package aux

import "database/sql"

func GetIdFranqByIdDisp(idDisp string, idFran *string) error {
	db, err := Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Consulta personalizada
	tab, err := db.Query(`
		SELECT cliente.ID_Franqueado

		FROM dispositivo 
		
		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente

		WHERE dispositivo.ID_Dispositivo = ?
	`, idDisp)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var tmp sql.NullString
		if err := tab.Scan(&tmp); err != nil {
			return err
		}
		*idFran = tmp.String
		return nil
	}

	*idFran = "NE"
	return nil
}

func VerificaBloqueado(id string, bloqueado *bool) error {
	db, err := Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Consulta personalizada
	tab, err := db.Query(`
		SELECT listaManutencao.ID_Alvo

		FROM listaManutencao 
		
		WHERE listaManutencao.ID_Alvo = ?
	`, id)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		*bloqueado = true
		return nil
	}

	*bloqueado = false
	return nil

}
