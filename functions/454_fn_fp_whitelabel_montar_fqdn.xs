// Monta FQDN a partir de subdominio + dominio base
function fn_fp_whitelabel_montar_fqdn {
  input {
    text subdominio? filters=trim|lower
    text dominio? filters=trim|lower
  }

  stack {
    var $fqdn {
      value = $input.dominio
    }
  
    conditional {
      if (($input.dominio|is_empty) == false && ($input.subdominio|is_empty) == false && $input.subdominio != "@") {
        var.update $fqdn {
          value = $input.subdominio|concat:$input.dominio:"."
        }
      }
    }
  }

  response = $fqdn
}