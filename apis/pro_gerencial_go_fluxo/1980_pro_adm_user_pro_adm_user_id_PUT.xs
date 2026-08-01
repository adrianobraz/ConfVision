// Update pro_adm_user record
query "pro_adm_user/{pro_adm_user_id}" verb=PUT {
  api_group = "ProGerencial GoFluxo"

  input {
    int pro_adm_user_id? filters=min:1
    dblink {
      table = "pro_adm_user"
    }
  }

  stack {
    db.edit pro_adm_user {
      field_name = "id"
      field_value = $input.pro_adm_user_id
      enforce_hidden_fields = false
      data = {
        nome      : $input.nome
        data      : $input.data
        pro_adm_id: $input.pro_adm_id
        email     : $input.email
      }
    } as $model
  }

  response = $model
}