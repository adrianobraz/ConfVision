package auxiliar

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func GeradorDeId() string {
	agora := time.Now()
	numero1 := rand.Intn(999999-100000) + 100000
	numero2 := rand.Intn(99999-10000) + 10000
	parte1 := agora.Format("20060102030405")
	parte2 := strconv.Itoa(numero1) + strconv.Itoa(numero2)
	ret := parte1 + parte2
	return fmt.Sprintf("%025s", ret)
}
