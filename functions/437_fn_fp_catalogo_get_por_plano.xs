// Busca item do catalogo por produto, plano e escopo (Central)
// id_central obrigatorio — NAO defaultar "CENTRAL" (mistura Centrais)
function fn_fp_catalogo_get_por_plano {
  input {
    text id_central? filters=trim
    text id_representante? filters=trim
    text produto? filters=trim
    text plano? filters=trim
  }

  stack {
    var $id_central {
      value = $input.id_central|trim
    }
  
    var $id_representante {
      value = $input.id_representante|first_notempty:""|trim
    }
  
    var $cat {
      value = null
    }
  
    conditional {
      if (($id_central|is_empty) == false && ($input.produto|is_empty) == false && ($input.plano|is_empty) == false) {
        db.query fp_produto_catalogo {
          where = $db.fp_produto_catalogo.produto == $input.produto && $db.fp_produto_catalogo.plano == $input.plano && $db.fp_produto_catalogo.id_central == $id_central && $db.fp_produto_catalogo.id_representante == $id_representante && $db.fp_produto_catalogo.ativo == "S"
          return = {type: "single"}
        } as $cat
      }
    }
  }

  response = $cat
}
