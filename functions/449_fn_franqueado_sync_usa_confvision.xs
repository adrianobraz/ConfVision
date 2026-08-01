// Sincroniza UsaConfVision no MySQL legado: plano ConfVision, bundle FP Pro+ ou licenca camera
// Falha na API legada NAO deve impedir ativacao de assinatura / pagamento
function fn_franqueado_sync_usa_confvision {
  input {
    text id_franqueado? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    function.run fn_fp_confvision_acesso_efetivo {
      input = {id_franqueado: $input.id_franqueado}
    } as $acesso
  
    db.get fp_config_financeiro {
      field_name = "chave"
      field_value = "api_legada_url"
    } as $cfg_url
  
    var $api_base {
      value = "http://185.130.61.4:2010"
    }
  
    conditional {
      if ($cfg_url != null && ($cfg_url.valor|is_empty) == false) {
        var.update $api_base {
          value = $cfg_url.valor|trim
        }
      }
    }
  
    var $url {
      value = $api_base ~ "/v4/franqueado/setUsaConfVision"
    }
  
    var $api_resp {
      value = null
    }
  
    var $api_ok {
      value = false
    }
  
    var $api_erro {
      value = ""
    }
  
    try_catch {
      try {
        api.request {
          url = $url
          method = "POST"
          params = {}
            |set:"fraId":$input.id_franqueado
            |set:"usaConfVision":$acesso.usa_confvision
          headers = []
            |push:"Content-Type: application/json"
        } as $api_resp
      
        var.update $api_ok {
          value = true
        }
      }
    
      catch {
        var.update $api_erro {
          value = "falha_api_legada_setUsaConfVision"
        }
      
        var.update $api_resp {
          value = {erro: $api_erro}
        }
      }
    }
  
    var $detalhe_log {
      value = $acesso.motivo ~ " -> " ~ $acesso.usa_confvision
    }
  
    conditional {
      if ($api_ok == false) {
        var.update $detalhe_log {
          value = $detalhe_log ~ " | sync_falhou:" ~ $api_erro
        }
      }
    }
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "sync_usa_confvision"
        id_franqueado: $input.id_franqueado
        ref_tipo     : "franqueado"
        ref_id       : $input.id_franqueado
        detalhe      : $detalhe_log
        origem       : "sistema"
      }
    } as $log
  }

  response = {
    acesso  : $acesso
    api_resp: $api_resp
    api_ok  : $api_ok
    log     : $log
  }
}
