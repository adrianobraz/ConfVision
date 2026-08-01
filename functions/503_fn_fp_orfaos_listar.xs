// Detecta inconsistencias: central divergente do REP, assinatura sem fatura, fatura sem assinatura
// Uso exclusivo Break-glass (diagnostico / limpeza)
function fn_fp_orfaos_listar {
  input {
    text id_franqueado? filters=trim
    int limite?=200 filters=min:1|max:500
  }

  stack {
    var $filtro_fra {
      value = $input.id_franqueado|trim
    }
  
    var $limite {
      value = $input.limite|first_notempty:200
    }
  
    var $saida {
      value = []
    }
  
    var $assinaturas {
      value = []
    }
  
    var $faturas_raw {
      value = []
    }
  
    conditional {
      if (($filtro_fra|is_empty) == false) {
        db.query fp_assinatura_produto {
          where = $db.fp_assinatura_produto.id_franqueado == $filtro_fra
          sort = {fp_assinatura_produto.id: "desc"}
          return = {type: "list"}
        } as $assinaturas
      
        db.query fp_fatura {
          where = $db.fp_fatura.id_franqueado == $filtro_fra
          sort = {fp_fatura.id: "desc"}
          return = {type: "list"}
        } as $faturas_raw
      }
    
      else {
        db.query fp_assinatura_produto {
          where = $db.fp_assinatura_produto.id > 0
          sort = {fp_assinatura_produto.id: "desc"}
          return = {type: "list"}
        } as $assinaturas
      
        db.query fp_fatura {
          where = $db.fp_fatura.id > 0
          sort = {fp_fatura.id: "desc"}
          return = {type: "list"}
        } as $faturas_raw
      }
    }
  
    var $faturas {
      value = []
    }
  
    foreach ($faturas_raw) {
      each as $f {
        var $tipo_f {
          value = $f.tipo|first_notempty:"assinatura"|trim
        }
      
        conditional {
          if (($tipo_f|contains:"repasse") == false) {
            array.push $faturas {
              value = $f
            }
          }
        }
      }
    }
  
    db.query fp_fatura_item {
      where = $db.fp_fatura_item.ref_tipo == "fp_assinatura_produto"
      return = {type: "list"}
    } as $itens
  
    var $ids_ass_com_item {
      value = []
    }
  
    foreach ($itens) {
      each as $it {
        var $rid {
          value = $it.ref_id|to_int|first_notempty:0
        }
      
        conditional {
          if ($rid > 0) {
            array.push $ids_ass_com_item {
              value = $rid
            }
          }
        }
      }
    }
  
    // Cache de central esperada por franqueado (via REP na API legada)
    var $cache_fra {
      value = []
    }
  
    var $fra_unicos {
      value = []
    }
  
    foreach ($assinaturas) {
      each as $a0 {
        var $idf0 {
          value = $a0.id_franqueado|trim
        }
      
        conditional {
          if (($idf0|is_empty) == false) {
            var $ja0 {
              value = false
            }
          
            foreach ($fra_unicos) {
              each as $fu {
                conditional {
                  if ($fu == $idf0) {
                    var.update $ja0 {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($ja0 == false) {
                array.push $fra_unicos {
                  value = $idf0
                }
              }
            }
          }
        }
      }
    }
  
    foreach ($faturas) {
      each as $f0 {
        var $idf1 {
          value = $f0.id_franqueado|trim
        }
      
        conditional {
          if (($idf1|is_empty) == false) {
            var $ja1 {
              value = false
            }
          
            foreach ($fra_unicos) {
              each as $fu2 {
                conditional {
                  if ($fu2 == $idf1) {
                    var.update $ja1 {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($ja1 == false) {
                array.push $fra_unicos {
                  value = $idf1
                }
              }
            }
          }
        }
      }
    }
  
    var $n_res {
      value = 0
    }
  
    foreach ($fra_unicos) {
      each as $idf {
        conditional {
          if ($n_res < 80) {
            try_catch {
              try {
                function.run fn_fp_franqueado_id_central {
                  input = {id_franqueado: $idf}
                } as $cen_res
              
                array.push $cache_fra {
                  value = {
                    id_franqueado   : $idf
                    id_central      : $cen_res|get:"id_central":""|trim
                    id_representante: $cen_res|get:"id_representante":""|trim
                  }
                }
              }
            
              catch {
                array.push $cache_fra {
                  value = {
                    id_franqueado   : $idf
                    id_central      : ""
                    id_representante: ""
                  }
                }
              }
            }
          
            var.update $n_res {
              value = $n_res + 1
            }
          }
        }
      }
    }
  
    // --- Assinaturas: mismatch de central + sem fatura ---
    foreach ($assinaturas) {
      each as $a {
        conditional {
          if (($saida|count) < $limite) {
            var $idf {
              value = $a.id_franqueado|trim
            }
          
            var $cen_atual {
              value = $a.id_central|trim
            }
          
            var $cen_esp {
              value = ""
            }
          
            var $rep_esp {
              value = ""
            }
          
            foreach ($cache_fra) {
              each as $c {
                conditional {
                  if (($c|get:"id_franqueado":"") == $idf) {
                    var.update $cen_esp {
                      value = $c|get:"id_central":""|trim
                    }
                  
                    var.update $rep_esp {
                      value = $c|get:"id_representante":""|trim
                    }
                  }
                }
              }
            }
          
            // Central do registro != Central do REP do franqueado
            conditional {
              if (($rep_esp|is_empty) == false && ($cen_esp|is_empty) == false && $cen_atual != $cen_esp) {
                array.push $saida {
                  value = {
                    tipo_problema      : "central_mismatch"
                    entidade           : "assinatura"
                    id                 : $a.id
                    id_franqueado      : $idf
                    id_central_atual   : $cen_atual
                    id_central_esperada: $cen_esp
                    id_representante   : $a.id_representante|first_notempty:$rep_esp
                    produto            : $a.produto
                    plano              : $a.plano
                    status             : $a.status
                    valor              : $a.valor|first_notempty:0
                    detalhe            : "Assinatura com id_central diferente da Central do REP do franqueado"
                  }
                }
              }
            }
          
            // Assinatura ativa/pendente sem fatura aberta vinculada
            conditional {
              if ($a.status == "ativa" || $a.status == "pendente") {
                var $tem_fat {
                  value = false
                }
              
                var $aid {
                  value = $a.id
                }
              
                foreach ($ids_ass_com_item) {
                  each as $ia {
                    conditional {
                      if ($ia == $aid) {
                        // Confirma se alguma fatura aberta aponta para esta assinatura
                        foreach ($itens) {
                          each as $it2 {
                            conditional {
                              if (($it2.ref_id|to_int) == $aid) {
                                foreach ($faturas) {
                                  each as $ff {
                                    conditional {
                                      if ($ff.id == $it2.fp_fatura_id && $ff.status == "aberta") {
                                        var.update $tem_fat {
                                          value = true
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
                    }
                  }
                }
              
                // Sem ciclo e sem fatura aberta do franqueado
                conditional {
                  if ($tem_fat == false) {
                    var $ciclo {
                      value = $a.ciclo_fatura_ref|trim
                    }
                  
                    conditional {
                      if (($ciclo|is_empty) == false) {
                        foreach ($faturas) {
                          each as $fx {
                            conditional {
                              if (($fx.id_franqueado|trim) == $idf && ($fx.ciclo_ref|trim) == $ciclo && $fx.status == "aberta") {
                                var.update $tem_fat {
                                  value = true
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  
                    conditional {
                      if ($tem_fat == false) {
                        array.push $saida {
                          value = {
                            tipo_problema      : "assinatura_sem_fatura"
                            entidade           : "assinatura"
                            id                 : $a.id
                            id_franqueado      : $idf
                            id_central_atual   : $cen_atual
                            id_central_esperada: $cen_esp
                            id_representante   : $a.id_representante|first_notempty:$rep_esp
                            produto            : $a.produto
                            plano              : $a.plano
                            status             : $a.status
                            valor              : $a.valor|first_notempty:0
                            detalhe            : "Assinatura sem fatura aberta vinculada"
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
      }
    }
  
    // --- Faturas: mismatch + sem assinatura ---
    foreach ($faturas) {
      each as $f {
        conditional {
          if (($saida|count) < $limite) {
            var $idf_f {
              value = $f.id_franqueado|trim
            }
          
            var $cen_f {
              value = $f.id_central|trim
            }
          
            var $cen_esp_f {
              value = ""
            }
          
            var $rep_esp_f {
              value = ""
            }
          
            foreach ($cache_fra) {
              each as $cf {
                conditional {
                  if (($cf|get:"id_franqueado":"") == $idf_f) {
                    var.update $cen_esp_f {
                      value = $cf|get:"id_central":""|trim
                    }
                  
                    var.update $rep_esp_f {
                      value = $cf|get:"id_representante":""|trim
                    }
                  }
                }
              }
            }
          
            conditional {
              if (($rep_esp_f|is_empty) == false && ($cen_esp_f|is_empty) == false && $cen_f != $cen_esp_f) {
                array.push $saida {
                  value = {
                    tipo_problema      : "central_mismatch"
                    entidade           : "fatura"
                    id                 : $f.id
                    id_franqueado      : $idf_f
                    id_central_atual   : $cen_f
                    id_central_esperada: $cen_esp_f
                    id_representante   : $f.id_representante|first_notempty:$rep_esp_f
                    produto            : ""
                    plano              : ""
                    status             : $f.status
                    valor              : $f.valor_total|first_notempty:0
                    detalhe            : "Fatura com id_central diferente da Central do REP do franqueado"
                  }
                }
              }
            }
          
            // Fatura aberta sem assinatura existente
            conditional {
              if ($f.status == "aberta") {
                var $tem_ass {
                  value = false
                }
              
                foreach ($itens) {
                  each as $it3 {
                    conditional {
                      if ($it3.fp_fatura_id == $f.id) {
                        var $ref_a {
                          value = $it3.ref_id|to_int|first_notempty:0
                        }
                      
                        foreach ($assinaturas) {
                          each as $ax {
                            conditional {
                              if ($ax.id == $ref_a) {
                                var.update $tem_ass {
                                  value = true
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }
              
                // Fallback: fatura do franqueado com ciclo batendo assinatura
                conditional {
                  if ($tem_ass == false) {
                    var $ciclo_f {
                      value = $f.ciclo_ref|trim
                    }
                  
                    conditional {
                      if (($ciclo_f|is_empty) == false) {
                        foreach ($assinaturas) {
                          each as $ay {
                            conditional {
                              if (($ay.id_franqueado|trim) == $idf_f && ($ay.ciclo_fatura_ref|trim) == $ciclo_f) {
                                var.update $tem_ass {
                                  value = true
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }
              
                conditional {
                  if ($tem_ass == false) {
                    array.push $saida {
                      value = {
                        tipo_problema      : "fatura_sem_assinatura"
                        entidade           : "fatura"
                        id                 : $f.id
                        id_franqueado      : $idf_f
                        id_central_atual   : $cen_f
                        id_central_esperada: $cen_esp_f
                        id_representante   : $f.id_representante|first_notempty:$rep_esp_f
                        produto            : ""
                        plano              : ""
                        status             : $f.status
                        valor              : $f.valor_total|first_notempty:0
                        detalhe            : "Fatura aberta sem assinatura vinculada"
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
  }

  response = {
    dados: $saida
    total: $saida|count
  }
}
