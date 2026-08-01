// Cancela conta a pagar aberta exclusiva do usuario logado
query fp_conta_pagar_cancelar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int conta_id? filters=min:1
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
      error = "Somente contas abertas podem ser canceladas"
    }
  
    db.patch fp_conta_pagar {
      field_name = "id"
      field_value = $input.conta_id
      data = {
        status    : "cancelada"
        observacao: $input.observacao|first_notempty:$conta.observacao
      }
    } as $conta_upd
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "conta_pagar_cancelar"
        ref_tipo     : "fp_conta_pagar"
        ref_id       : $input.conta_id|to_text
        detalhe      : $input.observacao
        origem       : "admin"
        valor        : $conta.valor
        admin_usuario: $usuario
        id_usuario   : $admin_check.idUsuario
      }
    } as $log
  }

  response = {conta: $conta_upd}
}
