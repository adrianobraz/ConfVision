// Limites do plano ativo (usuarios, setores, clientes, contas)
query fp_licenca_limites verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    text produto?=franqueadopro filters=trim
  }

  stack {
    function.run fn_fp_assinatura_get_ativa {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : $input.produto
      }
    } as $check
  
    var $regras {
      value = {modulos_json: {}, limites_json: {}, retencao_dias: 30}
    }
  
    conditional {
      if ($check.assinatura != null) {
        function.run fn_fp_assinatura_regras_efetivas {
          input = {
            produto               : $input.produto
            plano                 : $check.assinatura.plano
            tipo_contratacao      : $check.assinatura.tipo_contratacao
            modulos_snapshot      : $check.assinatura.modulos_json
            limites_snapshot      : $check.assinatura.limites_json
            fp_produto_catalogo_id: $check.assinatura.fp_produto_catalogo_id
          }
        } as $regras
      }
    }
  }

  response = {
    liberado     : $check.liberado
    plano        : $check.assinatura.plano
    limites_json : $regras.limites_json
    modulos_json : $regras.modulos_json
    retencao_dias: $regras.retencao_dias
  }
}