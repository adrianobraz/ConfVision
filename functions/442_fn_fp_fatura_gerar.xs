// Gera fatura de assinatura para o ciclo atual (idempotente). Abate credito_saldo de cupom se houver.
function fn_fp_fatura_gerar {
  input {
    int assinatura_id? filters=min:1
    text ciclo_ref? filters=trim
    text origem? filters=trim
    text admin_usuario? filters=trim
    text acao_log? filters=trim
    decimal valor_fatura?
    text descricao_extra? filters=trim
    timestamp? vencimento_em?
    bool atualizar_ciclo_ref?=true
  }

  stack {
    db.get fp_assinatura_produto {
      field_name = "id"
      field_value = $input.assinatura_id
    } as $assinatura
  
    precondition ($assinatura != null) {
      error = "Assinatura nao encontrada"
    }
  
    precondition ($assinatura.status == "ativa" || $assinatura.status == "pendente") {
      error = "Assinatura precisa estar pendente ou ativa para gerar fatura"
    }
  
    conditional {
      if (($assinatura.ciclo_fatura_ref|is_empty) == false) {
        db.query fp_fatura {
          where = ($db.fp_fatura.id_franqueado == $assinatura.id_franqueado) && ($db.fp_fatura.ciclo_ref == $assinatura.ciclo_fatura_ref) && ($db.fp_fatura.status != "cancelada")
          return = {type: "exists"}
        } as $fatura_ciclo_ativa
      
        conditional {
          if ($fatura_ciclo_ativa == null) {
            db.patch fp_assinatura_produto {
              field_name = "id"
              field_value = $assinatura.id
              data = {ciclo_fatura_ref: ""}
            } as $assinatura
          }
        }
      }
    }
  
    var $vencimento {
      value = $input.vencimento_em
        |first_notempty:$assinatura.proxima_cobranca_em
    }
  
    conditional {
      if ($vencimento == null) {
        var.update $vencimento {
          value = now|add_secs_to_timestamp:5 * 86400
        }
      }
    }
  
    var $ciclo {
      value = $input.ciclo_ref
    }
  
    conditional {
      if ($ciclo|is_empty) {
        function.run fn_fp_assinatura_ciclo_ref {
          input = {
            assinatura_id: $assinatura.id
            data_cobranca: $vencimento
          }
        } as $ciclo_gerado
      
        var.update $ciclo {
          value = $ciclo_gerado
        }
      }
    }
  
    db.query fp_fatura {
      where = ($db.fp_fatura.id_franqueado == $assinatura.id_franqueado)
      return = {type: "list"}
    } as $faturas_franq
  
    var $dup {
      value = null
    }
  
    foreach ($faturas_franq) {
      each as $f {
        conditional {
          if ($dup == null && $f.ciclo_ref == $ciclo && $f.status == "aberta") {
            var.update $dup {
              value = $f
            }
          }
        }
      }
    }
  
    var $criada {
      value = false
    }
  
    var $fatura {
      value = $dup
    }
  
    var $item {
      value = null
    }
  
    conditional {
      if ($dup != null) {
        var.update $criada {
          value = false
        }
      }
    
      else {
        var $valor_cheio {
          value = $input.valor_fatura
            |first_notempty:($assinatura.valor|first_notempty:0)
        }
      
        var $credito {
          value = $assinatura.credito_saldo|first_notempty:0
        }
      
        var $credito_usado {
          value = 0
        }
      
        var $valor {
          value = $valor_cheio
        }
      
        conditional {
          if ($credito > 0 && $valor_cheio > 0) {
            conditional {
              if ($credito >= $valor_cheio) {
                var.update $credito_usado {
                  value = $valor_cheio
                }
              
                var.update $valor {
                  value = 0
                }
              }
            
              else {
                var.update $credito_usado {
                  value = $credito
                }
              
                var.update $valor {
                  value = ($valor_cheio - $credito)|round:2
                }
              }
            }
          
            var $saldo_restante {
              value = ($credito - $credito_usado)|round:2
            }
          
            db.patch fp_assinatura_produto {
              field_name = "id"
              field_value = $assinatura.id
              data = {credito_saldo: $saldo_restante}
            } as $ass_cred
          }
        }
      
        var $desc_item {
          value = "Assinatura " ~ $assinatura.produto ~ " — plano " ~ $assinatura.plano
        }
      
        conditional {
          if (($input.descricao_extra|is_empty) == false) {
            var.update $desc_item {
              value = $input.descricao_extra
            }
          }
        }
      
        var $obs_fatura {
          value = "Fatura automatica " ~ $assinatura.produto ~ " " ~ $assinatura.plano
        }
      
        conditional {
          if ($input.origem == "admin") {
            var.update $obs_fatura {
              value = "Fatura manual admConfmonit — " ~ $assinatura.produto ~ " " ~ $assinatura.plano
            }
          }
        
          elseif ($input.origem == "franqueado") {
            var.update $obs_fatura {
              value = "Contratacao FranqueadoPro — plano " ~ $assinatura.plano ~ ". Efetue o pagamento conforme instrucoes da central ConfMonit."
            }
          }
        }
      
        conditional {
          if ($credito_usado > 0) {
            var.update $obs_fatura {
              value = $obs_fatura ~ " | Credito cupom -R$ " ~ ($credito_usado|to_text)
            }
          }
        }
      
        var $referencia {
          value = "FAT-" ~ ($vencimento|format_timestamp:"YmdHis":"UTC") ~ "-" ~ $assinatura.id_franqueado
        }
      
        db.add fp_fatura {
          data = {
            created_at             : "now"
            id_franqueado          : $assinatura.id_franqueado
            id_representante       : $assinatura.id_representante
            id_central             : $assinatura.id_central
            referencia             : $referencia
            status                 : "aberta"
            tipo                   : "assinatura"
            valor_total            : $valor
            valor_piso_central     : $assinatura.valor_piso_central
            valor_piso_breakglass  : $assinatura.valor_piso_breakglass
            margem_central         : $assinatura.margem_central
            margem_rep             : $assinatura.margem_rep
            vencimento_em          : $vencimento
            ciclo_ref              : $ciclo
            observacao             : $obs_fatura
          }
        } as $nova_fatura
      
        db.add fp_fatura_item {
          data = {
            created_at    : "now"
            fp_fatura_id  : $nova_fatura.id
            descricao     : $desc_item
            quantidade    : 1
            valor_unitario: $valor_cheio
            valor_total   : $valor_cheio
            ref_tipo      : "fp_assinatura_produto"
            ref_id        : $assinatura.id|to_text
          }
        } as $novo_item
      
        conditional {
          if ($credito_usado > 0) {
            db.add fp_fatura_item {
              data = {
                created_at    : "now"
                fp_fatura_id  : $nova_fatura.id
                descricao     : "Credito de cupom de desconto"
                quantidade    : 1
                valor_unitario: (0 - $credito_usado)
                valor_total   : (0 - $credito_usado)
                ref_tipo      : "fp_cupom_credito"
                ref_id        : $assinatura.id|to_text
              }
            } as $item_credito
          
            function.run fn_fp_financeiro_log {
              input = {
                acao         : "cupom_credito_abatido"
                id_franqueado: $assinatura.id_franqueado
                ref_tipo     : "fp_fatura"
                ref_id       : $nova_fatura.id|to_text
                detalhe      : "credito_usado=" ~ ($credito_usado|to_text) ~ " valor_final=" ~ ($valor|to_text)
                origem       : $input.origem|first_notempty:"sistema"
                valor        : $credito_usado
                produto      : $assinatura.produto
                plano        : $assinatura.plano
                admin_usuario: $input.admin_usuario
              }
            } as $log_cred
          }
        }
      
        conditional {
          if ($input.atualizar_ciclo_ref != false) {
            db.patch fp_assinatura_produto {
              field_name = "id"
              field_value = $assinatura.id
              data = {ciclo_fatura_ref: $ciclo}
            } as $ass_upd
          }
        }
      
        var.update $fatura {
          value = $nova_fatura
        }
      
        var.update $item {
          value = $novo_item
        }
      
        var.update $criada {
          value = true
        }
      
        var $acao {
          value = $input.acao_log|first_notempty:"fatura_gerar"
        }
      
        function.run fn_fp_financeiro_log {
          input = {
            acao         : $acao
            id_franqueado: $assinatura.id_franqueado
            ref_tipo     : "fp_fatura"
            ref_id       : $nova_fatura.id|to_text
            detalhe      : $ciclo
            origem       : $input.origem|first_notempty:"sistema"
            valor        : $valor
            produto      : $assinatura.produto
            plano        : $assinatura.plano
            admin_usuario: $input.admin_usuario
          }
        } as $log
      }
    }
  }

  response = {
    fatura   : $fatura
    item     : $item
    criada   : $criada
    ciclo_ref: $ciclo
  }
}