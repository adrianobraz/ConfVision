query nomecadwhats verb=GET {
  api_group = "testes"

  input {
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/cliente/listarByIdFranqueado"
      method = "POST"
      params = {}
        |set:"idFranqueado":"2025060911203125804370638"
      headers = []
        |push:"Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRob3JpemVkIjp0cnVlLCJleHAiOjE3ODEzMzcwMTksImlkIjoiIn0.tpJIKm3FRIRd5Fbr1ECAllZ4lr6Tk9BGQrJx1DD_UWc"
        |push:"Content-Type: application/json"
    } as $api1
  }

  response = {dados: $api1}
}