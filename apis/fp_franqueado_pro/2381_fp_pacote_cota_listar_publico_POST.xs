// Lista pacotes de cotas para o franqueado (com preco do seu REP)
query fp_pacote_cota_listar_publico verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
  }

  stack {
    var $id_cen {
      value = "CENTRAL"
    }
  
    var $id_rep {
      value = ""
    }
  
    conditional {
      if (($input.id_franqueado|is_empty) == false) {
        function.run fn_fp_franqueado_id_central {
          input = {id_franqueado: $input.id_franqueado}
        } as $cen
      
        var.update $id_cen {
          value = $cen.id_central|first_notempty:"CENTRAL"
        }
      
        var.update $id_rep {
          value = $cen.id_representante|first_notempty:""
        }
      }
    }
  
    db.query fp_pacote_cota {
      where = $db.fp_pacote_cota.id_central == $id_cen && $db.fp_pacote_cota.ativo == "S"
      sort = {fp_pacote_cota.quantidade: "asc"}
      return = {type: "list"}
    } as $lista
  
    // Fallback catalogo legado CENTRAL
    conditional {
      if (($lista|count) == 0 && $id_cen != "CENTRAL") {
        db.query fp_pacote_cota {
          where = $db.fp_pacote_cota.id_central == "CENTRAL" && $db.fp_pacote_cota.ativo == "S"
          sort = {fp_pacote_cota.quantidade: "asc"}
          return = {type: "list"}
        } as $lista
      }
    }
  
    var $saida {
      value = []
    }
  
    foreach ($lista) {
      each as $p {
        function.run fn_fp_pacote_cota_valor_efetivo {
          input = {
            fp_pacote_cota_id: $p.id
            id_franqueado    : $input.id_franqueado
            id_representante : $id_rep
            id_central       : $id_cen
          }
        } as $ef
      
        var $lim_json {
          value = $ef.limites_json
        }
      
        var $qtd {
          value = $p.quantidade|first_notempty:0
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
      
        array.push $saida {
          value = $p
            |set:"valor":$ef.valor_venda
            |set:"valor_piso":$ef.valor_piso
            |set:"valor_venda":$ef.valor_venda
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
    id_central      : $id_cen
    id_representante: $id_rep
  }
}
