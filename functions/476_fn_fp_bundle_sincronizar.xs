// Sincroniza assinaturas bundle (valor 0) conforme plano FranqueadoPro ativo/suspenso
function fn_fp_bundle_sincronizar {
  input {
    text id_franqueado? filters=trim
    text admin_usuario? filters=trim
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
  
    var $criadas {
      value = 0
    }
  
    var $atualizadas {
      value = 0
    }
  
    var $suspensas {
      value = 0
    }
  
    var $itens_desejados {
      value = []
    }
  
    var $fp_ativa {
      value = $fp.liberado
    }
  
    conditional {
      if ($fp_ativa && $fp.assinatura != null) {
        function.run fn_fp_bundle_mapa {
          input = {plano_fp: $fp.assinatura.plano}
        } as $mapa
      
        var.update $itens_desejados {
          value = $mapa.itens
        }
      }
    }
  
    // Suspende bundles que nao estao mais no mapa (downgrade / FP inativo)
    db.query fp_assinatura_produto {
      where = $db.fp_assinatura_produto.id_franqueado == $input.id_franqueado && $db.fp_assinatura_produto.status == "ativa"
      return = {type: "list"}
    } as $todas_ativas
  
    foreach ($todas_ativas) {
      each as $a {
        var $eh_bundle {
          value = (($a.observacao|first_notempty:"")|contains:"bundle_fp")
        }
      
        conditional {
          if ($eh_bundle && $a.produto != "franqueadopro") {
            var $ainda_incluso {
              value = false
            }
          
            foreach ($itens_desejados) {
              each as $d {
                conditional {
                  if ($d.produto == $a.produto) {
                    var.update $ainda_incluso {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($ainda_incluso == false) {
                db.patch fp_assinatura_produto {
                  field_name = "id"
                  field_value = $a.id
                  data = {
                    status             : "suspensa"
                    proxima_cobranca_em: null
                    observacao         : "bundle_fp|suspensa_sync"
                  }
                } as $sus
              
                var.update $suspensas {
                  value = $suspensas + 1
                }
              }
            }
          }
        }
      }
    }
  
    // Cria/atualiza bundles desejados
    foreach ($itens_desejados) {
      each as $item {
        function.run fn_fp_catalogo_get_por_plano {
          input = {
            produto   : $item.produto
            plano     : $item.plano
            id_central: $fp.assinatura.id_central|first_notempty:""
          }
        } as $cat
      
        conditional {
          if ($cat != null) {
            db.query fp_assinatura_produto {
              where = $db.fp_assinatura_produto.id_franqueado == $input.id_franqueado && $db.fp_assinatura_produto.produto == $item.produto
              sort = {fp_assinatura_produto.created_at: "desc"}
              return = {type: "single"}
            } as $exist
          
            var $pular_avulsa {
              value = false
            }
          
            conditional {
              if ($exist != null && $exist.status == "ativa") {
                var $obs {
                  value = $exist.observacao|first_notempty:""
                }
              
                conditional {
                  if (($obs|contains:"bundle_fp") == false && ($exist.valor|first_notempty:0) > 0) {
                    var.update $pular_avulsa {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($pular_avulsa == false) {
                var $valido {
                  value = null
                }
              
                conditional {
                  if ($fp.assinatura != null) {
                    var.update $valido {
                      value = $fp.assinatura.valido_ate
                    }
                  }
                }
              
                var $obs_bundle {
                  value = "bundle_fp|plano_fp=" ~ ($fp.assinatura.plano|first_notempty:"") ~ "|incluso=" ~ $item.plano
                }
              
                conditional {
                  if ($exist == null) {
                    db.add fp_assinatura_produto {
                      data = {
                        created_at            : "now"
                        id_franqueado         : $input.id_franqueado
                        produto               : $item.produto
                        plano                 : $item.plano
                        status                : "ativa"
                        periodicidade         : "mensal"
                        valor                 : 0
                        valido_ate            : $valido
                        proxima_cobranca_em   : null
                        limites_json          : $cat.limites_json
                        modulos_json          : $cat.modulos_json
                        fp_produto_catalogo_id: $cat.id
                        tipo_contratacao      : "pacote"
                        observacao            : $obs_bundle
                      }
                    } as $nova
                  
                    var.update $criadas {
                      value = $criadas + 1
                    }
                  
                    function.run fn_fp_financeiro_log {
                      input = {
                        acao         : "bundle_ativar"
                        id_franqueado: $input.id_franqueado
                        ref_tipo     : "fp_assinatura_produto"
                        ref_id       : $nova.id|to_text
                        detalhe      : $obs_bundle
                        origem       : "sistema"
                        valor        : 0
                        produto      : $item.produto
                        plano        : $item.plano
                        admin_usuario: $input.admin_usuario
                      }
                    } as $log_cria
                  }
                
                  else {
                    db.patch fp_assinatura_produto {
                      field_name = "id"
                      field_value = $exist.id
                      data = ```
                        {
                          plano                 : $item.plano
                          status                : "ativa"
                          valor                 : 0
                          valido_ate            : $valido
                          proxima_cobranca_em   : null
                          limites_json          : $cat.limites_json
                          modulos_json          : $cat.modulos_json
                          fp_produto_catalogo_id: $cat.id
                          tipo_contratacao      : "pacote"
                          observacao            : $obs_bundle
                          ciclo_fatura_ref      : ""
                        }
                        ```
                    } as $upd
                  
                    var.update $atualizadas {
                      value = $atualizadas + 1
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  response = {
    id_franqueado: $input.id_franqueado
    fp_liberado  : $fp_ativa
    criadas      : $criadas
    atualizadas  : $atualizadas
    suspensas    : $suspensas
    itens        : $itens_desejados
  }
}