// dataAtualFrmatada retorna
function dataAtualFormatada(data) {
  var data = new Date(data),
    dia = data.getDate().toString(),
    diaF = dia.length == 1 ? "0" + dia : dia,
    mes = (data.getMonth() + 1).toString(), //+1 pois no getMonth Janeiro começa com zero.
    mesF = mes.length == 1 ? "0" + mes : mes,
    anoF = data.getFullYear();
  return (
    diaF +
    "/" +
    mesF +
    "/" +
    anoF +
    " " +
    data.getHours() +
    ":" +
    data.getMinutes() +
    ":" +
    data.getSeconds()
  );
}

// retorna a diferença em dias
function diff(data1, data2) {
  d1 = new Date(data1)
  d2 = new Date(data2)

  const diff = Math.abs(data1.getTime() - data2.getTime())
  return Math.ceil(diff / (1000 * 60 * 60 * 24))
}

function horaBrToUs(horaBr) {
  // 12/10/2006 13:00:00
  const dia = horaBr.substring(0, 2)
  const mes = horaBr.substring(3, 5)
  const ano = horaBr.substring(6, 10)
  const hora = horaBr.substring(11)
  return `${ano}-${mes}-${dia} ${hora}`
}
