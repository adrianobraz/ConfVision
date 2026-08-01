// Executa um slot manualmente (debug / retry)
query cvg_slot_executar verb=POST {
  api_group = "confVisionGrade"

  input {
    text worker_key? filters=trim
    int slot_id? filters=min:1
    text data_ref? filters=trim
    text hora_ref? filters=trim
  }

  stack {
    function.run fn_cvg_worker_validar {
      input = {worker_key: $input.worker_key}
    } as $auth

    function.run fn_cvg_executar_slot {
      input = {
        slot_id : $input.slot_id
        data_ref: $input.data_ref
        hora_ref: $input.hora_ref
      }
    } as $resultado
  }

  response = $resultado
}
