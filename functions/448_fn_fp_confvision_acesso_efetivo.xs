// Define se franqueado deve ter UsaConfVision=S (plano ConfVision ativo/incluso + ou com licenca camera)
function fn_fp_confvision_acesso_efetivo {
  input {
    text id_franqueado? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    function.run fn_fp_confvision_reconciliar_licencas_pagas {
      input = {id_franqueado: $input.id_franqueado}
    } as $reconciliar
  
    var $liberado {
      value = false
    }
  
    var $motivo {
      value = "sem_acesso"
    }
  
    // Portal: assinatura ConfVision (avulsa ou bundle FP Pro+)
    function.run fn_fp_produto_acesso_efetivo {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : "confvision"
      }
    } as $cv_plano
  
    conditional {
      if ($cv_plano.liberado) {
        var.update $liberado {
          value = true
        }
      
        var.update $motivo {
          value = "plano_confvision_" ~ ($cv_plano.origem|first_notempty:"ativo")
        }
      }
    }
  
    // Fallback legado: licenca camera paga ainda libera portal
    conditional {
      if ($liberado == false) {
        db.query vis_licenca {
          where = $db.vis_licenca.id_franqueado == $input.id_franqueado && ($db.vis_licenca.status == "disponivel" || $db.vis_licenca.status == "em_uso") && $db.vis_licenca.pago_em != null && ($db.vis_licenca.valido_ate == null || $db.vis_licenca.valido_ate > now)
          return = {type: "list"}
        } as $licencas
      
        conditional {
          if (($licencas|count) > 0) {
            var.update $liberado {
              value = true
            }
          
            var.update $motivo {
              value = "licenca_ativa"
            }
          }
        }
      }
    }
  
    var $usa {
      value = "N"
    }
  
    conditional {
      if ($liberado) {
        var.update $usa {
          value = "S"
        }
      }
    }
  }

  response = {
    liberado      : $liberado
    motivo        : $motivo
    usa_confvision: $usa
    plano_origem  : $cv_plano.origem
    reconciliar   : $reconciliar
  }
}