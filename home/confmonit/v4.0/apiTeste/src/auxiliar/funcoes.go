package auxiliar

import (
	"errors"
	"strings"
)

func InverteEstado(e *string) error {
	if *e == "S" {
		*e = "N"
		return nil
	} else if *e == "N" {
		*e = "S"
		return nil
	} else {
		return errors.New("erro ao inverter")
	}
}

func ClearTel(d string) string {

	t1 := strings.ReplaceAll(d, "(", "")

	t2 := strings.ReplaceAll(t1, ")", "")

	t3 := strings.ReplaceAll(t2, " ", "")

	return strings.ReplaceAll(t3, "-", "")

}
