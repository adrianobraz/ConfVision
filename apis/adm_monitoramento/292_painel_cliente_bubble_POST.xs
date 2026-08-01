query PainelClienteBubble verb=POST {
  api_group = "admMonitoramento"

  input {
    text email? filters=trim
    text senha? filters=trim
    text token? filters=trim
  }

  stack {
    var $email {
      value = $input.email
    }
  
    var $senha {
      value = $input.senha
    }
  
    conditional {
      if (($input.token|is_empty) == false) {
        api.request {
          url = "https://xpcy-oyme-lno7.b2.xano.io/api:TYiUYsHb/auth/me"
          method = "GET"
          headers = []
            |push:"Accept: application/json"
            |push:"Authorization: " ~ $input.token
        } as $api2
      
        conditional {
          if ($api2.response.status != 401) {
            var.update $email {
              value = $api2.response.result.dados.usuario
            }
          
            var.update $senha {
              value = $api2.response.result.dados.password
            }
          }
        }
      }
    }
  
    conditional {
      if ($email|is_empty) {
        var.update $email {
          value = "!@#$%¨&*()_+@!@#$%¨&*()_+.!@#$%¨&*()_+"
        }
      }
    }
  
    db.get UserCliConf {
      field_name = "email"
      field_value = $email
      output = [
        "id"
        "created_at"
        "email"
        "senha"
        "Data"
        "idFranqueado"
        "idCliente"
        "nome"
        "telefone1"
        "telefone2"
        "Descricao"
        "Acesso"
      ]
    } as $UserCliConf2
  
    var $x1 {
      value = {}
    }
  
    conditional {
      if ($UserCliConf2 != null) {
        security.check_password {
          text_password = $senha
          hash_password = $UserCliConf2.senha
        } as $x2
      
        precondition ($x2) {
          error = " "
        }
      
        function.run WebLogarCache as $func2
        function.run ConfMonitCliente_DadosClienteID {
          input = {
            Authorization: $func2
            idcliente    : $UserCliConf2.idCliente
          }
        } as $func1|set:"":`$func1.response.result.dados`
      
        var.update $email {
          value = $func1.email1
        }
      
        var.update $func1.envioEmail {
          value = $UserCliConf2.id
        }
      
        var.update $x1 {
          value = $func1
            |set:"token":$func2
            |set:"usercliconf":`$var.UserCliConf2|safe_array`
        }
      
        function.run usercliconflog {
          input = {
            usercliconf_id: $UserCliConf2.id
            idFranqueado  : $func1.idFranqueado
            idCliente     : $func1.idCliente
            Descricao     : "Login"
            Ocorrencia    : ""
            Evento        : "Login"
            log           : ""
            idDisp        : $func1.idDispApp
            DispApp       : ""
          }
        } as $func_1
      }
    
      else {
        function.run ConfMonitAuth_Autenticar {
          input = {email: $email, senha: $senha}
        } as $api1
      
        precondition ($api1 != null) {
          error = " "
        }
      
        var.update $api1.envioEmail {
          value = ""
        }
      
        var.update $x1 {
          value = $api1|set:"usercliconf":{}
        }
      }
    }
  
    conditional {
      if ($input.token|is_empty) {
        db.query UserCliConfToken {
          where = $db.UserCliConfToken.usuario == $email && $db.UserCliConfToken.nivel == 1
          return = {type: "single"}
        } as $UserCliConfToken1
      
        conditional {
          if ($UserCliConfToken1|is_empty) {
            security.create_uuid as $uuid
            db.add UserCliConfToken {
              enforce_hidden_fields = false
              data = {
                created_at: "now"
                email     : $var.uuid ~ "@admmonitoramento.com.br"
                senha     : "1M2O3N4@5I6T7o8#9r0a1m2E3*4N5t6o7"
                usuario   : $email
                password  : $senha
                nivel     : 1
              }
            } as $UserCliConfToken2
          
            security.create_auth_token {
              table = "UserCliConfToken"
              extras = {}
              expiration = 3600
              id = $UserCliConfToken2.id
            } as $authToken
          }
        
          else {
            security.create_auth_token {
              table = "UserCliConfToken"
              extras = {}
              expiration = 3600
              id = $UserCliConfToken1.id
            } as $authToken
          }
        }
      }
    
      else {
        var $authToken {
          value = ""
        }
      }
    }
  }

  response = $x1|set:"tokenStorage":$authToken
}