// Storage Contabo do franqueado (credenciais mascaradas no response)
query vis_gravacao_storage_by_franqueado verb=GET {
  api_group = "confVision"

  input {
    text id_franqueado? filters=trim
    text status? filters=trim
  }

  stack {
    db.query vis_gravacao_storage {
      where = $db.vis_gravacao_storage.id_franqueado == $input.id_franqueado && $db.vis_gravacao_storage.status ==? $input.status
      sort = {vis_gravacao_storage.created_at: "desc"}
      return = {type: "list"}
    } as $lista
  
    var $dados {
      value = []
    }
  
    foreach ($lista) {
      each as $item {
        array.push $dados {
          value = {
            id              : $item.id
            created_at      : $item.created_at
            id_franqueado   : $item.id_franqueado
            s3_endpoint     : $item.s3_endpoint
            s3_bucket       : $item.s3_bucket
            s3_tenant_id    : $item.s3_tenant_id
            s3_access_key   : "****"
            segmento_minutos: $item.segmento_minutos
            status          : $item.status
            provisionado_em : $item.provisionado_em
            cancelado_em    : $item.cancelado_em
            observacao      : $item.observacao
          }
        }
      }
    }
  }

  response = {dados: $dados}
}