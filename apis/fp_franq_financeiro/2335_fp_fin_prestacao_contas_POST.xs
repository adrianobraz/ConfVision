// Prestacao de contas consolidada do periodo (escopo CEN/REP)
query fp_fin_prestacao_contas verb=POST {
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
  
    function.run fn_fp_fin_resumo {
      input = {
        competencia     : $input.competencia
        id_representante: $id_rep
        id_central      : $id_cen
        admin_usuario   : $escopo|get:"usuario":""
        permitir_global : $escopo|get:"permite_global":false
      }
    } as $resumo
  
    db.query fp_fatura {
      where = $db.fp_fatura.status == "aberta" && $db.fp_fatura.vencimento_em != null && $db.fp_fatura.vencimento_em < now
      return = {type: "list"}
    } as $vencidas_raw
  
    db.query fp_fatura {
      where = $db.fp_fatura.status == "aberta"
      return = {type: "list"}
    } as $a_receber_raw
  
    var $vencidas {
      value = []
    }
  
    var $a_receber {
      value = []
    }
  
    conditional {
      if (($id_rep|is_empty) == false) {
        function.run fn_fp_franqueados_ids_representante {
          input = {id_representante: $id_rep}
        } as $carteira
      
        var $ids {
          value = $carteira.ids|first_notempty:[]
        }
      
        foreach ($vencidas_raw) {
          each as $f {
            function.run fn_fp_fatura_na_carteira {
              input = {
                fatura          : $f
                id_representante: $id_rep
                ids_franqueado  : $ids
              }
            } as $chk
          
            conditional {
              if ($chk.ok) {
                array.push $vencidas {
                  value = $f
                }
              }
            }
          }
        }
      
        foreach ($a_receber_raw) {
          each as $f {
            function.run fn_fp_fatura_na_carteira {
              input = {
                fatura          : $f
                id_representante: $id_rep
                ids_franqueado  : $ids
              }
            } as $chk
          
            conditional {
              if ($chk.ok) {
                array.push $a_receber {
                  value = $f
                }
              }
            }
          }
        }
      }
    
      elseif ($escopo|get:"permite_global":false) {
        var.update $vencidas {
          value = $vencidas_raw
        }
      
        var.update $a_receber {
          value = $a_receber_raw
        }
      }
    
      else {
        foreach ($vencidas_raw) {
          each as $f {
            conditional {
              if (($f.id_central|trim) == $id_cen) {
                array.push $vencidas {
                  value = $f
                }
              }
            }
          }
        }
      
        foreach ($a_receber_raw) {
          each as $f {
            var $tf {
              value = $f.tipo|first_notempty:""|trim
            }
          
            conditional {
              if (($f.id_central|trim) == $id_cen && $tf != "repasse_rep_central" && $tf != "repasse_central_breakglass" && (($tf|contains:"repasse") == false)) {
                array.push $a_receber {
                  value = $f
                }
              }
            }
          }
        }
      }
    }
  }

  response = {
    competencia      : $input.competencia
    resumo           : $resumo
    saldo_caixa      : $resumo.saldo_caixa
    receitas         : $resumo.receita_periodo
    despesas         : $resumo.total_despesas_periodo
    resultado        : $resumo.resultado_periodo
    a_receber        : $resumo.valor_em_aberto
    vencido          : $resumo.valor_vencido
    a_pagar          : $resumo.valor_a_pagar
    qtd_inadimplentes: $resumo.qtd_inadimplentes
    faturas_a_receber: $a_receber
    faturas_vencidas : $vencidas
    gerado_em        : now
    idCentral        : $id_cen
  }
}
