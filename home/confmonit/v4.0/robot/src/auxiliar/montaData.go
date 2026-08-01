package auxiliar

import "time"

func MontaData(qtdMes int, tInicio, tFinal *time.Time) error {
	agora := time.Now().AddDate(0, -qtdMes, 0)
	ano := agora.Year()
	mes := agora.Month()
	var dia int

	switch int(agora.Month()) {
	case 1, 3, 5, 7, 8, 10, 12:
		dia = 31
	case 2:
		dia = 28
	case 4, 6, 9, 11:
		dia = 30

	}
	*tInicio = time.Date(ano, mes, 1, 0, 0, 0, 0, time.UTC)
	*tFinal = time.Date(ano, mes, dia, 23, 59, 59, 0, time.UTC)

	return nil
}
