// Query all pro_adm_user records
query pro_adm_user verb=GET {
  api_group = "ProGerencial GoFluxo"

  input {
    int idadm?
  }

  stack {
    db.query pro_adm_user {
      where = $db.pro_adm_user.pro_adm_id == $input.idadm
      sort = {pro_adm_user.id: "asc"}
      return = {type: "list"}
    } as $model
  }

  response = {data: $model}
}