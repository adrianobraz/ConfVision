query WhatsappLigar verb=POST {
  api_group = "task"

  input {
    text franqueado? filters=trim
    text telefone? filters=trim
    text nomecliente? filters=trim
    text empresanome? filters=trim
    text tipoevento? filters=trim
    text zona? filters=trim
    text local? filters=trim
    text datahorario? filters=trim
    int alarm_events_id?
    text ideventgo? filters=trim
    text idDispositivo? filters=trim
    text idProcesso? filters=trim
    bool notificarsempre?
    text DispNome? filters=trim
    text DispDescricao? filters=trim
    text DispTipo? filters=trim
  }

  stack {
    function.run WhatsappLigar {
      input = {
        idfranqueado   : $input.franqueado
        telefone       : $input.telefone
        nomecliente    : $input.nomecliente
        empresanome    : $input.empresanome
        tipoevento     : $input.tipoevento
        zona           : $input.zona
        local          : $input.local
        idevento       : $input.alarm_events_id
        datahorario    : $input.datahorario
        ideventogo     : $input.ideventgo
        idDispositivo  : $input.idDispositivo
        idProcesso     : $input.idProcesso
        notificarsempre: $input.notificarsempre
        DispNome       : $input.DispNome
        DispDescricao  : $input.DispDescricao
        DispTipo       : $input.DispTipo
      }
    } as $func1
  }

  response = {dados: $func1}
}