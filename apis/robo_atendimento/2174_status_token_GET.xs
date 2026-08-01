query StatusToken verb=GET {
  api_group = "RoboAtendimento"

  input {
    text token? filters=trim
  }

  stack {
    var $idDispositivo {
      value = ""
    }
  
    conditional {
      if (($input.token|is_empty) == false) {
        !db.query status_link {
          where = $db.status_link.token == $input.token && $db.status_link.expiresAt > now
          return = {type: "single"}
        } as $status_link1
      
        db.query status_link {
          where = $db.status_link.token == $input.token
          return = {type: "single"}
        } as $status_link1
      
        precondition (($status_link1|is_empty) == false) {
          error_type = "accessdenied"
          error = "Token Expirado"
        }
      
        conditional {
          if ($status_link1.grupo == "ALARME") {
            conditional {
              if (($status_link1.created_at|add_secs_to_timestamp:3600) > now) {
                var.update $idDispositivo {
                  value = $status_link1.idDispositivo
                }
              }
            }
          }
        
          else {
            conditional {
              if ($status_link1.created_at > now) {
                var.update $idDispositivo {
                  value = $status_link1.idDispositivo
                }
              }
            }
          }
        }
      }
    }
  }

  response = {dados: $idDispositivo}
}