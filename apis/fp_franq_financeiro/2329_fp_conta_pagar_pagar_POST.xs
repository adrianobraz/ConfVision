// Baixa (paga) conta a pagar exclusiva do usuario e registra saida de caixa
query fp_conta_pagar_pagar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int conta_id? filters=min:1
    text observacao? filters=trim
    text admin_usuario? filters=trim
    timestamp? pago_em?
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
  
    db.get fp_conta_pagar {
      field_name = "id"
      field_value = $input.conta_id
    } as $conta
  
    precondition ($conta != null) {
      error = "Conta a pagar nao encontrada"
    }
  
    precondition (($conta.admin_usuario|trim) == $usuario) {
      error = "Conta a pagar de outro usuario"
    }
  
    precondition ($conta.status == "aberta") {
      error = "Somente contas abertas podem ser pagas"
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
  
    function.run fn_fp_caixa_saldo {
      input = {
        id_central      : $id_cen
        id_representante: $id_rep
        admin_usuario   : $usuario
      }
    } as $saldo
  
    precondition (($conta.valor|first_notempty:0) <= $saldo.saldo) {
      error = "Saldo de caixa insuficiente para pagar esta conta"
    }
  
    var $quando {
      value = $input.pago_em|first_notempty:now
    }
  
    db.patch fp_conta_pagar {
      field_name = "id"
      field_value = $input.conta_id
      data = {
        status    : "paga"
        pago_em   : $quando
        observacao: $input.observacao|first_notempty:$conta.observacao
      }
    } as $conta_upd
  
    var $nome_ref {
      value = $conta.fornecedor|first_notempty:$conta.descricao
    }
  
    var $desc {
      value = "Pagamento: " ~ ($nome_ref|first_notempty:"conta")
    }
  
    db.add fp_caixa_movimento {
      data = {
        created_at   : "now"
        tipo         : "retirada"
        valor        : $conta.valor
        descricao    : $desc
        movimento_em : $quando
        admin_usuario: $usuario
        ref_tipo     : "fp_conta_pagar"
        ref_id       : $input.conta_id|to_text
        observacao   : $input.observacao
      }
    } as $mov
  
    function.run fn_fp_financeiro_log {
      input = {
        acao            : "conta_pagar_pagar"
        ref_tipo        : "fp_conta_pagar"
        ref_id          : $input.conta_id|to_text
        detalhe         : $desc
        origem          : "admin"
        valor           : $conta.valor
        admin_usuario   : $usuario
        id_central      : $id_cen
        id_representante: $id_rep
        id_usuario      : $admin_check.idUsuario
      }
    } as $log
  
    function.run fn_fp_caixa_saldo {
      input = {
        id_central      : $id_cen
        id_representante: $id_rep
        admin_usuario   : $usuario
      }
    } as $saldo_apos
  }

  response = {conta: $conta_upd, movimento: $mov, saldo: $saldo_apos}
}
