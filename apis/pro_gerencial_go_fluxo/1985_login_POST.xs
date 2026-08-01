query Login verb=POST {
  api_group = "ProGerencial GoFluxo"

  input {
    email email? filters=trim|lower
    text password? filters=trim
  }

  stack {
    db.get pro_adm_user {
      field_name = "email"
      field_value = $input.email
    } as $pro_adm_user1
  
    precondition ($pro_adm_user1 != null) {
      error = "Usuario não encontrado"
    }
  
    security.check_password {
      text_password = $input.password
      hash_password = $pro_adm_user1.senha
    } as $x1
  
    precondition ($x1) {
      error_type = "unauthorized"
      error = "Acesso Negado"
    }
  
    db.get Pro_adm {
      field_name = "id"
      field_value = $pro_adm_user1.pro_adm_id
    } as $Pro_adm1
  }

  response = {}
    |set:"Usuario":$pro_adm_user1
    |set:"Empresa":$Pro_adm1
}