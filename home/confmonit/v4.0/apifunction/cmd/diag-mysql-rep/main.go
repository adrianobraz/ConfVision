package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	user := strings.TrimSpace(os.Getenv("BD_USER_MV4"))
	pass := strings.TrimSpace(os.Getenv("BD_PASS_MV4"))
	host := strings.TrimSpace(os.Getenv("BD_HOST_MV4"))
	port := strings.TrimSpace(os.Getenv("BD_PORT_MV4"))
	base := strings.TrimSpace(os.Getenv("BD_BASE_MV4"))
	if port == "" {
		port = "3306"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, base)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	usuario := strings.TrimSpace(os.Getenv("USUARIO"))
	if usuario == "" {
		usuario = "usuario@teste.rep"
	}

	fmt.Printf("=== usuarios (usuario=%s) ===\n", usuario)
	printRows(db, `
		SELECT ID_Usuario, Usuario, Nome, Master, Ativo, ID_Vinculo
		FROM usuarios
		WHERE Usuario = ?
		LIMIT 5
	`, usuario)

	fmt.Printf("\n=== representante vinculado ===\n")
	printRows(db, `
		SELECT r.ID_Representante, r.RazaoSocial, r.NomeFantasia, r.ID_Central, r.Ativo
		FROM usuarios u
		JOIN representante r ON r.ID_Representante = u.ID_Vinculo
		WHERE u.Usuario = ?
		LIMIT 1
	`, usuario)
}

func printRows(db *sql.DB, query string, args ...any) {
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("query erro: %v", err)
		return
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var out []map[string]any
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			log.Printf("scan: %v", err)
			continue
		}
		row := map[string]any{}
		for i, c := range cols {
			switch v := vals[i].(type) {
			case []byte:
				row[c] = string(v)
			default:
				row[c] = v
			}
		}
		out = append(out, row)
	}
	if len(out) == 0 {
		fmt.Println("(nenhum registro)")
		return
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(b))
}
