query UserCliConf_Equip verb=POST {
  api_group = "UserCliConf"

  input {
    text idCliente? filters=trim
    text idDisp? filters=trim
    uuid? usercliconf_id?
    text DispApp? filters=trim
  }

  stack {
    var $x1 {
      value = {}
    }
  
    conditional {
      if ($input.usercliconf_id == null) {
        db.query UserCliConf_Equipamentos {
          where = $db.UserCliConf_Equipamentos.idCliente == $input.idCliente
          return = {type: "list"}
        } as $UserCliConf_Equipamentos1
      
        var.update $x1 {
          value = $UserCliConf_Equipamentos1
        }
      }
    
      else {
        db.add UserCliConf_Equipamentos {
          enforce_hidden_fields = false
          data = {
            created_at    : "now"
            idDisp        : $input.idDisp
            usercliconf_id: $input.usercliconf_id
            DispApp       : $input.DispApp
            idCliente     : $input.idCliente
          }
        } as $UserCliConf_Equipamentos2
      
        var.update $x1 {
          value = `$var.UserCliConf_Equipamentos2|safe_array`
        }
      }
    }
  }

  response = $x1
}