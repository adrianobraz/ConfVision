// Update WhatsappProcFila record
query "whatsappprocfila/{whatsappprocfila_id}" verb=PUT {
  api_group = "task"

  input {
    int whatsappprocfila_id? filters=min:1
    dblink {
      table = "WhatsappProcFila"
    }
  }

  stack {
    db.edit WhatsappProcFila {
      field_name = "id"
      field_value = $input.whatsappprocfila_id
      enforce_hidden_fields = false
      data = {
        SendMsgWhats            : $input.SendMsgWhats
        numerowhatsapp          : $input.numerowhatsapp
        enviaSom                : $input.enviaSom
        SendAudioWhats          : $input.SendAudioWhats
        tblWhatsAppEnviadosTexto: $input.tblWhatsAppEnviadosTexto
        tblWhatsAppEnviadosAudio: $input.tblWhatsAppEnviadosAudio
        tblWhatsAppEnviadosLigar: $input.tblWhatsAppEnviadosLigar
        ligar                   : $input.ligar
        idFranqueado            : $input.idFranqueado
        nomecliente             : $input.nomecliente
        empresanome             : $input.empresanome
        GRUPOFALHA              : $input.GRUPOFALHA
        tipoevento              : $input.tipoevento
        zona                    : $input.zona
        local                   : $input.local
        idevento                : $input.idevento
        datahorario             : $input.datahorario
        ideventgo               : $input.ideventgo
        enviartexto             : $input.enviartexto
        idProcesso              : $input.idProcesso
        idDispositivo           : $input.idDispositivo
        notificarsempre         : $input.notificarsempre
        DispNome                : $input.DispNome
        DispDescricao           : $input.DispDescricao
        DispTipo                : $input.DispTipo
        analise                 : $input.analise
        zonauser                : $input.zonauser
        particao                : $input.particao
        disparo                 : $input.disparo
        dtDisparo               : now
      }
    } as $model
  }

  response = {dados: $model}
}