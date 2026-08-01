// Atualizar storage Contabo do franqueado
query "vis_gravacao_storage/{vis_gravacao_storage_id}" verb=PUT {
  api_group = "confVision"

  input {
    int vis_gravacao_storage_id? filters=min:1
    dblink {
      table = "vis_gravacao_storage"
    }
  }

  stack {
    db.get vis_gravacao_storage {
      field_name = "id"
      field_value = $input.vis_gravacao_storage_id
    } as $atual
  
    precondition ($atual != null) {
      error_type = "notfound"
      error = "Storage nao encontrado"
    }
  
    var $s3_access_key {
      value = $input.s3_access_key
    }
  
    conditional {
      if ($s3_access_key|is_empty) {
        var.update $s3_access_key {
          value = $atual.s3_access_key
        }
      }
    }
  
    var $s3_secret_key {
      value = $input.s3_secret_key
    }
  
    conditional {
      if ($s3_secret_key|is_empty) {
        var.update $s3_secret_key {
          value = $atual.s3_secret_key
        }
      }
    }
  
    db.edit vis_gravacao_storage {
      field_name = "id"
      field_value = $input.vis_gravacao_storage_id
      enforce_hidden_fields = false
      data = {
        s3_endpoint     : $input.s3_endpoint
        s3_bucket       : $input.s3_bucket
        s3_tenant_id    : $input.s3_tenant_id
        s3_access_key   : $s3_access_key
        s3_secret_key   : $s3_secret_key
        segmento_minutos: $input.segmento_minutos
        status          : $input.status
        cancelado_em    : $input.cancelado_em
        observacao      : $input.observacao
      }
    } as $model
  }

  response = {
    id              : $model.id
    id_franqueado   : $model.id_franqueado
    s3_endpoint     : $model.s3_endpoint
    s3_bucket       : $model.s3_bucket
    s3_tenant_id    : $model.s3_tenant_id
    s3_access_key   : "****"
    segmento_minutos: $model.segmento_minutos
    status          : $model.status
    provisionado_em : $model.provisionado_em
    cancelado_em    : $model.cancelado_em
    observacao      : $model.observacao
  }
}