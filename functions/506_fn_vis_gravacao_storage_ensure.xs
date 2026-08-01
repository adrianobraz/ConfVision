// Garante vis_gravacao_storage ativo para o franqueado (provisiona a partir de fp_config_financeiro se ausente)
function fn_vis_gravacao_storage_ensure {
  input {
    text id_franqueado? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }

    db.query vis_gravacao_storage {
      where = $db.vis_gravacao_storage.id_franqueado == $input.id_franqueado && $db.vis_gravacao_storage.status == "ativo"
      return = {type: "single"}
    } as $storage

    conditional {
      if ($storage == null) {
        var $s3_endpoint {
          value = "https://usc1.contabostorage.com"
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_s3_endpoint"
        } as $cfg_endpoint

        conditional {
          if ($cfg_endpoint != null && ($cfg_endpoint.valor|is_empty) == false) {
            var.update $s3_endpoint {
              value = $cfg_endpoint.valor|trim
            }
          }
        }

        var $s3_bucket {
          value = "confvision"
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_s3_bucket"
        } as $cfg_bucket

        conditional {
          if ($cfg_bucket != null && ($cfg_bucket.valor|is_empty) == false) {
            var.update $s3_bucket {
              value = $cfg_bucket.valor|trim
            }
          }
        }

        var $s3_tenant_id {
          value = ""
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_s3_tenant_id"
        } as $cfg_tenant

        conditional {
          if ($cfg_tenant != null && ($cfg_tenant.valor|is_empty) == false) {
            var.update $s3_tenant_id {
              value = $cfg_tenant.valor|trim
            }
          }
        }

        var $s3_access_key {
          value = ""
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_s3_access_key"
        } as $cfg_access

        conditional {
          if ($cfg_access != null && ($cfg_access.valor|is_empty) == false) {
            var.update $s3_access_key {
              value = $cfg_access.valor|trim
            }
          }
        }

        var $s3_secret_key {
          value = ""
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_s3_secret_key"
        } as $cfg_secret

        conditional {
          if ($cfg_secret != null && ($cfg_secret.valor|is_empty) == false) {
            var.update $s3_secret_key {
              value = $cfg_secret.valor|trim
            }
          }
        }

        var $segmento_minutos {
          value = 5
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_s3_segmento_minutos"
        } as $cfg_segmento

        conditional {
          if ($cfg_segmento != null && ($cfg_segmento.valor|is_empty) == false) {
            var.update $segmento_minutos {
              value = $cfg_segmento.valor|to_int
            }
          }
        }

        precondition (($s3_access_key|is_empty) == false && ($s3_secret_key|is_empty) == false) {
          error = "Credenciais S3 ConfVision nao configuradas — cadastre confvision_s3_access_key e confvision_s3_secret_key em Config Financeiro"
        }

        db.add vis_gravacao_storage {
          enforce_hidden_fields = false
          data = {
            created_at      : "now"
            id_franqueado   : $input.id_franqueado
            s3_endpoint     : $s3_endpoint
            s3_bucket       : $s3_bucket
            s3_tenant_id    : $s3_tenant_id
            s3_access_key   : $s3_access_key
            s3_secret_key   : $s3_secret_key
            segmento_minutos: $segmento_minutos
            status          : "ativo"
            provisionado_em : "now"
            observacao      : "Provisionado automaticamente (fn_vis_gravacao_storage_ensure)"
          }
        } as $novo

        var.update $storage {
          value = $novo
        }
      }
    }
  }

  response = $storage
}
