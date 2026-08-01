function ConfMonitAuth_VerificaEmail {
  input {
    text Authorization? filters=trim
    text email1? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/cliente/getEmail1LivreByEmail1"
      method = "POST"
      params = {}|set:"email1":$input.email1
      headers = []
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
        |push:"Content-Type: application/json"
    } as $api1
  }

  response = $api1
  cache = {
    ttl       : 86400
    input     : true
    auth      : true
    datasource: true
    ip        : false
    headers   : []
    env       : []
  }
}