// Fluxo de caixa do periodo (entradas, saidas, saldo)
query fp_fin_fluxo_caixa verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text competencia? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    var $id_rep {
      value = $escopo|get:"idRepresentante":""
    }
  
    var $id_cen {
      value = $escopo|get:"idCentral":""
    }
  
    var $usuario {
      value = $escopo|get:"usuario":""
    }
  
    var $escopo_rep {
      value = ($id_rep|is_empty) == false
    }
  
    function.run fn_fp_fin_resumo {
      input = {
        competencia     : $input.competencia
        id_representante: $id_rep
        id_central      : $id_cen
        admin_usuario   : $usuario
        permitir_global : $escopo|get:"permite_global":false
      }
    } as $resumo
  
    var $caixa_saldo {
      value = 0
    }
  
    conditional {
      if (($usuario|is_empty) == false) {
        function.run fn_fp_caixa_saldo {
          input = {
            id_central      : $id_cen
            id_representante: $id_rep
            admin_usuario   : $usuario
          }
        } as $caixa_escopo
      
        var.update $caixa_saldo {
          value = $caixa_escopo|get:"saldo":0
        }
      }
    }
  
    db.query fp_pagamento {
      where = $db.fp_pagamento.id > 0
      sort = {fp_pagamento.pago_em: "desc"}
      return = {type: "list"}
    } as $pagamentos_raw
  
    db.query fp_fatura {
      where = $db.fp_fatura.id > 0
      return = {type: "list"}
    } as $fats_all
  
    var $ids_fat_repasse {
      value = []
    }
  
    foreach ($fats_all) {
      each as $fx {
        var $tfx {
          value = $fx.tipo|first_notempty:""|trim
        }
      
        conditional {
          if ($tfx == "repasse_rep_central" || $tfx == "repasse_central_breakglass" || ($tfx|contains:"repasse")) {
            array.push $ids_fat_repasse {
              value = $fx.id
            }
          }
        }
      }
    }
  
    var $pagamentos {
      value = []
    }
  
    conditional {
      if ($escopo|get:"permite_global":false) {
        var.update $pagamentos {
          value = $pagamentos_raw
        }
      }
    
      elseif ($escopo_rep || (($id_cen|is_empty) == false)) {
        var $ids_fat {
          value = []
        }
      
        conditional {
          if ($escopo_rep) {
            function.run fn_fp_franqueados_ids_representante {
              input = {id_representante: $id_rep}
            } as $carteira
          
            var $ids_fra {
              value = $carteira.ids|first_notempty:[]
            }
          
            foreach ($fats_all) {
              each as $f {
                function.run fn_fp_fatura_na_carteira {
                  input = {
                    fatura          : $f
                    id_representante: $id_rep
                    ids_franqueado  : $ids_fra
                  }
                } as $chk
              
                conditional {
                  if ($chk.ok) {
                    array.push $ids_fat {
                      value = $f.id
                    }
                  }
                }
              }
            }
          }
        
          else {
            foreach ($fats_all) {
              each as $f {
                conditional {
                  if (($f.id_central|trim) == $id_cen) {
                    array.push $ids_fat {
                      value = $f.id
                    }
                  }
                }
              }
            }
          }
        }
      
        var $pags {
          value = []
        }
      
        foreach ($pagamentos_raw) {
          each as $p {
            var $pag_ok {
              value = false
            }
          
            foreach ($ids_fat) {
              each as $fid {
                conditional {
                  if ($pag_ok == false && $p.fp_fatura_id == $fid) {
                    var.update $pag_ok {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($pag_ok) {
                array.push $pags {
                  value = $p
                }
              }
            }
          }
        }
      
        var.update $pagamentos {
          value = $pags
        }
      }
    }
  
    var $retiradas {
      value = []
    }
  
    conditional {
      if (($usuario|is_empty) == false) {
        db.query fp_caixa_movimento {
          where = $db.fp_caixa_movimento.tipo == "retirada" && $db.fp_caixa_movimento.admin_usuario == $usuario
          sort = {fp_caixa_movimento.movimento_em: "desc"}
          return = {type: "list"}
        } as $retiradas_user
      
        var.update $retiradas {
          value = $retiradas_user
        }
      }
    }
  
    var $comp {
      value = $input.competencia|first_notempty:""
    }
  
    var $movimentos {
      value = []
    }
  
    foreach ($pagamentos) {
      each as $p {
        var $ok {
          value = true
        }
      
        conditional {
          if (($comp|strlen) >= 7) {
            var.update $ok {
              value = false
            }
          
            conditional {
              if ($p.pago_em != null) {
                conditional {
                  if (($p.pago_em|format_timestamp:"Y-m":"UTC") == $comp) {
                    var.update $ok {
                      value = true
                    }
                  }
                }
              }
            }
          }
        }
      
        conditional {
          if ($ok) {
            var $pag_rep {
              value = false
            }
          
            foreach ($ids_fat_repasse) {
              each as $rid {
                conditional {
                  if ($pag_rep == false && $p.fp_fatura_id == $rid) {
                    var.update $pag_rep {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($pag_rep) {
                array.push $movimentos {
                  value = {
                    data     : $p.pago_em
                    tipo     : "repasse"
                    natureza : "saida"
                    valor    : $p.valor|first_notempty:0
                    descricao: "Pagamento repasse fatura #" ~ ($p.fp_fatura_id|to_text)
                    ref_tipo : "fp_pagamento"
                    ref_id   : $p.id|to_text
                  }
                }
              }
            
              else {
                array.push $movimentos {
                  value = {
                    data     : $p.pago_em
                    tipo     : "entrada"
                    natureza : "entrada"
                    valor    : $p.valor|first_notempty:0
                    descricao: "Pagamento fatura #" ~ ($p.fp_fatura_id|to_text)
                    ref_tipo : "fp_pagamento"
                    ref_id   : $p.id|to_text
                  }
                }
              }
            }

          }
        }
      }
    }
  
    foreach ($retiradas) {
      each as $r {
        var $ok2 {
          value = true
        }
      
        var $dt {
          value = $r.movimento_em|first_notempty:$r.created_at
        }
      
        conditional {
          if (($comp|strlen) >= 7) {
            var.update $ok2 {
              value = false
            }
          
            conditional {
              if ($dt != null) {
                conditional {
                  if (($dt|format_timestamp:"Y-m":"UTC") == $comp) {
                    var.update $ok2 {
                      value = true
                    }
                  }
                }
              }
            }
          }
        }
      
        conditional {
          if ($ok2) {
            array.push $movimentos {
              value = ```
                {
                  data         : $dt
                  tipo         : "retirada"
                  natureza     : "saida"
                  valor        : $r.valor|first_notempty:0
                  descricao    : $r.descricao|first_notempty:"Retirada"
                  ref_tipo     : "fp_caixa_movimento"
                  ref_id       : $r.id|to_text
                  admin_usuario: $r.admin_usuario
                }
                ```
            }
          }
        }
      }
    }
  
    var $entradas_periodo {
      value = $resumo|get:"entradas_periodo":0
    }
  
    var $despesas_periodo {
      value = $resumo|get:"total_despesas_periodo":0
    }
  
    var $saldo_periodo {
      value = $entradas_periodo - $despesas_periodo
    }
  }

  response = {
    saldo_atual     : $caixa_saldo
    total_entradas  : $entradas_periodo
    total_saidas    : $despesas_periodo
    saldo_periodo   : $saldo_periodo
    competencia     : $comp
    movimentos      : $movimentos
    total_movimentos: $movimentos|count
    escopo_rep      : $escopo_rep
  }
}