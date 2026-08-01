// Valida acesso e devolve conversation_id para o Go buscar o MP3 na ElevenLabs
query fp_cd_ligacao_audio verb=POST {
  api_group = "Relatoriocentraldisparos"

  input {
    int id
    text idFranqueado filters=trim
  }

  stack {
    precondition (($input.idFranqueado|is_empty) == false) {
      error = "idFranqueado obrigatorio"
    }
  
    precondition ($input.id > 0) {
      error = "id obrigatorio"
    }
  
    db.get LigacaoHistorico {
      field_name = "id"
      field_value = $input.id
    } as $ligacao
  
    precondition (($ligacao|is_empty) == false) {
      error = "Ligacao nao encontrada"
    }
  
    precondition (($ligacao.conversation_id|is_empty) == false) {
      error = "Ligacao sem conversation_id (audio indisponivel)"
    }
  
    db.get whatsappLigarErro {
      field_name = "id"
      field_value = $ligacao.whatsappligarerro_id
    } as $erro
  
    precondition (($erro|is_empty) == false) {
      error = "Registro de ligacao nao encontrado"
    }
  
    precondition ($erro.franqueado == $input.idFranqueado) {
      error = "Acesso negado a esta ligacao"
    }
  }

  response = {
    conversation_id: $ligacao.conversation_id
    content_type   : "audio/mpeg"
    id             : $input.id
  }
}
