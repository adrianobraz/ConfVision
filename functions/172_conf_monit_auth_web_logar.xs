function ConfMonitAuth_WebLogar {
  input {
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/cliente/webLogar"
      method = "POST"
      params = {}
        |set:"chave":"senha"
        |set:"senha":"WHdQkY&RX%W%4RArwm1Q"
      headers = []
        |push:"Content-Type: application/json"
    } as $api1|set:"":`$api1.response.result.dados`
  }

  response = $api1
  cache = {
    ttl       : 43200
    input     : true
    auth      : true
    datasource: true
    ip        : false
    headers   : []
    env       : []
  }
}