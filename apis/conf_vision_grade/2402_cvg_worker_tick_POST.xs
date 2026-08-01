// Worker tick: executa slots pendentes do minuto (America/Sao_Paulo via taskxano)
query cvg_worker_tick verb=POST {
  api_group = "confVisionGrade"

  input {
    text worker_key? filters=trim
    int dia_semana? filters=min:1|max:7
    text hora? filters=trim
    text data_ref? filters=trim
  }

  stack {
    function.run fn_cvg_worker_validar {
      input = {worker_key: $input.worker_key}
    } as $auth

    precondition ($input.dia_semana != null) {
      error = "dia_semana obrigatorio (1=seg .. 7=dom)"
    }

    precondition (($input.hora|is_empty) == false) {
      error = "hora obrigatoria (HH:MM)"
    }

    precondition (($input.data_ref|is_empty) == false) {
      error = "data_ref obrigatorio (YYYY-MM-DD)"
    }

    function.run fn_cvg_worker_tick {
      input = {
        dia_semana: $input.dia_semana
        hora      : $input.hora
        data_ref  : $input.data_ref
      }
    } as $resultado
  }

  response = $resultado
}
