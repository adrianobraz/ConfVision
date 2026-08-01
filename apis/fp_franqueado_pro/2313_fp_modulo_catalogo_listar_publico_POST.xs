// Lista catalogo de modulos a la carte (publico — franqueado)
query fp_modulo_catalogo_listar_publico verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text produto?=franqueadopro filters=trim
  }

  stack {
    db.query fp_modulo_catalogo {
      where = $db.fp_modulo_catalogo.id_central == "CENTRAL" && $db.fp_modulo_catalogo.id_representante == "" && $db.fp_modulo_catalogo.produto == $input.produto && $db.fp_modulo_catalogo.ativo == "S"
      sort = {
        fp_modulo_catalogo.ordem: "asc"
        fp_modulo_catalogo.grupo: "asc"
      }
    
      return = {type: "list"}
    } as $lista
  }

  response = {dados: $lista, total: $lista|count}
}