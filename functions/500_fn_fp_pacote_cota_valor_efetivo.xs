// Resolve pacote + valor efetivo (piso ou markup do REP)
function fn_fp_pacote_cota_valor_efetivo {
  input {
    int fp_pacote_cota_id?
    text id_franqueado? filters=trim
    text id_representante? filters=trim
    text id_central? filters=trim
  }

  stack {
    precondition ($input.fp_pacote_cota_id != null && $input.fp_pacote_cota_id > 0) {
      error = "fp_pacote_cota_id obrigatorio"
    }
  
    db.get fp_pacote_cota {
      field_name = "id"
      field_value = $input.fp_pacote_cota_id
    } as $pacote
  
    precondition ($pacote != null && $pacote.ativo == "S") {
      error = "Pacote de cota nao encontrado ou inativo"
    }
  
    var $id_rep {
      value = $input.id_representante|first_notempty:""|trim
    }
  
    conditional {
      if (($id_rep|is_empty) && (($input.id_franqueado|is_empty) == false)) {
        function.run fn_fp_franqueado_id_representante {
          input = {id_franqueado: $input.id_franqueado}
        } as $rep_res
      
        var.update $id_rep {
          value = $rep_res.id_representante|first_notempty:""|trim
        }
      }
    }
  
    var $piso {
      value = $pacote.valor|first_notempty:0
    }
  
    var $qtd {
      value = $pacote.quantidade|first_notempty:0
    }
  
    var $valor_venda {
      value = $piso
    }
  
    var $preco_rep_id {
      value = null
    }
  
    // Usar $input.fp_pacote_cota_id (nunca |get:"id" — no Xano vira null e quebra o where)
    conditional {
      if (($id_rep|is_empty) == false) {
        db.query fp_preco_pacote_cota {
          where = $db.fp_preco_pacote_cota.id_representante == $id_rep && $db.fp_preco_pacote_cota.fp_pacote_cota_id == $input.fp_pacote_cota_id && $db.fp_preco_pacote_cota.ativo == "S"
          return = {type: "single"}
        } as $preco_rep
      
        conditional {
          if ($preco_rep != null) {
            var $venda_rep {
              value = $preco_rep.valor_venda|first_notempty:0
            }
          
            conditional {
              if ($venda_rep >= $piso) {
                var.update $valor_venda {
                  value = $venda_rep
                }
              
                var.update $preco_rep_id {
                  value = $preco_rep.id
                }
              }
            }
          }
        }
      }
    }
  
    function.run fn_fp_cota_limites_de_quantidade {
      input = {quantidade: $qtd}
    } as $lim
  }

  response = {
    pacote          : $pacote
    quantidade      : $qtd
    valor_piso      : $piso
    valor_venda     : $valor_venda
    id_representante: $id_rep
    preco_rep_id    : $preco_rep_id
    limites_json    : $lim.limites_json
  }
}
