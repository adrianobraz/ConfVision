// Cria cupom hierarquico: CEN→REP ou REP→FRA
function fn_fp_cupom_criar {
  input {
    text codigo? filters=trim
    text tipo?=percentual filters=trim
    decimal valor?
    text produto?=franqueadopro filters=trim
    text id_franqueado? filters=trim
    text id_representante? filters=trim
    text criado_por_tipo? filters=trim
    text alvo_nivel? filters=trim
    text observacao? filters=trim
    text admin_usuario? filters=trim
    text id_central? filters=trim
  }

  stack {
    var $codigo {
      value = $input.codigo
        |first_notempty:""
        |trim
        |to_upper
    }
  
    precondition (($codigo|is_empty) == false) {
      error = "codigo obrigatorio"
    }
  
    var $tipo {
      value = $input.tipo
        |first_notempty:"percentual"
        |to_lower
    }
  
    precondition ($tipo == "percentual" || $tipo == "valor_fixo") {
      error = "tipo deve ser percentual ou valor_fixo"
    }
  
    var $valor {
      value = $input.valor|first_notempty:0
    }
  
    precondition ($valor > 0) {
      error = "valor do cupom deve ser maior que zero"
    }
  
    conditional {
      if ($tipo == "percentual") {
        precondition ($valor <= 100) {
          error = "percentual nao pode ser maior que 100"
        }
      }
    }
  
    var $produto {
      value = $input.produto
        |first_notempty:"franqueadopro"
        |to_lower
    }
  
    var $criado_por_tipo {
      value = $input.criado_por_tipo|first_notempty:""|to_upper
    }
  
    precondition ($criado_por_tipo == "CEN" || $criado_por_tipo == "REP") {
      error = "criado_por_tipo deve ser CEN ou REP"
    }
  
    var $alvo {
      value = $input.alvo_nivel|first_notempty:""|to_upper
    }
  
    conditional {
      if ($alvo|is_empty) {
        conditional {
          if ($criado_por_tipo == "CEN") {
            var.update $alvo {
              value = "REP"
            }
          }
        
          else {
            var.update $alvo {
              value = "FRA"
            }
          }
        }
      }
    }
  
    precondition (($criado_por_tipo == "CEN" && $alvo == "REP") || ($criado_por_tipo == "REP" && $alvo == "FRA")) {
      error = "Hierarquia invalida: Central cria para Representante; Representante cria para Franqueado"
    }
  
    var $id_rep {
      value = $input.id_representante|first_notempty:""
    }
  
    // REP: id_representante obrigatorio (dono do cupom)
    conditional {
      if ($criado_por_tipo == "REP") {
        precondition (($id_rep|is_empty) == false) {
          error = "id_representante obrigatorio para cupom do Representante"
        }
      }
    }
  
    db.query fp_cupom_desconto {
      where = $db.fp_cupom_desconto.codigo == $codigo
      return = {type: "exists"}
    } as $existe
  
    precondition ($existe == null) {
      error = "Ja existe cupom com este codigo"
    }
  
    db.add fp_cupom_desconto {
      data = {
        created_at       : "now"
        codigo           : $codigo
        tipo             : $tipo
        valor            : $valor
        produto          : $produto
        id_franqueado    : $input.id_franqueado
        id_representante : $id_rep
        id_central       : $input.id_central
        criado_por_tipo  : $criado_por_tipo
        alvo_nivel       : $alvo
        status           : "ativo"
        criado_por       : $input.admin_usuario
        observacao       : $input.observacao
      }
    } as $cupom
  
    function.run fn_fp_financeiro_log {
      input = {
        acao            : "cupom_criar"
        id_franqueado   : $input.id_franqueado
        ref_tipo        : "fp_cupom_desconto"
        ref_id          : $cupom.id|to_text
        detalhe         : "codigo=" ~ $codigo ~ " tipo=" ~ $tipo ~ " valor=" ~ ($valor|to_text) ~ " por=" ~ $criado_por_tipo ~ " alvo=" ~ $alvo ~ " rep=" ~ $id_rep
        origem          : "admin"
        valor           : $valor
        produto         : $produto
        admin_usuario   : $input.admin_usuario
        id_central      : $input.id_central
        id_representante: $id_rep
      }
    } as $log
  }

  response = {cupom: $cupom, registro_financeiro: $log}
}
