query RelatorioEventosCli verb=POST {
  api_group = "ConfMonitCliente"

  input {
    text dispId? filters=trim
    text dataInicio? filters=trim
    text dataFim? filters=trim
    text grupos? filters=trim
    text Authorization? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/evento/listarByDispStartEndGrupo"
      method = "POST"
      params = {}
        |set:"dispId":$input.dispId
        |set:"dataInicio":$input.dataInicio
        |set:"dataFim":$input.dataFim
        |set:"grupos":$input.grupos
      headers = []
        |push:"Content-Type: application/json"
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $api1
  }

  response = $api1
}