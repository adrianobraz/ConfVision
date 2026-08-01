// Delete pro_adm_user record
query "pro_adm_user/{pro_adm_user_id}" verb=DELETE {
  api_group = "ProGerencial GoFluxo"

  input {
    int pro_adm_user_id? filters=min:1
  }

  stack {
    db.del pro_adm_user {
      field_name = "id"
      field_value = $input.pro_adm_user_id
    }
  }

  response = null
}