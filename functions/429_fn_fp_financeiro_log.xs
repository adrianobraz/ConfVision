// Registra evento em fp_financeiro_log
function fn_fp_financeiro_log {
  input {
    text acao? filters=trim
    text id_franqueado? filters=trim
    text ref_tipo? filters=trim
    text ref_id? filters=trim
    text detalhe? filters=trim
    text origem? filters=trim
    decimal valor?
    text produto? filters=trim
    text plano? filters=trim
    text admin_usuario? filters=trim
    text id_central? filters=trim
    text id_representante? filters=trim
    text id_usuario? filters=trim
  }

  stack {
    db.add fp_financeiro_log {
      data = {
        created_at       : "now"
        acao             : $input.acao
        id_franqueado    : $input.id_franqueado
        ref_tipo         : $input.ref_tipo
        ref_id           : $input.ref_id
        detalhe          : $input.detalhe
        origem           : $input.origem|first_notempty:"sistema"
        valor            : $input.valor
        produto          : $input.produto
        plano            : $input.plano
        admin_usuario    : $input.admin_usuario
        id_central       : $input.id_central
        id_representante : $input.id_representante
        id_usuario       : $input.id_usuario
      }
    } as $log
  }

  response = $log
}
