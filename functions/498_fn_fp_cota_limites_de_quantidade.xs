// Deriva limites a partir da quantidade total de cota
// clientes=qtd | dispositivos=qtd*2 | usuarios_alarme=qtd*4 | setores=qtd*10
function fn_fp_cota_limites_de_quantidade {
  input {
    int quantidade?
  }

  stack {
    var $q {
      value = $input.quantidade|first_notempty:0
    }
  
    conditional {
      if ($q < 0) {
        var.update $q {
          value = 0
        }
      }
    }
  
    var $limites {
      value = {
        clientes_max       : $q
        contas_max         : $q * 2
        usuarios_alarme_max: $q * 4
        setores_alarme_max : $q * 10
        cota_quantidade    : $q
      }
    }
  }

  response = {quantidade: $q, limites_json: $limites}
}