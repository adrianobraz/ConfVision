// Retorna saldo e totais de caixa do usuario logado (escopo CEN/REP)
query fp_caixa_saldo verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    var $id_cen {
      value = $escopo|get:"idCentral":""
    }
  
    var $id_rep {
      value = $escopo|get:"idRepresentante":""
    }
  
    var $usuario {
      value = $escopo|get:"usuario":""
    }
  
    function.run fn_fp_caixa_saldo {
      input = {
        id_central      : $id_cen
        id_representante: $id_rep
        admin_usuario   : $usuario
      }
    } as $saldo
  }

  response = {
    saldo             : $saldo|get:"saldo":0
    total_entradas    : $saldo|get:"total_entradas":0
    total_retiradas   : $saldo|get:"total_retiradas":0
    total_contas_pagas: $saldo|get:"total_contas_pagas":0
    total_repasses    : $saldo|get:"total_repasses":0
    total_saidas      : $saldo|get:"total_saidas":0
    qtd_pagamentos    : $saldo|get:"qtd_pagamentos":0
    qtd_retiradas     : $saldo|get:"qtd_retiradas":0
    idCentral         : $id_cen
    idRepresentante   : $id_rep
    usuario           : $usuario
  }
}
