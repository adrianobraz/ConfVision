// White label do franqueado (leitura)
query fp_whitelabel_get verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    db.query fp_whitelabel {
      where = $db.fp_whitelabel.id_franqueado == $input.id_franqueado
      return = {type: "single"}
    } as $row
  }

  response = {dados: $row}
}