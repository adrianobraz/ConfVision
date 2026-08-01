function WebLogarCache {
  input {
  }

  stack {
    db.get TokenConfmonit {
      field_name = "id"
      field_value = 1
    } as $TokenConfmonit1
  
    var $x1 {
      value = ""
    }
  
    var $x2 {
      value = now
    }
  
    conditional {
      if ($TokenConfmonit1 != null && ($x2 - $TokenConfmonit1.created_at) < `40000`) {
        var.update $x1 {
          value = $TokenConfmonit1.token
        }
      }
    
      else {
        !api.request {
          url = "http://185.130.61.4:2000/v4/cliente/webLogar"
          method = "POST"
          params = {}
            |set:"chave":"senha"
            |set:"senha":"WHdQkY&RX%W%4RArwm1Q"
          headers = []
            |push:"Content-Type: application/json"
        } as $api1|set:"":`$api1.response.result.dados`
      
        function.run ConfMonitAuth_WebLogar as $func1
        var.update $x1 {
          value = $func1
        }
      
        conditional {
          if ($TokenConfmonit1 == null) {
            db.add TokenConfmonit {
              enforce_hidden_fields = false
              data = {created_at: "now", token: $x1}
            } as $TokenConfmonit2
          }
        
          else {
            db.edit TokenConfmonit {
              field_name = "id"
              field_value = 1
              enforce_hidden_fields = false
              data = {created_at: now, token: $x1}
            } as $TokenConfmonit3
          }
        }
      }
    }
  }

  response = $x1
  cache = {
    ttl       : 21600
    input     : true
    auth      : true
    datasource: true
    ip        : false
    headers   : []
    env       : []
  }
}