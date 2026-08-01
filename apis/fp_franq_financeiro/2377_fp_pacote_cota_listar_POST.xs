// Lista pacotes de cotas (CEN/BG) ou com preco de venda (REP)
query fp_pacote_cota_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_central? filters=trim
    text ativo?=S filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    var $id_central {
      value = $admin.idCentralCatalogo
        |first_notempty:($admin.idVinculo|trim)
    }
  
    conditional {
      if ($admin.userTipo == "REP") {
        var.update $id_central {
          value = $admin.idCentralCatalogo|first_notempty:""|trim
        }
      }
    
      elseif ($admin.breakglass) {
        conditional {
          if (($input.id_central|is_empty) == false) {
            var.update $id_central {
              value = $input.id_central|trim
            }
          }
        
          elseif ($id_central|is_empty) {
            var.update $id_central {
              value = "CENTRAL"
            }
          }
        }
      }
    
      else {
        precondition (($id_central|is_empty) == false) {
          error = "Sessao sem Central vinculada"
        }
      }
    }
  
    conditional {
      if ($id_central|is_empty) {
        var.update $id_central {
          value = "CENTRAL"
        }
      }
    }
  
    db.query fp_pacote_cota {
      where = $db.fp_pacote_cota.id_central == $id_central && $db.fp_pacote_cota.ativo ==? $input.ativo
      sort = {fp_pacote_cota.quantidade: "asc"}
      return = {type: "list"}
    } as $lista
  
    var $saida {
      value = []
    }
  
    var $id_rep {
      value = ""
    }
  
    conditional {
      if ($admin.userTipo == "REP") {
        var.update $id_rep {
          value = $admin.idVinculo|trim
        }
      }
    }
  
    foreach ($lista) {
      each as $p {
        var $pid {
          value = $p.id
        }
      
        var $qtd {
          value = $p.quantidade|first_notempty:0
        }
      
        var $piso {
          value = $p.valor|first_notempty:0
        }
      
        function.run fn_fp_cota_limites_de_quantidade {
          input = {quantidade: $qtd}
        } as $lim
      
        var $lim_json {
          value = $lim.limites_json
        }
      
        var $n_cli {
          value = $lim_json.clientes_max|first_notempty:$qtd
        }
      
        var $n_disp {
          value = $lim_json.contas_max|first_notempty:0
        }
      
        conditional {
          if ($n_disp == 0) {
            var.update $n_disp {
              value = $qtd * 2
            }
          }
        }
      
        var $n_usu {
          value = $lim_json.usuarios_alarme_max|first_notempty:0
        }
      
        conditional {
          if ($n_usu == 0) {
            var.update $n_usu {
              value = $qtd * 4
            }
          }
        }
      
        var $n_set {
          value = $lim_json.setores_alarme_max|first_notempty:0
        }
      
        conditional {
          if ($n_set == 0) {
            var.update $n_set {
              value = $qtd * 10
            }
          }
        }
      
        var $venda {
          value = $piso
        }
      
        var $preco_id {
          value = null
        }
      
        conditional {
          if (($id_rep|is_empty) == false) {
            db.query fp_preco_pacote_cota {
              where = $db.fp_preco_pacote_cota.id_representante == $id_rep && $db.fp_preco_pacote_cota.fp_pacote_cota_id == $pid && $db.fp_preco_pacote_cota.ativo == "S"
              return = {type: "single"}
            } as $pr
          
            conditional {
              if ($pr != null) {
                var $venda_rep {
                  value = $pr.valor_venda|first_notempty:0
                }
              
                conditional {
                  if ($venda_rep >= $piso) {
                    var.update $venda {
                      value = $venda_rep
                    }
                  
                    var.update $preco_id {
                      value = $pr.id
                    }
                  }
                }
              }
            }
          }
        }
      
        // Partir do registro do banco (garante id/nome) e so acrescentar campos calculados
        array.push $saida {
          value = $p
            |set:"valor_piso":$piso
            |set:"valor_venda":$venda
            |set:"preco_rep_id":$preco_id
            |set:"limites_json":$lim_json
            |set:"clientes":$n_cli
            |set:"dispositivos":$n_disp
            |set:"usuarios_alarme":$n_usu
            |set:"setores_alarme":$n_set
        }
      }
    }
  }

  response = {
    dados           : $saida
    total           : $saida|count
    id_central      : $id_central
    id_representante: $id_rep
    userTipo        : $admin.userTipo
  }
}
