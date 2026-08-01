// Lista produtos do ecossistema com status (incluso / comprar / ativo) — FranqueadoPro
query fp_ecossistema_listar verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    function.run fn_fp_assinatura_get_ativa {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : "franqueadopro"
      }
    } as $fp
  
    var $plano_fp {
      value = ""
    }
  
    conditional {
      if ($fp.assinatura != null) {
        var.update $plano_fp {
          value = $fp.assinatura.plano|first_notempty:""
        }
      }
    }
  
    function.run fn_fp_bundle_mapa {
      input = {plano_fp: $plano_fp}
    } as $mapa
  
    function.run fn_fp_franqueado_id_representante {
      input = {id_franqueado: $input.id_franqueado}
    } as $rep
  
    function.run fn_fp_catalogo_com_preco_rep {
      input = {
        id_representante: $rep.id_representante
        ativo           : "S"
      }
    } as $cat_preco
  
    var $catalogo {
      value = $cat_preco.dados|first_notempty:[]
    }
  
    var $produtos_ids {
      value = [
        "webterminal"
        "terminalmovel"
        "confvision"
        "webambiente"
        "dialyze"
      ]
    }
  
    var $dados {
      value = []
    }
  
    foreach ($produtos_ids) {
      each as $pid {
        function.run fn_fp_produto_acesso_efetivo {
          input = {id_franqueado: $input.id_franqueado, produto: $pid}
        } as $acesso
      
        var $planos_cat {
          value = []
        }
      
        foreach ($catalogo) {
          each as $c {
            conditional {
              if ($c.produto == $pid) {
                array.push $planos_cat {
                  value = {
                    plano        : $c.plano
                    nome_exibicao: $c.nome_exibicao
                    valor_mensal : $c.valor_venda|first_notempty:$c.valor_mensal
                  }
                }
              }
            }
          }
        }
      
        var $plano_incluso {
          value = ""
        }
      
        foreach ($mapa.itens) {
          each as $mi {
            conditional {
              if ($mi.produto == $pid) {
                var.update $plano_incluso {
                  value = $mi.plano
                }
              }
            }
          }
        }
      
        var $status {
          value = "disponivel"
        }
      
        conditional {
          if ($acesso.liberado && $acesso.origem == "bundle_fp") {
            var.update $status {
              value = "incluso"
            }
          }
        
          elseif ($acesso.liberado && $acesso.origem == "assinatura") {
            var.update $status {
              value = "ativo"
            }
          }
        
          elseif ($fp.liberado == false) {
            var.update $status {
              value = "disponivel"
            }
          }
        }
      
        var $pode_contratar {
          value = true
        }
      
        conditional {
          if ($status == "incluso") {
            var.update $pode_contratar {
              value = false
            }
          }
        
          elseif ($status == "ativo") {
            var.update $pode_contratar {
              value = false
            }
          }
        }
      
        // Upgrade: WT lite incluso pode comprar pro_plus; WA pro incluso pode comprar pro_plus
        conditional {
          if ($status == "incluso") {
            conditional {
              if ($pid == "webterminal" && $plano_incluso == "pro") {
                var.update $pode_contratar {
                  value = true
                }
              }
            
              elseif ($pid == "webterminal" && $plano_incluso == "lite") {
                var.update $pode_contratar {
                  value = true
                }
              }
            
              elseif ($pid == "webambiente" && $plano_incluso == "pro") {
                var.update $pode_contratar {
                  value = true
                }
              }
            }
          }
        }
      
        array.push $dados {
          value = {
            produto         : $pid
            status          : $status
            liberado        : $acesso.liberado
            origem          : $acesso.origem
            plano_efetivo   : $acesso.plano
            plano_incluso_fp: $plano_incluso
            pode_contratar  : $pode_contratar
            planos          : $planos_cat
            assinatura      : $acesso.assinatura
            motivo          : $acesso.motivo
          }
        }
      }
    }
  }

  response = {
    id_franqueado: $input.id_franqueado
    franqueadopro: ```
      {
        liberado: $fp.liberado
        plano   : $plano_fp
        motivo  : $fp.motivo
      }
      ```
    bundle_itens : $mapa.itens
    dados        : $dados
    total        : $dados|count
  }
}