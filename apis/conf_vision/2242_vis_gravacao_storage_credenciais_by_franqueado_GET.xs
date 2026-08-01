// Credenciais Contabo completas (uso interno worker / backend)
query vis_gravacao_storage_credenciais_by_franqueado verb=GET {
  api_group = "confVision"

  input {
    text id_franqueado? filters=trim
  }

  stack {
    db.query vis_gravacao_storage {
      where = $db.vis_gravacao_storage.id_franqueado == $input.id_franqueado && $db.vis_gravacao_storage.status == "ativo"
      return = {type: "single"}
    } as $storage
  
    precondition ($storage != null) {
      error_type = "notfound"
      error = "Storage ativo nao encontrado para o franqueado"
    }
  }

  response = $storage
}