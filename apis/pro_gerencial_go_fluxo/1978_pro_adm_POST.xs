// Add Pro_adm record
query pro_adm verb=POST {
  api_group = "ProGerencial GoFluxo"

  input {
    dblink {
      table = "Pro_adm"
    }
  }

  stack {
    db.add Pro_adm {
      enforce_hidden_fields = false
      data = {
        created_at: "now"
        nome      : $input.nome
        data      : $input.data
      }
    } as $model
  }

  response = $model
}