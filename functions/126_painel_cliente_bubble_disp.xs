function PainelClienteBubbleDisp {
  input {
    text idDispositivo? filters=trim
    text token? filters=trim
    text armado? filters=trim
    text idCliente? filters=trim
  }

  stack {
    var $func1 {
      value = {}
    }
  
    conditional {
      if ($input.idDispositivo != null || ($input.idDispositivo|is_empty) == false) {
        for (3) {
          each as $index {
            function.run ConfMonitEquipamento_DadosEquip {
              input = {
                Authorization: $input.token
                idDispositivo: $input.idDispositivo
              }
            } as $func1
          
            conditional {
              if ($input.armado == $func1.armado) {
                util.sleep {
                  value = 2
                }
              }
            
              else {
                break
              }
            }
          }
        }
      }
    }
  
    function.run ListaEquipamentos {
      input = {Authorization: $input.token, idCliente: $input.idCliente}
    } as $func2|set:"":`$func2.response.result.dados`
  
    precondition ($func1 != null)
    function.run ConfMonitEquipamento_DadosEquip {
      input = {
        Authorization: $input.token
        idDispositivo: `$func2.0.idDispositivo`
      }
    } as $func1
  }

  response = "Dados"
    |set:"Equipamento":$func1
    |set:"Dispositivos":$func2
}