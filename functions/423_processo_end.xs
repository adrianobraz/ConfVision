function ProcessoEnd {
  input {
    text idUsuario? filters=trim
    text Descricao? filters=trim
    text UsuarioNome? filters=trim
    text idProcesso? filters=trim
    text telefone? filters=trim
    text token? filters=trim
    text ip? filters=trim
    text ipcidade? filters=trim
    text ipestado? filters=trim
    text ipcep? filters=trim
    int geolat?
    int geolon?
    text georua? filters=trim
    text geonumero? filters=trim
    text geobairro? filters=trim
    text geocidade? filters=trim
    text geoestado? filters=trim
    text device? filters=trim
    text browser? filters=trim
    text sitema? filters=trim
    text timezone? filters=trim
    text fingerprint? filters=trim
  }

  stack {
    function.run WebLogarCache as $func1
    api.request {
      url = "http://185.130.61.4:2010/v4/terminal/getDadosProcessoById"
      method = "POST"
      params = {}
        |set:"idProcesso":$input.idProcesso
      headers = []
        |push:"Authorization: Bearer " ~ $func1
        |push:"Content-Type: application/json"
    } as $api1
  
    conditional {
      if ($api1.response.result.dados.dataAtenFim == "01/01/0001 00:00:00") {
        api.request {
          url = "http://185.130.61.4:2010/v4/terminal/finalizarProcesso"
          method = "POST"
          params = {}
            |set:"idProcesso":$input.idProcesso
            |set:"idCliente":$input.idUsuario
            |set:"descricao":$input.Descricao
            |set:"nome":$input.UsuarioNome
          headers = []
            |push:"Authorization: Bearer " ~ $func1
            |push:"Content-Type: application/json"
        } as $api1
      
        var $x1 {
          value = $api1.response.result.dados
        }
      
        db.add alarmEvent_Finalizados {
          enforce_hidden_fields = false
          data = {
            idprocesso : $input.idProcesso
            telefone   : $input.telefone
            token      : $input.token
            mensagem   : $input.Descricao
            ip         : $input.ip
            ipcidade   : $input.ipcidade
            ipestado   : $input.ipestado
            ipcep      : $input.ipcep
            geolat     : $input.geolat
            geolon     : $input.geolon
            georua     : $input.georua
            geonumero  : $input.geonumero
            geobairro  : $input.geobairro
            geocidade  : $input.geocidade
            geoestado  : $input.geoestado
            device     : $input.device
            browser    : $input.browser
            sitema     : $input.sitema
            timezone   : $input.timezone
            fingerprint: $input.fingerprint
            idUsuario  : $input.idUsuario
            Usuario    : $input.UsuarioNome
          }
        } as $alarmEvent_Finalizados1
      }
    
      else {
        var $x1 {
          value = []
        }
      }
    }
  }

  response = {dados: $x1}
}