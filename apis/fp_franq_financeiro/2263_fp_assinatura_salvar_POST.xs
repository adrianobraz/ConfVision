// Criar ou atualizar assinatura de produto
query fp_assinatura_salvar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int id?
    text id_franqueado? filters=trim
    text produto? filters=trim
    text plano? filters=trim
    text status? filters=trim
    text periodicidade? filters=trim
    decimal valor?
    timestamp? valido_ate?
    timestamp? proxima_cobranca_em?
    int fp_produto_catalogo_id?
    text observacao? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $admin_escopo
  
    var $admin_check {
      value = $admin_escopo|get:"admin":null
    }
  
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    function.run fn_fp_admin_assert_escopo_franqueado {
      input = {
        admin_token  : $input.admin_token
        id_franqueado: $input.id_franqueado
      }
    } as $escopo
  
    var $id_cen_ass {
      value = $admin_escopo|get:"idCentral":""|trim
    }
  
    var $id_rep_ass {
      value = $admin_escopo|get:"idRepresentante":""|trim
    }
  
    conditional {
      if (($id_rep_ass|is_empty) && (($input.id_franqueado|is_empty) == false)) {
        function.run fn_fp_franqueado_id_representante {
          input = {id_franqueado: $input.id_franqueado}
        } as $rep_fra
      
        var.update $id_rep_ass {
          value = $rep_fra.id_representante|first_notempty:""|trim
        }
      }
    }
  
    precondition (($input.produto|is_empty) == false) {
      error = "produto obrigatorio"
    }
  
    var $produto_norm {
      value = $input.produto|trim|to_lower
    }
  
    var $limites {
      value = null
    }
  
    var $modulos {
      value = null
    }
  
    var $valor {
      value = $input.valor
    }
  
    var $valor_piso_central {
      value = 0
    }
  
    var $valor_piso_breakglass {
      value = 0
    }
  
    var $margem_central {
      value = 0
    }
  
    var $margem_rep {
      value = 0
    }
  
    var $catalogo_id {
      value = $input.fp_produto_catalogo_id
    }
  
    var $cat {
      value = null
    }
  
    conditional {
      if ($catalogo_id != null && $catalogo_id > 0) {
        db.get fp_produto_catalogo {
          field_name = "id"
          field_value = $catalogo_id
        } as $cat
      }
    }
  
    // Busca no catalogo da Central da sessao (nunca default "CENTRAL")
    conditional {
      if ($cat == null && ($input.plano|is_empty) == false && ($id_cen_ass|is_empty) == false) {
        function.run fn_fp_catalogo_get_por_plano {
          input = {
            produto         : $produto_norm
            plano           : $input.plano
            id_central      : $id_cen_ass
            id_representante: ""
          }
        } as $cat
      }
    }
  
    conditional {
      if ($cat != null) {
        var.update $catalogo_id {
          value = $cat.id
        }
      
        conditional {
          if ($input.id == null || $input.id <= 0) {
            function.run fn_fp_assinatura_regras_efetivas {
              input = {
                produto               : $produto_norm
                plano                 : $input.plano
                modulos_snapshot      : $cat.modulos_json
                limites_snapshot      : $cat.limites_json
                fp_produto_catalogo_id: $cat.id
              }
            } as $regras
          
            var.update $limites {
              value = $regras.limites_json
            }
          
            var.update $modulos {
              value = $regras.modulos_json
            }
          
            function.run fn_fp_preco_efetivo {
              input = {
                id_franqueado         : $input.id_franqueado
                id_representante      : $id_rep_ass
                id_central            : $id_cen_ass
                produto               : $produto_norm
                plano                 : $input.plano
                fp_produto_catalogo_id: $cat.id
              }
            } as $preco
          
            conditional {
              if ($valor == null || $valor == 0) {
                var.update $valor {
                  value = $preco.valor_venda|first_notempty:($cat.valor_mensal|first_notempty:0)

                }
              }
            }
          
            var.update $valor_piso_central {
              value = $preco.valor_piso_central|first_notempty:($cat.valor_mensal|first_notempty:0)
            }
          
            var.update $valor_piso_breakglass {
              value = $preco.valor_piso_breakglass|first_notempty:($cat.valor_piso_breakglass|first_notempty:0)
            }
          
            var.update $margem_central {
              value = $preco.margem_central|first_notempty:0
            }
          
            var.update $margem_rep {
              value = $preco.margem_rep|first_notempty:0
            }
          
            conditional {
              if (($id_cen_ass|is_empty) && (($preco.id_central|trim)|is_empty) == false) {
                var.update $id_cen_ass {
                  value = $preco.id_central|trim
                }
              }
            }
          
            conditional {
              if (($id_rep_ass|is_empty) && (($preco.id_representante|trim)|is_empty) == false) {
                var.update $id_rep_ass {
                  value = $preco.id_representante|trim
                }
              }
            }

          }
        }
      }
    
      elseif (($input.plano|is_empty) == false) {
        conditional {
          if ($input.id == null || $input.id <= 0) {
            function.run fn_fp_assinatura_regras_efetivas {
              input = {produto: $produto_norm, plano: $input.plano}
            } as $regras
          
            var.update $limites {
              value = $regras.limites_json
            }
          
            var.update $modulos {
              value = $regras.modulos_json
            }
          }
        }
      }
    }
  
    precondition (($input.id != null && $input.id > 0) || (($valor|first_notempty:0) > 0)) {
      error = "Valor da assinatura ficou zero — confira o preco do plano no catalogo da sua Central"
    }
  
    conditional {
      if ($input.id != null && $input.id > 0) {
        db.patch fp_assinatura_produto {
          field_name = "id"
          field_value = $input.id
          data = ```
            {
              id_franqueado      : $input.id_franqueado
              produto            : $produto_norm
              plano              : $input.plano
              status             : $input.status|first_notempty:"pendente"
              periodicidade      : $input.periodicidade|first_notempty:"mensal"
              valido_ate         : $input.valido_ate
              proxima_cobranca_em: $input.proxima_cobranca_em
              observacao         : $input.observacao
            }
            ```
        } as $model
      }
    
      else {
        db.add fp_assinatura_produto {
          data = {
            created_at            : "now"
            id_franqueado         : $input.id_franqueado
            id_representante      : $id_rep_ass
            id_central            : $id_cen_ass
            produto               : $produto_norm
            plano                 : $input.plano
            status                : $input.status|first_notempty:"pendente"
            periodicidade         : $input.periodicidade|first_notempty:"mensal"
            valor                 : $valor
            valor_piso_central    : $valor_piso_central
            valor_piso_breakglass : $valor_piso_breakglass
            margem_central        : $margem_central
            margem_rep            : $margem_rep
            valido_ate            : $input.valido_ate
            proxima_cobranca_em   : $input.proxima_cobranca_em
            limites_json          : $limites
            modulos_json          : $modulos
            fp_produto_catalogo_id: $catalogo_id
            observacao            : $input.observacao
          }
        } as $model
      }
    }
  
    function.run fn_fp_financeiro_log {
      input = {
        acao            : "assinatura_salvar"
        id_franqueado   : $input.id_franqueado
        ref_tipo        : "fp_assinatura_produto"
        ref_id          : $model.id|to_text
        detalhe         : $produto_norm ~ "/" ~ $input.plano
        origem          : "admin"
        valor           : $valor
        produto         : $produto_norm
        plano           : $input.plano
        id_central      : $id_cen_ass
        id_representante: $id_rep_ass
        id_usuario      : $admin_escopo|get:"idUsuario":""
        admin_usuario: $input.admin_usuario

      }
    } as $log
  
    var $fatura_resultado {
      value = {criada: false, fatura: null}
    }
  
    conditional {
      if (($input.id == null || $input.id <= 0) && ($model.status|first_notempty:"pendente") == "pendente") {
        function.run fn_fp_fatura_gerar {
          input = {
            assinatura_id: $model.id
            origem       : "admin"
            admin_usuario: $input.admin_usuario
            acao_log     : "fatura_gerar_cadastro"
          }
        } as $fatura_resultado
      }
    }
  }

  response = {
    assinatura   : $model
    fatura       : $fatura_resultado.fatura
    fatura_criada: $fatura_resultado.criada
  }
}