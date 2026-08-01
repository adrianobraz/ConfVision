function ListaEquipamentos {
  input {
    text Authorization? filters=trim
    text idCliente? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/dispositivo/listarByIdCliente"
      method = "POST"
      params = {}|set:"idCliente":$input.idCliente
      headers = []
        |push:"Content-Type: application/json"
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $api1
  }

  response = $api1
}