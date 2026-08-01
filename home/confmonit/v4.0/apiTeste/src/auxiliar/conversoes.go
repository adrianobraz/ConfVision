package auxiliar

import (
	"fmt"
	"strconv"
	"strings"
)

func StrigToFloat(in string, out *float64) error {
	// Converte valor exedente pora float

	valor, erro := strconv.ParseFloat(strings.ReplaceAll(in, ",", "."), 32)
	if erro != nil {

		return erro
	}
	*out = valor
	return nil
}

func FloatToString(in float64, out *string) {
	*out = strings.ReplaceAll(fmt.Sprintf("%.2f", in), ".", ",")
}

func TimeUsToBr(dataIn string) (DataOut string) {
	// 2006-01-02 15:04:05"

	dia := dataIn[8:10]
	mes := dataIn[5:7]
	ano := dataIn[0:4]
	hora := dataIn[12:]
	DataOut = fmt.Sprintf("%s/%s/%s %s", dia, mes, ano, hora)
	return
}

func TimeBrToUs(dataIn string) (DataOut string) {
	// 02/01/2006 15:04:05"
	dia := dataIn[0:2]
	mes := dataIn[3:5]
	ano := dataIn[6:10]
	hora := dataIn[12:]
	DataOut = fmt.Sprintf("%s-%s-%s %s", ano, mes, dia, hora)
	return
}
