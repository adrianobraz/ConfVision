query "wh_VerProcesso/{idProcesso}" verb=GET {
  api_group = "RoboAtendimento"

  input {
    text idProcesso? filters=trim
  }

  stack {
    db.query alarm_events {
      where = $db.alarm_events.idProcesso == $input.idProcesso
      return = {type: "single"}
    } as $alarm_events1
  
    function.run WebLogarCache as $func1
    function.run ConfMonitCliente_DadosClienteID {
      input = {
        Authorization: $func1
        idcliente    : $alarm_events1.idCliente
      }
    } as $func2
  
    function.run ConfMonitEquipamento_DadosEquip {
      input = {
        Authorization: $func1
        idDispositivo: $alarm_events1.idDispositivo
      }
    } as $func3
  
    api.request {
      url = "http://185.130.61.4:2010/v4/terminal/getDadosProcessoById"
      method = "POST"
      params = {}
        |set:"idProcesso":$input.idProcesso
      headers = []
        |push:"Authorization: Bearer " ~ $func1
        |push:"Content-Type: application/json"
    } as $api1
  }

  response = ""
    |set:"Eventos":$alarm_events1
    |set:"Cliente":$func2.response.result.dados
    |set:"Dispositivo":$func3
    |set:"Processo":$api1.response.result.dados
}