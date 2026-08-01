// Cria conta a pagar exclusiva do usuario logado
query fp_conta_pagar_criar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text fornecedor? filters=trim
    text descricao? filters=trim
    text categoria? filters=trim
    decimal valor?
    timestamp? vencimento_em?
    text observacao? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin_check
  
    var $usuario {
      value = $admin_check.usuario|first_notempty:""|trim
    }
  
    precondition (($usuario|is_empty) == false) {
      error = "Sessao sem usuario — faca logout e login novamente"
    }
  
    precondition (($input.valor|first_notempty:0) > 0) {
      error = "Informe o valor da conta a pagar"
    }
  
    precondition (($input.descricao|strlen) > 0 || ($input.fornecedor|strlen) > 0) {
      error = "Informe fornecedor ou descricao"
    }
  
    var $id_cen {
      value = ""
    }
  
    var $id_rep {
      value = ""
    }
  
    conditional {
      if ($admin_check.userTipo == "REP") {
        var.update $id_rep {
          value = $admin_check.idVinculo|first_notempty:""|trim
        }
      }
    
      else {
        var.update $id_cen {
          value = $admin_check.idCentral
            |first_notempty:($admin_check.idVinculo|first_notempty:"")
            |trim
        }
      }
    }
  
    db.add fp_conta_pagar {
      data = {
        created_at   : "now"
        fornecedor   : $input.fornecedor
        descricao    : $input.descricao
        categoria    : $input.categoria|first_notempty:"geral"
        valor        : $input.valor
        vencimento_em: $input.vencimento_em
        status       : "aberta"
        admin_usuario: $usuario
        observacao   : $input.observacao
      }
    } as $conta
  
    function.run fn_fp_financeiro_log {
      input = {
        acao            : "conta_pagar_criar"
        ref_tipo        : "fp_conta_pagar"
        ref_id          : $conta.id|to_text
        detalhe         : ($input.fornecedor|first_notempty:"") ~ " " ~ ($input.descricao|first_notempty:"")
        origem          : "admin"
        valor           : $input.valor
        admin_usuario   : $usuario
        id_central      : $id_cen
        id_representante: $id_rep
        id_usuario      : $admin_check.idUsuario
      }
    } as $log
  }

  response = {conta: $conta}
}
