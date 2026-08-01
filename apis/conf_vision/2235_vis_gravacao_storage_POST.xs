// Cadastrar storage Contabo do franqueado (gravacao continua)
query vis_gravacao_storage verb=POST {
  api_group = "confVision"

  input {
    dblink {
      table = "vis_gravacao_storage"
    }
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    precondition (($input.s3_endpoint|is_empty) == false) {
      error = "s3_endpoint obrigatorio"
    }
  
    precondition (($input.s3_bucket|is_empty) == false) {
      error = "s3_bucket obrigatorio"
    }
  
    precondition (($input.s3_access_key|is_empty) == false) {
      error = "s3_access_key obrigatorio"
    }
  
    precondition (($input.s3_secret_key|is_empty) == false) {
      error = "s3_secret_key obrigatorio"
    }
  
    var $segmento_minutos {
      value = $input.segmento_minutos
    }
  
    conditional {
      if ($segmento_minutos == null || $segmento_minutos == 0) {
        var.update $segmento_minutos {
          value = 5
        }
      }
    }
  
    var $status {
      value = $input.status
    }
  
    conditional {
      if ($status|is_empty) {
        var.update $status {
          value = "ativo"
        }
      }
    }
  
    db.add vis_gravacao_storage {
      enforce_hidden_fields = false
      data = {
        created_at      : "now"
        id_franqueado   : $input.id_franqueado
        s3_endpoint     : $input.s3_endpoint
        s3_bucket       : $input.s3_bucket
        s3_tenant_id    : $input.s3_tenant_id
        s3_access_key   : $input.s3_access_key
        s3_secret_key   : $input.s3_secret_key
        segmento_minutos: $segmento_minutos
        status          : $status
        provisionado_em : "now"
        observacao      : $input.observacao
      }
    } as $model
  }

  response = {
    id              : $model.id
    created_at      : $model.created_at
    id_franqueado   : $model.id_franqueado
    s3_endpoint     : $model.s3_endpoint
    s3_bucket       : $model.s3_bucket
    s3_tenant_id    : $model.s3_tenant_id
    s3_access_key   : "****"
    segmento_minutos: $model.segmento_minutos
    status          : $model.status
    provisionado_em : $model.provisionado_em
    observacao      : $model.observacao
  }
}