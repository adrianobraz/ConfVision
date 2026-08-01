// Forca sincronizacao UsuarioTeminal do master no MySQL legado (admin)
query fp_terminal_sync_manual verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_franqueado? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin_check
  
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    function.run fn_franqueado_sync_usuario_terminal {
      input = {id_franqueado: $input.id_franqueado}
    } as $sync
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "terminal_sync_manual"
        id_franqueado: $input.id_franqueado
        ref_tipo     : "franqueado"
        ref_id       : $input.id_franqueado
        detalhe      : $sync.acesso.motivo ~ " -> " ~ $sync.acesso.usuario_terminal
        origem       : "admin"
        admin_usuario: $input.admin_usuario
        produto      : $sync.acesso.produto
      }
    } as $log
  }

  response = {sync: $sync, log: $log}
}