query VerificaEmail verb=POST {
  api_group = "ConfMonitAuth"

  input {
    text Authorization? filters=trim
    text email1? filters=trim
  }

  stack {
    function.run ConfMonitAuth_VerificaEmail {
      input = {
        Authorization: $input.Authorization
        email1       : $input.email1
      }
    } as $func_1
  }

  response = $func_1
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