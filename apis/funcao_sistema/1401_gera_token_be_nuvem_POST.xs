query GeraTokenBeNuvem verb=POST {
  api_group = "FuncaoSistema"

  input {
    text email? filters=trim
    text password? filters=trim
  }

  stack {
    function.run FuncaoSistema_GeraTokenBeNuvem {
      input = {email: $input.email, password: $input.password}
    } as $func_1
  }

  response = $func_1
  cache = {
    ttl       : 3600
    input     : true
    auth      : true
    datasource: true
    ip        : false
    headers   : []
    env       : []
  }
}