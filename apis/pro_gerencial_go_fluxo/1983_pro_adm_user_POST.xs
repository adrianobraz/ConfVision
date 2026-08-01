// Add pro_adm_user record
query pro_adm_user verb=POST {
  api_group = "ProGerencial GoFluxo"

  input {
    dblink {
      table = "pro_adm_user"
    }
  }

  stack {
    db.add pro_adm_user {
      enforce_hidden_fields = false
      data = {
        created_at: "now"
        nome      : $input.nome
        data      : $input.data
        pro_adm_id: $input.pro_adm_id
        email     : $input.email
      }
    } as $model
  }

  response = $model
}