function FuncaoSistema_sendWhatsEventCentralAlarm_0 {
  input {
    int idEvento?
  }

  stack {
    return {
      value = ""
    }
  
    db.get alarm_events {
      field_name = "id"
      field_value = $input.idEvento
    } as $alarm_events1
  
    function.run WebLogarCache as $func1
    function.run ConfMonitCliente_DadosClienteID {
      input = {
        Authorization: `$func1`
        idcliente    : $alarm_events1.idCliente
      }
    } as $func2
  
    function.run ConfMonitEquipamento_DadosEquip {
      input = {
        Authorization: `$func1`
        idDispositivo: $alarm_events1.idDispositivo
      }
    } as $func3
  
    function.run uazapi_EnvioMsgTexto {
      input = {
        number      : 5519992478859
        textMensagem: ```
          $func3.response.result.dados.nomeFranqueado ~ 
          "\nCentral de Monitoramento Informa:" ~
          "\nCliente: " ~ $alarm_events1.nomeCliente ~ 
          "\nEvento da sua Central - " ~ $alarm_events1.ctiGrupo ~ " - " ~ $alarm_events1.ctiDescricao ~ 
          "\nDispositivo: " ~ $func3.response.result.dados.nome ~ 
          "\nPartição: " ~ $func3.response.result.dados.particao ~ " - Conta: " ~ $alarm_events1.conta ~ 
          "\nCordialmente, " ~ $func2.response.result.dados.fraRazao
          ```|text_unescape
      }
    } as $func4
  }

  response = $func4
}