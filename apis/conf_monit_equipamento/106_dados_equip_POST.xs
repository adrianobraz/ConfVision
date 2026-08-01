query DadosEquip verb=POST {
  api_group = "ConfMonitEquipamento"

  input {
    text Authorization? filters=trim
    text idDispositivo? filters=trim
  }

  stack {
    function.run ConfMonitEquipamento_DadosEquip {
      input = {
        Authorization: $input.Authorization
        idDispositivo: $input.idDispositivo
      }
    } as $func_1
  }

  response = $func_1
}