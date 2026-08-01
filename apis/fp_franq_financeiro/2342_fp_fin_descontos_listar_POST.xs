// Relatorio de descontos e abatimentos (cupom, credito, Pro+) — admConfmonit
query fp_fin_descontos_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text competencia? filters=trim
    timestamp? data_inicio?
    timestamp? data_fim?
    text id_franqueado? filters=trim
    text produto? filters=trim
    text tipo_desconto? filters=trim
    text codigo? filters=trim
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
  
    var $ids_fra {
      value = []
    }
  
    conditional {
      if (($escopo|get:"userTipo":"") == "REP") {
        function.run fn_fp_franqueados_ids_representante {
          input = {id_representante: $id_rep}
        } as $carteira
      
        var.update $ids_fra {
          value = $carteira|get:"ids":[]
        }
      }
    }
  
    function.run fn_fp_fin_descontos_resumo {
      input = {
        competencia     : $input.competencia
        data_inicio     : $input.data_inicio
        data_fim        : $input.data_fim
        id_franqueado   : $input.id_franqueado
        produto         : $input.produto
        tipo_desconto   : $input.tipo_desconto
        codigo          : $input.codigo
        id_representante: $id_rep
        id_central      : $id_cen
        permitir_global : $escopo|get:"permite_global":false
        ids_franqueado  : $ids_fra
      }
    } as $rel
  }

  response = {
    dados                       : $rel.dados
    total                       : $rel.total
    total_bruto                 : $rel.total_bruto
    total_descontos             : $rel.total_descontos
    total_liquido               : $rel.total_liquido
    total_cupons                : $rel.total_cupons
    total_pro_plus              : $rel.total_pro_plus
    total_credito_abatido       : $rel.total_credito_abatido
    qtd_cupons                  : $rel.qtd_cupons
    qtd_pro_plus                : $rel.qtd_pro_plus
    qtd_credito                 : $rel.qtd_credito
    qtd_franqueados_beneficiados: $rel.qtd_franqueados_beneficiados
    competencia                 : $rel.competencia
    idCentral                   : $id_cen
    gerado_em                   : now
  }
}
