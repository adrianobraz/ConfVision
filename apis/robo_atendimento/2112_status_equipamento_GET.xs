query StatusEquipamento verb=GET {
  api_group = "RoboAtendimento"

  input {
    text token? filters=trim
    text Dispositivo? filters=trim
  }

  stack {
    conditional {
      if (($input.token|is_empty) == false) {
        db.query status_link {
          where = $db.status_link.token == $input.token && $db.status_link.expiresAt > now
          return = {type: "single"}
        } as $status_link1
      
        precondition (($status_link1|is_empty) == false) {
          error_type = "accessdenied"
          error = "Token Expirado"
        }
      
        var $idDispositivo {
          value = $status_link1.idDispositivo
        }
      }
    
      elseif (($input.Dispositivo|is_empty) == false) {
        var $idDispositivo {
          value = $input.Dispositivo
        }
      }
    }
  
    precondition (($idDispositivo|is_empty) == false) {
      error_type = "accessdenied"
      error = "Token Expirado"
    }
  
    function.run WebLogarCache as $func1
    function.run ConfMonitEquipamento_DadosEquip {
      input = {Authorization: $func1, idDispositivo: $idDispositivo}
    } as $func_2
  }

  response = {dados: $func_2}
}