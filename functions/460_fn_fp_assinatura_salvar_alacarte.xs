// Salva ou atualiza assinatura a la carte (primeira contratacao, pendente ou plano ativo)
function fn_fp_assinatura_salvar_alacarte {
  input {
    text id_franqueado? filters=trim
    text produto?=franqueadopro filters=trim
    text plano? filters=trim
    json addons?
    text origem?=franqueado filters=trim
    text admin_usuario? filters=trim
    text cupom_codigo? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    precondition (($input.plano|is_empty) == false) {
      error = "plano base obrigatorio"
    }
  
    function.run fn_fp_assinatura_get_ativa {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : $input.produto
      }
    } as $check
  
    var $assinatura {
      value = $check.assinatura
    }
  
    conditional {
      if ($assinatura != null && ($assinatura.plano|is_empty) == false) {
        precondition ($input.plano == $assinatura.plano) {
          error = "Plano base fixo. Voce so pode adicionar ou remover modulos extras."
        }
      }
    }
  
    var $lista_addons {
      value = $input.addons|first_notempty:[]
    }
  
    function.run fn_fp_alacarte_addons_normalizar {
      input = {produto: $input.produto, addons: $lista_addons}
    } as $norm_addons
  
    var.update $lista_addons {
      value = $norm_addons.addons|first_notempty:[]
    }
  
    var $id_cen_ala {
      value = ""
    }
  
    conditional {
      if ($assinatura != null) {
        var.update $id_cen_ala {
          value = $assinatura.id_central|first_notempty:""|trim
        }
      }
    }
  
    conditional {
      if ($id_cen_ala|is_empty) {
        function.run fn_fp_franqueado_id_central {
          input = {id_franqueado: $input.id_franqueado}
        } as $cen_ala
      
        var.update $id_cen_ala {
          value = $cen_ala.id_central|first_notempty:"CENTRAL"|trim
        }
      }
    }
  
    function.run fn_fp_alacarte_calcular {
      input = {
        produto   : $input.produto
        plano     : $input.plano
        id_central: $id_cen_ala
        addons    : $lista_addons
      }
    } as $calc
  
    var $acao {
      value = "contratacao"
    }
  
    var $fatura_resultado {
      value = null
    }
  
    var $cancel_resultado {
      value = null
    }
  
    conditional {
      if ($check.liberado && $assinatura != null) {
        function.run fn_fp_alacarte_addons_efetivos {
          input = {
            produto              : $input.produto
            plano                : $assinatura.plano
            modulos_snapshot     : $assinatura.modulos_json
            addons_json          : $assinatura.addons_json
            addons_pendentes_json: $assinatura.addons_pendentes_json
          }
        } as $efetivos_info
      
        var $efetivos {
          value = $efetivos_info.addons_efetivos|first_notempty:[]
        }
      
        function.run fn_fp_alacarte_addons_normalizar {
          input = {
            produto: $input.produto
            addons : $assinatura.addons_json|first_notempty:[]
          }
        } as $norm_base
      
        var $baseline_addons {
          value = $norm_base.addons|first_notempty:[]
        }
      
        var $novos_cobrados {
          value = $calc.addons_cobrados|first_notempty:[]
        }
      
        var $addons_adicionar {
          value = []
        }
      
        foreach ($novos_cobrados) {
          each as $chave_nova {
            var $ja_tem {
              value = false
            }
          
            foreach ($efetivos) {
              each as $chave_ef {
                conditional {
                  if ($chave_ef == $chave_nova) {
                    var.update $ja_tem {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($ja_tem == false) {
                var.update $addons_adicionar {
                  value = $addons_adicionar|push:$chave_nova
                }
              }
            }
          }
        }
      
        var $addons_json_igual {
          value = true
        }
      
        var $tem_remocao {
          value = false
        }
      
        conditional {
          if (($baseline_addons|count) != ($lista_addons|count)) {
            var.update $addons_json_igual {
              value = false
            }
          }
        }
      
        conditional {
          if ($addons_json_igual) {
            foreach ($lista_addons) {
              each as $ch {
                var $achou {
                  value = false
                }
              
                foreach ($baseline_addons) {
                  each as $ba {
                    conditional {
                      if ($ba == $ch) {
                        var.update $achou {
                          value = true
                        }
                      }
                    }
                  }
                }
              
                conditional {
                  if ($achou == false) {
                    var.update $addons_json_igual {
                      value = false
                    }
                  }
                }
              }
            }
          }
        }
      
        conditional {
          if ($addons_json_igual) {
            foreach ($baseline_addons) {
              each as $ch_base {
                var $achou_base {
                  value = false
                }
              
                foreach ($lista_addons) {
                  each as $ch_sel {
                    conditional {
                      if ($ch_sel == $ch_base) {
                        var.update $achou_base {
                          value = true
                        }
                      }
                    }
                  }
                }
              
                conditional {
                  if ($achou_base == false) {
                    var.update $addons_json_igual {
                      value = false
                    }
                  }
                }
              }
            }
          }
        }
      
        foreach ($efetivos) {
          each as $ch_ef {
            var $achou_sel {
              value = false
            }
          
            foreach ($lista_addons) {
              each as $ch_sel {
                conditional {
                  if ($ch_sel == $ch_ef) {
                    var.update $achou_sel {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($achou_sel == false) {
                var.update $tem_remocao {
                  value = true
                }
              }
            }
          }
        }
      
        precondition ($addons_json_igual == false || ($addons_adicionar|count) > 0 || $tem_remocao) {
          error = "Nenhuma alteracao no plano personalizado."
        }
      
        function.run fn_fp_fatura_cancelar_abertas_assinatura {
          input = {
            assinatura_id: $assinatura.id
            observacao   : "Cancelada — alteracao do plano personalizado"
            origem       : $input.origem|first_notempty:"franqueado"
            admin_usuario: $input.admin_usuario
          }
        } as $cancel_resultado
      
        var $pendentes_merge {
          value = []
        }
      
        foreach ($lista_addons) {
          each as $ch_sel {
            foreach ($assinatura.addons_pendentes_json|first_notempty:[]) {
              each as $ch_pend {
                conditional {
                  if ($ch_sel == $ch_pend) {
                    var.update $pendentes_merge {
                      value = $pendentes_merge|push:$ch_sel
                    }
                  }
                }
              }
            }
          }
        }
      
        var $valor_fatura {
          value = 0
        }
      
        var $descricao_fatura {
          value = ""
        }
      
        conditional {
          if (($addons_adicionar|count) > 0) {
            var.update $acao {
              value = "adicionar_modulos"
            }
          
            foreach ($addons_adicionar) {
              each as $ch_add {
                db.query fp_modulo_catalogo {
                  where = $db.fp_modulo_catalogo.id_central == "CENTRAL" && $db.fp_modulo_catalogo.id_representante == "" && $db.fp_modulo_catalogo.produto == $input.produto && $db.fp_modulo_catalogo.chave == $ch_add && $db.fp_modulo_catalogo.ativo == "S"
                  return = {type: "single"}
                } as $item_add
              
                conditional {
                  if ($item_add != null) {
                    var $preco_add {
                      value = $item_add.valor_mensal|first_notempty:0
                    }
                  
                    var.update $valor_fatura {
                      value = ($valor_fatura + $preco_add)|round:2
                    }
                  
                    var.update $pendentes_merge {
                      value = $pendentes_merge|push:$ch_add
                    }
                  }
                }
              }
            }
          
            var.update $descricao_fatura {
              value = "Modulos extras do plano personalizado"
            }
          }
        
          else {
            var.update $acao {
              value = "reduzir_proxima_fatura"
            }
          }
        }
      
        var $vencimento_addon {
          value = now|add_secs_to_timestamp:5 * 86400
        }
      
        var $ciclo_addon {
          value = "ALAC-" ~ ($assinatura.id|to_text) ~ "-" ~ (now|format_timestamp:"YmdHis":"UTC")
        }
      
        conditional {
          if (($addons_adicionar|count) > 0 && $valor_fatura > 0) {
            function.run fn_fp_fatura_gerar {
              input = {
                assinatura_id      : $assinatura.id
                origem             : $input.origem|first_notempty:"franqueado"
                admin_usuario      : $input.admin_usuario
                acao_log           : "fatura_gerar_alacarte_addon"
                valor_fatura       : $valor_fatura
                descricao_extra    : $descricao_fatura
                ciclo_ref          : $ciclo_addon
                vencimento_em      : $vencimento_addon
                atualizar_ciclo_ref: false
              }
            } as $fatura_resultado
          
            precondition ($fatura_resultado.criada) {
              error = "Nao foi possivel gerar a fatura dos modulos extras. Tente novamente ou contate a central ConfMonit."
            }
          }
        }
      
        db.patch fp_assinatura_produto {
          field_name = "id"
          field_value = $assinatura.id
          data = {
            valor                 : $calc.valor_total
            limites_json          : $calc.limites_json
            fp_produto_catalogo_id: $calc.fp_produto_catalogo_id
            tipo_contratacao      : "alacarte"
            addons_json           : $lista_addons
            addons_pendentes_json : $pendentes_merge
            observacao            : "Plano personalizado a la carte"
          }
        } as $assinatura
      }
    
      else {
        conditional {
          if ($assinatura != null) {
            function.run fn_fp_fatura_cancelar_abertas_assinatura {
              input = {
                assinatura_id: $assinatura.id
                observacao   : "Cancelada — nova contratacao do plano personalizado"
                origem       : $input.origem|first_notempty:"franqueado"
                admin_usuario: $input.admin_usuario
              }
            } as $cancel_resultado
          }
        }
      
        conditional {
          if ($assinatura != null) {
            db.patch fp_assinatura_produto {
              field_name = "id"
              field_value = $assinatura.id
              data = {
                plano                 : $input.plano
                status                : "pendente"
                valor                 : $calc.valor_total
                limites_json          : $calc.limites_json
                modulos_json          : $calc.modulos_json
                fp_produto_catalogo_id: $calc.fp_produto_catalogo_id
                tipo_contratacao      : "alacarte"
                addons_json           : $lista_addons
                addons_pendentes_json : []
                valido_ate            : null
                proxima_cobranca_em   : null
                ciclo_fatura_ref      : ""
                observacao            : "Plano personalizado a la carte"
              }
            } as $assinatura
          }
        
          else {
            db.add fp_assinatura_produto {
              data = {
                created_at            : "now"
                id_franqueado         : $input.id_franqueado
                produto               : $input.produto
                plano                 : $input.plano
                status                : "pendente"
                periodicidade         : "mensal"
                valor                 : $calc.valor_total
                limites_json          : $calc.limites_json
                modulos_json          : $calc.modulos_json
                fp_produto_catalogo_id: $calc.fp_produto_catalogo_id
                tipo_contratacao      : "alacarte"
                addons_json           : $lista_addons
                addons_pendentes_json : []
                observacao            : "Plano personalizado a la carte"
              }
            } as $assinatura
          }
        }
      
        function.run fn_fp_fatura_gerar {
          input = {
            assinatura_id: $assinatura.id
            origem       : $input.origem|first_notempty:"franqueado"
            admin_usuario: $input.admin_usuario
            acao_log     : "fatura_gerar_contratacao_alacarte"
          }
        } as $fatura_resultado
      }
    }
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "assinatura_salvar_alacarte"
        id_franqueado: $input.id_franqueado
        ref_tipo     : "fp_assinatura_produto"
        ref_id       : $assinatura.id|to_text
        detalhe      : $acao ~ " " ~ $input.plano
        origem       : $input.origem|first_notempty:"franqueado"
        valor        : $calc.valor_total
        produto      : $input.produto
        plano        : $input.plano
        admin_usuario: $input.admin_usuario
      }
    } as $log
  
    var $cupom_resultado {
      value = null
    }
  
    conditional {
      if ((($input.cupom_codigo|is_empty) == false) && $fatura_resultado != null && $fatura_resultado.criada && $fatura_resultado.fatura != null) {
        function.run fn_fp_cupom_aplicar_fatura {
          input = {
            codigo       : $input.cupom_codigo
            id_franqueado: $input.id_franqueado
            fatura_id    : $fatura_resultado.fatura.id
            origem       : $input.origem|first_notempty:"franqueado"
            admin_usuario: $input.admin_usuario
          }
        } as $cupom_resultado
      }
    }
  }

  response = {
    assinatura         : $assinatura
    calculo            : $calc
    acao               : $acao
    fatura             : $fatura_resultado.fatura
    fatura_criada      : $fatura_resultado.criada
    faturas_canceladas : $cancel_resultado.canceladas
    cupom              : $cupom_resultado
    registro_financeiro: $log
  }
}