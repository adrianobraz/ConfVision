query "autofim/processo-end" verb=POST {
  api_group = "task"

  input {
    text idProcesso
    text motivo
    text perfilLocal?
  }

  stack {
    precondition (($input.idProcesso|trim|is_empty) == false) {
      error = "idProcesso obrigatório"
      payload = false
    }
  
    precondition (($input.motivo|trim|is_empty) == false) {
      error = "motivo obrigatório"
      payload = false
    }
  
    var $perfil {
      value = "NORMAL"
    }
  
    conditional {
      if (($input.perfilLocal|trim|is_empty) == false) {
        var.update $perfil {
          value = $input.perfilLocal
        }
      }
    }
  
    function.run ProcessoEnd {
      input = {
        idUsuario  : "0ROBOAUTO"
        Descricao  : "[AUTO] " ~ $input.motivo ~ " [" ~ $perfil ~ "]"
        UsuarioNome: "ROBO AUTO"
        idProcesso : $input.idProcesso
        telefone   : "ROBO AUTO"
        token      : "ROBO AUTO"
        ip         : ""
        ipcidade   : ""
        ipestado   : ""
        ipcep      : ""
        geolat     : 0
        geolon     : 0
        georua     : ""
        geonumero  : ""
        geobairro  : ""
        geocidade  : ""
        geoestado  : ""
        device     : ""
        browser    : ""
        sitema     : ""
        timezone   : ""
        fingerprint: "ROBO AUTO"
      }
    } as $procEnd
  
    var $acao {
      value = "FINALIZOU"
    }
  
    conditional {
      if ($procEnd.dados|is_empty) {
        var.update $acao {
          value = "JA_FINALIZADO"
        }
      }
    }
  }

  response = {
    dados: ""|set:"acao":$acao|set:"retornoProcessoEnd":$procEnd
  }
}