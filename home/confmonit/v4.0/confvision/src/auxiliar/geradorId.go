package auxiliar

import (
	"fmt"
	"time"
)

// GeradorId gera um id com comprimento 10
func GeradorDeId() string {
	return fmt.Sprintf("%010s", time.Now().Format("0601020304"))
}
