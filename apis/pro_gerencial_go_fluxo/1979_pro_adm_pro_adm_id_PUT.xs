// Update Pro_adm record
query "pro_adm/{pro_adm_id}" verb=PUT {
  api_group = "ProGerencial GoFluxo"

  input {
    int pro_adm_id? filters=min:1
    dblink {
      table = "Pro_adm"
    }
  }

  stack {
    db.edit Pro_adm {
      field_name = "id"
      field_value = $input.pro_adm_id
      enforce_hidden_fields = false
      data = {nome: $input.nome, data: $input.data}
    } as $model
  }

  response = $model
}