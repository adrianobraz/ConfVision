query PainelClienteBubbleDisp verb=POST {
  api_group = "admMonitoramento"

  input {
    text token? filters=trim
    text armado? filters=trim
    text idCliente? filters=trim
    text Dispositivo_ID? filters=trim
    uuid? idCLI?
  }

  stack {
    var $UserCliConf_Equipamentos1 {
      value = ""
    }
  
    conditional {
      if (($input.idCLI|is_uuid) && ($input.idCLI|is_empty) == false) {
        var $func_1 {
          value = {}
        }
      
        db.query UserCliConf_Equipamentos {
          where = $db.UserCliConf_Equipamentos.usercliconf_id == $input.idCLI
          return = {type: "list"}
        } as $UserCliConf_Equipamentos1
      
        conditional {
          if ($UserCliConf_Equipamentos1 != null) {
            var $func1 {
              value = ""
            }
          
            var $func2 {
              value = {}
            }
          
            conditional {
              if (($UserCliConf_Equipamentos1|count) > 1) {
                foreach ($UserCliConf_Equipamentos1) {
                  each as $item {
                    function.run ConfMonitEquipamento_DadosEquip {
                      input = {
                        Authorization: $input.token
                        idDispositivo: `$item.idDisp`
                      }
                    } as $func0
                  
                    conditional {
                      if ($func1 == null) {
                        var.update $func1 {
                          value = $func0
                        }
                      
                        var.update $func2 {
                          value = ""|append:$func0:""
                        }
                      }
                    
                      else {
                        var.update $func2 {
                          value = ""|set:"":$func2|append:$func0:""
                        }
                      }
                    }
                  }
                }
              }
            
              else {
                function.run ConfMonitEquipamento_DadosEquip {
                  input = {
                    Authorization: $input.token
                    idDispositivo: `$UserCliConf_Equipamentos1.0.idDisp`
                  }
                } as $func0
              
                var.update $func1 {
                  value = $func0
                }
              
                var.update $func2 {
                  value = $func0|safe_array
                }
              }
            }
          
            var.update $func_1 {
              value = "Dados"
                |set:"Equipamento":$func1
                |set:"Dispositivos":$func2
            }
          }
        }
      }
    
      else {
        function.run PainelClienteBubbleDisp {
          input = {
            idDispositivo: $input.Dispositivo_ID
            token        : $input.token
            armado       : $input.armado
            idCliente    : $input.idCliente
          }
        } as $func_1
      }
    }
  }

  response = $func_1
}