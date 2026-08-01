// Registra retirada de caixa exclusiva do usuario logado
query fp_caixa_retirada verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    decimal valor?
    text descricao? filters=trim
    text observacao? filters=trim
    text admin_usuario? filters=trim
    timestamp? movimento_em?
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
      error = "Informe um valor de retirada maior que zero"
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
  
    precondition ($input.valor <= $saldo.saldo) {
      error = "Valor maior que o saldo de caixa disponivel"
    }
  
    var $quando {
      value = $input.movimento_em|first_notempty:now
    }
  
    var $desc {
      value = $input.descricao
        |first_notempty:"Retirada de caixa"
    }
  
    db.add fp_caixa_movimento {
      data = {
        created_at   : "now"
        tipo         : "retirada"
        valor        : $input.valor
        descricao    : $desc
        movimento_em : $quando
        admin_usuario: $usuario
        observacao   : $input.observacao
      }
    } as $mov
  
    function.run fn_fp_financeiro_log {
      input = {
        acao           : "caixa_retirada"
        ref_tipo       : "fp_caixa_movimento"
        ref_id         : $mov.id|to_text
        detalhe        : $desc
        origem         : "admin"
        valor          : $input.valor
        admin_usuario  : $usuario
        id_central     : $id_cen
        id_representante: $id_rep
        id_usuario     : $admin_check.idUsuario
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

  response = {movimento: $mov, saldo: $saldo_apos}
}
