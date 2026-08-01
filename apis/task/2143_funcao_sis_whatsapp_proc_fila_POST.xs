query FuncaoSis_WhatsappProcFila verb=POST {
  api_group = "task"

  input {
    text SendMsgWhats? filters=trim
    text numerowhatsapp? filters=trim
    bool enviaSom?
    text SendAudioWhats? filters=trim
    int tblWhatsAppEnviadosTexto?
    int tblWhatsAppEnviadosAudio?
    int tblWhatsAppEnviadosLigar?
    bool ligar?
    text idFranqueado? filters=trim
    text nomecliente? filters=trim
    text empresanome? filters=trim
    text tipoevento? filters=trim
    text zona? filters=trim
    text local? filters=trim
    int idevento?
    text datahorario? filters=trim
    text ideventgo? filters=trim
    bool enviartexto?
    text idProcesso? filters=trim
    text idDispositivo? filters=trim
    bool notificarsempre?
    text DispNome? filters=trim
    text DispDescricao? filters=trim
    text DispTipo? filters=trim
  }

  stack {
    function.run FuncaoSis_WhatsappProcFila {
      input = {
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
      }
    } as $func4
  }

  response = {dados: $func4.dados}
}