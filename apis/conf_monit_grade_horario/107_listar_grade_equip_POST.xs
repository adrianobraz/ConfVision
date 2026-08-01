query ListarGradeEquip verb=POST {
  api_group = "ConfMonitGradeHorario"

  input {
    text Authorization? filters=trim
    text idDispositivo? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/grade/listarByIdDispositivo"
      method = "POST"
      params = {}
        |set:"idDispositivo":$input.idDispositivo
      headers = []
        |push:"Content-Type: application/json"
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $api1
  }

  response = $api1
}