query "whatsappeventcadjanela/Queryall/WhatsEventCadJan" verb=GET {
  api_group = "RoboAtendimento"

  input {
    // Dia da semana a ser filtrado (0-7)
    enum[1:1] DiaSemana? {
      values = ["0", "1", "2", "3", "4", "5", "6", "7"]
    }
  }

  stack {
    // Define o início da janela (agora - 19 minutos)
    var $HoraInicioJanela {
      value = `now|add_secs_to_timestamp:-1140`
    }
  
    // Define o fim da janela (agora + 20 minutos)
    var $HoraFimJanela {
      value = `now|add_secs_to_timestamp:1200`
    }
  
    // Formata os timestamps para o formato HH:mm (H:i) usado no banco de dados
    var $MinStr {
      value = $HoraInicioJanela
        |format_timestamp:"H:i":"America/Sao_Paulo"
    }
  
    var $MaxStr {
      value = $HoraFimJanela
        |format_timestamp:"H:i":"America/Sao_Paulo"
    }
  
    // Verifica se a janela cruza a meia-noite (ex: 23:50 até 00:10)
    var $IsCrossMidnight {
      value = $MinStr > $MaxStr
    }
  
    // Busca registros onde a hora de início está dentro da janela
    conditional {
      if ($IsCrossMidnight) {
        db.query WhatsappEventCadJanela {
          where = ($db.WhatsappEventCadJanela.hora_inicio >= $MinStr || $db.WhatsappEventCadJanela.hora_inicio <= $MaxStr) && $db.WhatsappEventCadJanela.dias contains $input.DiaSemana
          sort = {WhatsappEventCadJanela.id: "asc"}
          return = {type: "list"}
        } as $WhatsappEventCadJanela1
      }
    
      else {
        db.query WhatsappEventCadJanela {
          where = ($db.WhatsappEventCadJanela.hora_inicio >= $MinStr && $db.WhatsappEventCadJanela.hora_inicio <= $MaxStr) && $db.WhatsappEventCadJanela.dias contains $input.DiaSemana
          sort = {WhatsappEventCadJanela.id: "asc"}
          return = {type: "list"}
        } as $WhatsappEventCadJanela1
      }
    }
  
    // Busca registros onde a hora de fim está dentro da janela
    conditional {
      if ($IsCrossMidnight) {
        db.query WhatsappEventCadJanela {
          where = ($db.WhatsappEventCadJanela.hora_fim >= $MinStr || $db.WhatsappEventCadJanela.hora_fim <= $MaxStr) && $db.WhatsappEventCadJanela.dias contains $input.DiaSemana
          sort = {WhatsappEventCadJanela.id: "asc"}
          return = {type: "list"}
        } as $WhatsappEventCadJanela2
      }
    
      else {
        db.query WhatsappEventCadJanela {
          where = ($db.WhatsappEventCadJanela.hora_fim >= $MinStr && $db.WhatsappEventCadJanela.hora_fim <= $MaxStr) && $db.WhatsappEventCadJanela.dias contains $input.DiaSemana
          sort = {WhatsappEventCadJanela.id: "asc"}
          return = {type: "list"}
        } as $WhatsappEventCadJanela2
      }
    }
  }

  response = {
    dados: ```
      {
        desarme: $WhatsappEventCadJanela1,
        arme: $WhatsappEventCadJanela2
      }
      ```
  }
}