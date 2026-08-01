// Break-glass: move assinatura e/ou fatura para outra Central (corrige id_central)
query fp_orfaos_mover_central verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text entidade? filters=trim
    int id?
    text id_central? filters=trim
    text admin_usuario? filters=trim
    text observacao? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    precondition (($escopo|get:"breakglass":false) == true) {
      error = "Somente Break-glass pode mover Central de registros orfaos"
    }
  
    var $ent {
      value = $input.entidade|trim|to_lower
    }
  
    var $id_cen {
      value = $input.id_central|trim
    }
  
    precondition ($ent == "assinatura" || $ent == "fatura") {
      error = "entidade deve ser assinatura ou fatura"
    }
  
    precondition ($input.id != null && $input.id > 0) {
      error = "id obrigatorio"
    }
  
    precondition (($id_cen|is_empty) == false) {
      error = "id_central destino obrigatorio"
    }
  
    var $antes {
      value = null
    }
  
    var $depois {
      value = null
    }
  
    conditional {
      if ($ent == "assinatura") {
        db.get fp_assinatura_produto {
          field_name = "id"
          field_value = $input.id
        } as $ass
      
        precondition ($ass != null) {
          error = "Assinatura nao encontrada"
        }
      
        var.update $antes {
          value = $ass.id_central|trim
        }
      
        db.patch fp_assinatura_produto {
          field_name = "id"
          field_value = $input.id
          data = {id_central: $id_cen}
        } as $depois
      
        // Move tambem faturas abertas vinculadas a esta assinatura
        var $id_ass_txt {
          value = $input.id|to_text
        }
      
        db.query fp_fatura_item {
          where = $db.fp_fatura_item.ref_tipo == "fp_assinatura_produto" && $db.fp_fatura_item.ref_id == $id_ass_txt
          return = {type: "list"}
        } as $itens
      
        foreach ($itens) {
          each as $it {
            db.get fp_fatura {
              field_name = "id"
              field_value = $it.fp_fatura_id
            } as $fat
          
            conditional {
              if ($fat != null && $fat.status == "aberta") {
                db.patch fp_fatura {
                  field_name = "id"
                  field_value = $fat.id
                  data = {id_central: $id_cen}
                } as $fat_mov
              }
            }
          }
        }
      }
    
      else {
        db.get fp_fatura {
          field_name = "id"
          field_value = $input.id
        } as $fat2
      
        precondition ($fat2 != null) {
          error = "Fatura nao encontrada"
        }
      
        var.update $antes {
          value = $fat2.id_central|trim
        }
      
        db.patch fp_fatura {
          field_name = "id"
          field_value = $input.id
          data = {id_central: $id_cen}
        } as $depois
      }
    }
  
    var $ref_tipo {
      value = "fp_fatura"
    }
  
    conditional {
      if ($ent == "assinatura") {
        var.update $ref_tipo {
          value = "fp_assinatura_produto"
        }
      }
    }
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "orfaos_mover_central"
        id_franqueado: $depois|get:"id_franqueado":""
        ref_tipo     : $ref_tipo
        ref_id       : $input.id|to_text
        detalhe      : "Moveu " ~ $ent ~ " de " ~ ($antes|first_notempty:"(vazio)") ~ " para " ~ $id_cen ~ " — " ~ ($input.observacao|first_notempty:"Break-glass")
        origem       : "admin"
        valor        : 0
        admin_usuario: $input.admin_usuario
      }
    } as $log
  }

  response = {
    ok                 : true
    entidade           : $ent
    id                 : $input.id
    id_central_anterior: $antes
    id_central_novo    : $id_cen
    registro           : $depois
  }
}
