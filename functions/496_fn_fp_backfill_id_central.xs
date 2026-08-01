// Backfill id_central em assinaturas/faturas a partir do catalogo / assinatura
// Uso: Central/Break-glass apos deploy do isolamento multi-tenant
function fn_fp_backfill_id_central {
  input {
    int limite?=500 filters=min:1|max:2000
  }

  stack {
    var $ass_ok {
      value = 0
    }
  
    var $fat_ok {
      value = 0
    }
  
    db.query fp_assinatura_produto {
      where = $db.fp_assinatura_produto.id > 0
      return = {type: "list"}
    } as $assinaturas
  
    foreach ($assinaturas) {
      each as $a {
        conditional {
          if (`($a.id_central|trim)|is_empty && ($a.fp_produto_catalogo_id|first_notempty:0) > 0`) {
            db.get fp_produto_catalogo {
              field_name = "id"
              field_value = $a.fp_produto_catalogo_id
            } as $cat
          
            conditional {
              if ($cat != null && (($cat.id_central|trim)|is_empty) == false) {
                db.patch fp_assinatura_produto {
                  field_name = "id"
                  field_value = $a.id
                  data = {id_central: $cat.id_central|trim}
                } as $patched
              
                var.update $ass_ok {
                  value = $ass_ok + 1
                }
              }
            }
          }
        }
      }
    }
  
    // Recarrega assinaturas apos patch
    db.query fp_assinatura_produto {
      where = $db.fp_assinatura_produto.id > 0
      return = {type: "list"}
    } as $assinaturas2
  
    db.query fp_fatura {
      where = $db.fp_fatura.id > 0
      return = {type: "list"}
    } as $faturas
  
    foreach ($faturas) {
      each as $f {
        conditional {
          if (($f.id_central|trim)|is_empty) {
            var $cen_fill {
              value = ""
            }
          
            // 1) Via ciclo_ref / franqueado + produto na assinatura
            foreach ($assinaturas2) {
              each as $a2 {
                conditional {
                  if (($cen_fill|is_empty) && ($a2.id_franqueado|trim) == ($f.id_franqueado|trim) && (($a2.id_central|trim)|is_empty) == false) {
                    conditional {
                      if (($f.id_representante|is_empty) || ($f.id_representante|trim) == ($a2.id_representante|trim)) {
                        var.update $cen_fill {
                          value = $a2.id_central|trim
                        }
                      }
                    }
                  }
                }
              }
            }
          
            conditional {
              if (($cen_fill|is_empty) == false) {
                db.patch fp_fatura {
                  field_name = "id"
                  field_value = $f.id
                  data = {id_central: $cen_fill}
                } as $fp
              
                var.update $fat_ok {
                  value = $fat_ok + 1
                }
              }
            }
          }
        }
      }
    }
  }

  response = {
    assinaturas_atualizadas: $ass_ok
    faturas_atualizadas    : $fat_ok
  }
}