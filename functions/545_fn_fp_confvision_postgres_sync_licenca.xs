// Cliente ConfCam /ops/vis_licenca — Postgres como fonte unica de licenca
function fn_fp_confvision_postgres_sync_licenca {
  input {
    text id_franqueado? filters=trim
    int vis_licenca_id?
    int fp_fatura_id?
    int fp_pagamento_id?
    timestamp? pago_em?
    text status? filters=trim
    text unidade? filters=trim
    int dias_antecedencia?=30 filters=min:1
    bool incluir_expiradas_vinculadas?=true
    text acao?=ativar_pagamento filters=trim
    object[] licencas? {
      schema {
        text id_franqueado? filters=trim
        text plano? filters=trim
        text unidade? filters=trim
        decimal valor?
        text status? filters=trim
        text observacao? filters=trim
      }
    }
  }

  stack {
    function.run fn_cvg_config_valor {
      input = {
        chave   : "confvision_api_url"
        fallback: "https://vision.confmonit2.com.br"
      }
    } as $api_url_cfg

    function.run fn_cvg_config_valor {
      input = {
        chave   : "confvision_ops_base_url"
        fallback: ""
      }
    } as $ops_url_cfg

    var $base {
      value = $api_url_cfg|trim
    }

    conditional {
      if (($base|is_empty) == true && ($ops_url_cfg|is_empty) == false) {
        var.update $base {
          value = $ops_url_cfg|trim
        }
      }
    }

    conditional {
      if (($base|is_empty) == true) {
        var.update $base {
          value = "https://vision.confmonit2.com.br"
        }
      }
    }

    function.run fn_cvg_config_valor {
      input = {
        chave   : "vis_worker_key"
        fallback: ""
      }
    } as $worker_key_cfg

    var $worker_key {
      value = $worker_key_cfg|trim
    }

    conditional {
      if (($worker_key|is_empty) == true) {
        function.run fn_cvg_config_valor {
          input = {
            chave   : "worker_secret"
            fallback: ""
          }
        } as $worker_secret_cfg

        var.update $worker_key {
          value = $worker_secret_cfg|trim
        }
      }
    }

    var $api_ok {
      value = false
    }

    var $api_resp {
      value = null
    }

    var $api_erro {
      value = ""
    }

    conditional {
      if (($base|is_empty) == true || ($worker_key|is_empty) == true) {
        var.update $api_erro {
          value = "sync_postgres_skip_config_ausente"
        }
      }

      else {
        try_catch {
          try {
            var $url {
              value = $base|rtrim:"/"
            }

            var $params {
              value = {}
            }

            conditional {
              if ($input.acao == "listar_franqueado") {
                var.update $url {
                  value = $url ~ "/ops/vis_licenca/listar_franqueado"
                }

                var.update $params {
                  value = {id_franqueado: $input.id_franqueado}
                }

                conditional {
                  if (($input.status|is_empty) == false) {
                    var.update $params {
                      value = $params|set:"status":$input.status
                    }
                  }
                }

                conditional {
                  if (($input.unidade|is_empty) == false) {
                    var.update $params {
                      value = $params|set:"unidade":$input.unidade
                    }
                  }
                }
              }

              elseif ($input.acao == "obter") {
                var.update $url {
                  value = $url ~ "/ops/vis_licenca/obter"
                }

                var.update $params {
                  value = {vis_licenca_id: $input.vis_licenca_id}
                }
              }

              elseif ($input.acao == "criar_lote") {
                var.update $url {
                  value = $url ~ "/ops/vis_licenca/criar_lote"
                }

                var.update $params {
                  value = {licencas: $input.licencas|first_notempty:[]}
                }
              }

              elseif ($input.acao == "expirar_vencidas") {
                var.update $url {
                  value = $url ~ "/ops/vis_licenca/expirar_vencidas"
                }
              }

              elseif ($input.acao == "listar_renovacao") {
                var.update $url {
                  value = $url ~ "/ops/vis_licenca/listar_renovacao"
                }

                var.update $params {
                  value = {
                    dias_antecedencia           : $input.dias_antecedencia|first_notempty:30
                    incluir_expiradas_vinculadas: $input.incluir_expiradas_vinculadas|first_notempty:true
                  }
                }

                conditional {
                  if (($input.id_franqueado|is_empty) == false) {
                    var.update $params {
                      value = $params|set:"id_franqueado":$input.id_franqueado
                    }
                  }
                }
              }

              else {
                var.update $url {
                  value = $url ~ "/ops/vis_licenca/ativar_pagamento"
                }

                var $pago {
                  value = $input.pago_em|first_notempty:now
                }

                var.update $params {
                  value = {
                    vis_licenca_id : $input.vis_licenca_id
                    fp_fatura_id   : $input.fp_fatura_id
                    fp_pagamento_id: $input.fp_pagamento_id
                    pago_em        : $pago
                  }
                }
              }
            }

            api.request {
              url = $url
              method = "POST"
              params = $params
              headers = []
                |push:"Content-Type: application/json"
                |push:"X-Vis-Worker-Key: " ~ $worker_key
              timeout = 30
              verify_host = false
              verify_peer = false
            } as $raw

            var.update $api_resp {
              value = $raw.response.result
            }

            var $ops_ok_flag {
              value = $api_resp|get:"ok":null
            }

            conditional {
              if ($ops_ok_flag == false) {
                var.update $api_erro {
                  value = $api_resp|get:"erro":"ops_retornou_ok_false"
                }
              }

              else {
                var.update $api_ok {
                  value = true
                }
              }
            }
          }

          catch {
            var.update $api_erro {
              value = "falha_ops_confvision_postgres"
            }
          }
        }
      }
    }
  }

  response = {
    ok      : $api_ok
    resposta: $api_resp
    erro    : $api_erro
  }
}
