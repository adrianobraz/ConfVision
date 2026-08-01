package S004

import (
	"fmt"
	"os"
	"path/filepath"
	"robot/src/auxiliar"
	fatCliente "robot/src/tarefas/S004/cliente"
	fatFranqueado "robot/src/tarefas/S004/franqueado"
	fatNucleo "robot/src/tarefas/S004/nucleo"
	"time"
)

var Processar bool = true

func Start(tempo int) {
	fmt.Println("Subindo modulo S004 -> Autogerar fatura")

	for {
		hora := time.Now().Hour()
		dia := time.Now().Day()

		if (dia == 1 || dia == 16) && hora == 1 {
			if Processar && !competenciaJaProcessada(dia) {
				fmt.Printf("S004: iniciando fechamento dia=%d\n", dia)

				if dia == 1 {
					if err := fecharCliente(); err != nil {
						fmt.Println("S004 fecharCliente:", err)
					}
				}

				if err := fecharFranqueado(); err != nil {
					fmt.Println("S004 fecharFranqueado:", err)
				}

				if err := fecharNucleo(); err != nil {
					fmt.Println("S004 fecharNucleo:", err)
				}

				marcarCompetenciaProcessada(dia)
			}

			// Evita reprocessar a cada ~10 min na mesma hora
			Processar = false
		} else {
			Processar = true
		}

		time.Sleep(10 * time.Minute)
		fmt.Println("Processando -> Autogerar fatura")
	}
}

func competenciaKey(dia int) string {
	now := time.Now()
	quinzena := 1
	if dia == 16 {
		quinzena = 2
	}
	return fmt.Sprintf("%d-%02d-Q%d", now.Year(), int(now.Month()), quinzena)
}

func lockPath(dia int) string {
	return filepath.Join(os.TempDir(), "confmonit-s004-"+competenciaKey(dia)+".done")
}

func competenciaJaProcessada(dia int) bool {
	_, err := os.Stat(lockPath(dia))
	return err == nil
}

func marcarCompetenciaProcessada(dia int) {
	path := lockPath(dia)
	_ = os.WriteFile(path, []byte(time.Now().Format(time.RFC3339)), 0644)
	fmt.Println("S004: competência marcada como processada ->", path)
}

func fecharCliente() error {
	db, err := auxiliar.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	data := time.Now().Add(-720 * time.Hour).Format("2006-01-02 15:04:05")
	tab, err := db.Query(`
		SELECT cliente.ID_Cliente
		FROM cliente
		WHERE cliente.DataCancelamento IS NULL
		OR cliente.DataCancelamento > ?
	`, data)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var id string
		if err := tab.Scan(&id); err != nil {
			return err
		}
		if err := fatCliente.Fechar(id); err != nil {
			fmt.Println("S004 Fechar cliente", id, err)
		}
	}
	return nil
}

func fecharFranqueado() error {
	db, err := auxiliar.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	data := time.Now().Add(-720 * time.Hour).Format("2006-01-02 15:04:05")
	tab, err := db.Query(`
		SELECT franqueado.ID_Franqueado
		FROM franqueado
		WHERE franqueado.Ativo = 'S'
		AND (
			franqueado.DataCancelamento IS NULL
			OR franqueado.DataCancelamento > ?
		)
	`, data)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var id string
		if err := tab.Scan(&id); err != nil {
			return err
		}
		if err := fatFranqueado.Fechar(id); err != nil {
			fmt.Println("S004 Fechar franqueado", id, err)
		}
	}
	return nil
}

func fecharNucleo() error {
	fmt.Println("Fechar Nucleo Inicializado")
	db, err := auxiliar.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	data := time.Now().Add(-720 * time.Hour).Format("2006-01-02 15:04:05")
	tab, err := db.Query(`
		SELECT representante.ID_Representante
		FROM representante
		WHERE representante.Ativo = 'S'
		AND representante.visivel = 'S'
		AND (
			representante.DataCancelamento IS NULL
			OR representante.DataCancelamento > ?
		)
	`, data)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var id string
		if err := tab.Scan(&id); err != nil {
			return err
		}
		if err := fatNucleo.Fechar(id); err != nil {
			fmt.Println("S004 Fechar nucleo", id, err)
		}
	}
	fmt.Println("Fechar Nucleo Finalizado")
	return nil
}
